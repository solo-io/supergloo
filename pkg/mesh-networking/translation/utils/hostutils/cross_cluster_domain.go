package hostutils

import (
	"fmt"
	"github.com/solo-io/skv2/pkg/ezkube"
)

// GetCrossClusterServiceFQDN returns the fully qualified service hostname used to resolve traffic to a remote cluster
func GetCrossClusterServiceFQDN(serviceRef ezkube.ClusterResourceId) string {
	return fmt.Sprintf("%s.%s.svc.%s", serviceRef.GetName(), serviceRef.GetNamespace(), serviceRef.GetClusterName())
}
