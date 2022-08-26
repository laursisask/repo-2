package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/humio/cli/api"
	"github.com/humio/cli/shipper"
	"log"
	"net/url"
	"os"
)

var awsSession = session.New()
var client = lambda.New(awsSession)

type incommingEvent struct {
	Records []record `json:"Records"`
}

type record struct {
	events.SNSEventRecord
	events.SQSMessage
}

func callLambda() (string, error) {
	input := &lambda.GetAccountSettingsInput{}
	req, resp := client.GetAccountSettingsRequest(input)
	err := req.Send()
	output, _ := json.Marshal(resp.AccountUsage)
	return string(output), err
}

type cloudTrailFile struct {
	Records []json.RawMessage `json:"Records"`
}

type internalSNSMessage struct {
	S3Bucket    string   `json:"s3Bucket"`    // S3 bucket name
	S3ObjectKey []string `json:"s3ObjectKey"` // Path inside S3 bucket
}

func pushS3ContentToHumio(ctx context.Context, humio *shipper.LogShipper, bucketName, path string) error {
	s3client := s3.New(awsSession)
	r, err := s3client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot acquire object %s from bucket %s: %s", path, bucketName, err))
	}
	defer r.Body.Close()
	buf := new(bytes.Buffer)
	bytesRead, err := buf.ReadFrom(r.Body)
	if err != nil {
		return fmt.Errorf("failed to get S3 result: %w", err)
	}
	log.Printf("Successfully read %d bytes from S3 %s", bytesRead, path)

	var content cloudTrailFile
	err = json.Unmarshal(buf.Bytes(), &content)
	if err != nil {
		return fmt.Errorf("failed to parse cloud trail file located at %s in %s. Error: %w", path, bucketName, err)
	}
	log.Printf("parsed %d records", len(content.Records))
	log.Printf("first record: %v", string(content.Records[0]))
	for i, _ := range content.Records {
		humio.HandleLine(string(content.Records[i]))
	}

	return nil
}

func newHumioShipper() (*shipper.LogShipper, error) {
	config := api.DefaultConfig()
	address := os.Getenv("HUMIO_ADDRESS")
	parsedUrl, err := url.Parse(address)
	if err != nil {
		return nil, fmt.Errorf("Cannot parse HUMIO_ADDRESS %s: %w", address, err)
	}
	config.Address = parsedUrl
	config.Token = os.Getenv("HUMIO_TOKEN")
	config.CACertificatePEM = os.Getenv("HUMIO_CA_CERTIFICATE")
	config.UserAgent = "AWS Lambda Ingest Function"
	client := api.NewClient(config)
	if client == nil {
		return nil, fmt.Errorf("Could not instantiate new humio client")
	}

	humioShipper := shipper.LogShipper{
		APIClient:           client,
		URL:                 "api/v1/repositories/simon/ingest-messages", // TODO: repo url
		ParserName:          "json",
		MaxAttemptsPerBatch: 3,
		BatchSizeLines:      1,
		BatchSizeBytes:      1024 * 80,
		Logger:              log.Printf,
		// TODO: ingest tags
	}

	humioShipper.ErrorBehaviour = shipper.ErrorBehaviourPanic
	humioShipper.Start()
	// TODO: ensure graceful shutdown
	return &humioShipper, nil
}

func handleSNSNotification(ctx context.Context, notification events.SNSEntity) error {
	if notification.Type != "Notification" {
		return errors.New(fmt.Sprintf("Unexpected SNS entity type: %s", notification.Type))
	}
	var msg internalSNSMessage
	err := json.Unmarshal([]byte(notification.Message), &msg)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot parse Message inside SNSEntity: %s", err))
	}
	humioShipper, err := newHumioShipper()
	if err != nil {
		return fmt.Errorf("Cannot connect to Humio: %w", err)
	}

	for _, s3path := range msg.S3ObjectKey {
		err = pushS3ContentToHumio(ctx, humioShipper, msg.S3Bucket, s3path)
		if err != nil {
			return err
		}
	}

	return nil
}

func handleRequest(ctx context.Context, event incommingEvent) (string, error) {
	if event.Records == nil {
		return "", errors.New("Unexpected event format, resources not present")
	}

	for _, record := range event.Records {
		// TODO "Type": "Notification",
		switch {
		case record.SNSEventRecord.EventSource == "aws:sns":
			err := handleSNSNotification(ctx, record.SNSEventRecord.SNS)
			if err != nil {
				return "", err
			}
		case record.SQSMessage.EventSource == "aws:sqs":
			recordJson, _ := json.MarshalIndent(record.SQSMessage, "", "  ")
			log.Printf("handling an SQS: %s", recordJson)
		default:
			eventJson, _ := json.MarshalIndent(record, "", "  ")
			log.Printf("Cannot process following event: %s", eventJson)
			return "", errors.New(fmt.Sprintf("Unexpected event source/type: %s", eventJson))
		}
	}

	// AWS SDK call
	usage, err := callLambda()
	if err != nil {
		return "ERROR", err
	}
	return usage, nil
}

func main() {
	runtime.Start(handleRequest)
}
