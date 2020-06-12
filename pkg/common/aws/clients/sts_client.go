package clients

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type STSClientFactory func(creds *credentials.Credentials, region string) (STSClient, error)

func STSClientFactoryProvider() STSClientFactory {
	return func(creds *credentials.Credentials, region string) (STSClient, error) {
		sess, err := session.NewSession(&aws.Config{
			Credentials: creds,
			Region:      aws.String(region),
		})
		if err != nil {
			return nil, err
		}
		return NewSTSClient(sess), nil
	}
}

type stsClient struct {
	client *sts.STS
}

func NewSTSClient(sess *session.Session) STSClient {
	return &stsClient{client: sts.New(sess)}
}

func (s *stsClient) GetCallerIdentity() (*sts.GetCallerIdentityOutput, error) {
	callerIdentity, err := s.client.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	return callerIdentity, nil
}
