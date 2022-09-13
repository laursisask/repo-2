package main

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

func TestMain(t *testing.T) {
	d := time.Now().Add(50 * time.Millisecond)
	os.Setenv("AWS_LAMBDA_FUNCTION_NAME", "humio-ingest")
	ctx, _ := context.WithDeadline(context.Background(), d)
	ctx = lambdacontext.NewContext(ctx, &lambdacontext.LambdaContext{
		AwsRequestID:       "495b12a8-xmpl-4eca-8168-160484189f99",
		InvokedFunctionArn: "arn:aws:lambda:us-east-2:123456789012:function:humio-ingest",
	})
	inputJson := ReadJSONFromFile(t, "../event.json")
	var event incommingEvent
	err := json.Unmarshal(inputJson, &event)
	if err != nil {
		t.Errorf("could not unmarshal event. details: %v", err)
	}
	//var inputEvent SQSEvent
	err = handleRequest(ctx, event)
	if err != nil {
		t.Log(err)
	}
}

func ReadJSONFromFile(t *testing.T, inputFile string) []byte {
	inputJSON, err := ioutil.ReadFile(inputFile)
	if err != nil {
		t.Errorf("could not open test file. details: %v", err)
	}

	return inputJSON
}
