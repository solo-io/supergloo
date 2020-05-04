// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/security/certgen"
	mc_wire "github.com/solo-io/service-mesh-hub/services/common/compute-target/wire"
	csr_generator "github.com/solo-io/service-mesh-hub/services/csr-agent/pkg/csr-generator"
)

func InitializeCsrAgent(ctx context.Context) (CsrAgentContext, error) {
	wire.Build(
		mc_wire.ClusterProviderSet,
		certgen.NewSigner,
		k8s_core.SecretClientProvider,
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
