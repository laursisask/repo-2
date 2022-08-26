package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	runtime "github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/lambda"
	"log"
)

var client = lambda.New(session.New())

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

func handleSNSNotification(ctx context.Context, notification events.SNSEntity) error {
	if notification.Type != "Notification" {
		return errors.New(fmt.Sprintf("Unexpected SNS entity type: %s", notification.Type))
	}
	var msg internalSNSMessage
	err := json.Unmarshal([]byte(notification.Message), &msg)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot parse Message inside SNSEntity: %s", err))
	}

	log.Printf("TODO: acquire %s ----> %s", msg.S3Bucket, msg.S3ObjectKey)
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
