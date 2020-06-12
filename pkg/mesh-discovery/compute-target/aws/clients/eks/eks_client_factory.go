package eks

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/compute-target/aws/clients/eks_temp"
	"github.com/solo-io/skv2/pkg/multicluster/discovery/cloud"
)

type EksClientFactory func(sess *session.Session) cloud.EksClient

func EksClientFactoryProvider() EksClientFactory {
	return func(sess *session.Session) cloud.EksClient {
		return eks_temp.NewEksClient(sess)
	}
}
