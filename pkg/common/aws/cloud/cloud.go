package cloud

import (
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/clients"
	"github.com/solo-io/skv2/pkg/multicluster/discovery/cloud"
)

// Contains rest clients to AWS cloud services.
type AwsCloud struct {
	Appmesh clients.AppmeshClient
	Eks     cloud.EksClient
	Sts     clients.STSClient
}
