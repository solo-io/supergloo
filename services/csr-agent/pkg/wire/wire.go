// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	"github.com/solo-io/mesh-projects/pkg/security/certgen"
	mc_wire "github.com/solo-io/mesh-projects/services/common/multicluster/wire"
	csr_generator "github.com/solo-io/mesh-projects/services/csr-agent/pkg/csr-generator"
)

func InitializeCsrAgent(ctx context.Context) (CsrAgentContext, error) {
	wire.Build(
		mc_wire.ClusterProviderSet,
		certgen.NewSigner,
		kubernetes_core.NewSecretsClient,
		csr_generator.NewCertClient,
		zephyr_security.NewMeshGroupCSRClient,
		csr_generator.NewMeshGroupCSRDataSourceFactory,
		csr_generator.CsrControllerProviderLocal,
		csr_generator.IstioCSRGeneratorSet,
		CsrAgentContextProvider,
	)

	return CsrAgentContext{}, nil
}
