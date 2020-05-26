package sts

import "github.com/aws/aws-sdk-go/service/sts"

//go:generate mockgen -source interfaces.go -destination ./mocks/interfaces.go

type STSClient interface {
	// Retrieves caller identity metadata by making a request to AWS STS (Secure Token Service).
	GetCallerIdentity() (*sts.GetCallerIdentityOutput, error)
}
