// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	"github.com/solo-io/service-mesh-hub/pkg/security/certgen"
	mc_wire "github.com/solo-io/service-mesh-hub/services/common/multicluster/wire"
	csr_generator "github.com/solo-io/service-mesh-hub/services/csr-agent/pkg/csr-generator"
)

func InitializeCsrAgent(ctx context.Context) (CsrAgentContext, error) {
	wire.Build(
		mc_wire.ClusterProviderSet,
		certgen.NewSigner,
		kubernetes_core.NewSecretClient,
		csr_generator.NewCertClient,
		csr_generator.NewVirtualMeshCSRDataSourceFactory,
		csr_generator.CsrControllerProviderLocal,
		csr_generator.IstioCSRGeneratorSet,
		csr_generator.NewPrivateKeyGenerator,
		CsrAgentContextProvider,
		zephyr_security.VirtualMeshCertificateSigningRequestClientProvider,
	)

	return CsrAgentContext{}, nil
}
