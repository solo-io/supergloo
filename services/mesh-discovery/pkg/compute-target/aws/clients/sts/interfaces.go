package sts

import "github.com/aws/aws-sdk-go/service/sts"

//go:generate mockgen -source interfaces.go -destination ./mocks/interfaces.go

type STSClient interface {
	GetCallerIdentity() (*sts.GetCallerIdentityOutput, error)
}
