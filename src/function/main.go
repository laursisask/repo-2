package main

import (
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
	"log"
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

type internalSNSMessage struct {
	S3Bucket    string   `json:"s3Bucket"`    // S3 bucket name
	S3ObjectKey []string `json:"s3ObjectKey"` // Path inside S3 bucket
}

func acquireS3File(ctx context.Context, bucketName, path string) error {
	s3client := s3.New(awsSession)
	r, err := s3client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot acquire object %s from bucket %s: %s", path, bucketName, err))
	}

	log.Printf("TODO: acquired: %v", r)
	return nil
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
	for _, s3path := range msg.S3ObjectKey {
		err = acquireS3File(ctx, msg.S3Bucket, s3path)
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
