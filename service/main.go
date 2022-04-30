package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
)

func HandleRequest(ctx context.Context) (string, error) {
	return fmt.Sprintf("what is happening.."), nil
}

func main() {
	lambda.Start(HandleRequest)
}
