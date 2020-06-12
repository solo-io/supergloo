package cloud

import (
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/clients"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/compute-target/aws/clients/eks_temp"
)

// Contains rest clients to AWS cloud services.
type AwsCloud struct {
	Appmesh clients.AppmeshClient
	Eks     eks_temp.EksClient
}
