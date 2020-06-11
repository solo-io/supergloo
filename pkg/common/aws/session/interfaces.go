package session

import "github.com/aws/aws-sdk-go/aws/session"

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

// Initialize, store, and retrieve AWS API sessions.
type AwsSessionStore interface {
	// Initialize and store a new AWS API session for the given account ID and region.
	Add(accountId, region string) error
	// Retrieve the stored session by AWS account ID and region.
	Get(accountId, region string) (*session.Session, error)
}
