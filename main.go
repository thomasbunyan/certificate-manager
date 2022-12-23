package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/thomasbunyan/certificate-manager/internal/certlambda"
)

func HandleRequest(event certlambda.Event) error {
	return certlambda.Run(event)
}

func main() {
	lambda.Start(HandleRequest)
}
