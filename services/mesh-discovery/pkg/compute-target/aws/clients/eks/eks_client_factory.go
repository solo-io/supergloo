package eks

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/solo-io/skv2/pkg/multicluster/discovery/cloud"
)

type EksClientFactory func(creds *credentials.Credentials, region string) (cloud.EksClient, error)

func EksClientFactoryProvider() EksClientFactory {
	return func(creds *credentials.Credentials, region string) (cloud.EksClient, error) {
		return cloud.NewEksClient(region, creds)
	}
}
