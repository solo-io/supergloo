package aws

import "github.com/aws/aws-sdk-go/aws/credentials"

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

// Map an AWS account ID to AWS credentials.
type AwsCredentialsGetter interface {
	Get(accountId string) (*credentials.Credentials, error)
	Set(accountId string, creds *credentials.Credentials)
	Remove(accountId string)
}
