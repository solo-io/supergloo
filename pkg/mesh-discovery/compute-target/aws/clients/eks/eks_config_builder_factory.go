package eks

import (
	"github.com/solo-io/skv2/pkg/multicluster/discovery"
	"github.com/solo-io/skv2/pkg/multicluster/discovery/cloud"
)

type EksConfigBuilderFactory func(eksClient cloud.EksClient) discovery.EksConfigBuilder

func EksConfigBuilderFactoryProvider() EksConfigBuilderFactory {
	return func(eksClient cloud.EksClient) discovery.EksConfigBuilder {
		return discovery.NewEksConfigBuilder(eksClient)
	}
}
