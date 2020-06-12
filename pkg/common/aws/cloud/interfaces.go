package cloud

import "github.com/aws/aws-sdk-go/aws/credentials"

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

// Initialize, awsCloudStore, and retrieve AwsCloud interface.
type AwsCloudStore interface {
	// Store AWS credentials by accountID, which may be used to instantiate a new client upon calling Get.
	Add(accountId string, creds *credentials.Credentials)
	// Retrieve the stored session by AwsCloud ID and region.
	Get(accountId, region string) (*AwsCloud, error)
}
