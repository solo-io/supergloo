// +build wireinject

package wire

import (
	"context"

	k8s_core_providers "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/providers"
	smh_security_providers "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/providers"

	"github.com/google/wire"
	mc_wire "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/wire"
	csr_generator "github.com/solo-io/service-mesh-hub/pkg/common/csr-generator"
	"github.com/solo-io/service-mesh-hub/pkg/common/csr/certgen"
)

func InitializeCsrAgent(ctx context.Context) (CsrAgentContext, error) {
	wire.Build(
		mc_wire.ClusterProviderSet,
		certgen.NewSigner,
		k8s_core_providers.SecretClientProvider,
		csr_generator.NewCertClient,
		csr_generator.NewVirtualMeshCSRDataSourceFactory,
		csr_generator.CsrControllerProviderLocal,
		csr_generator.IstioCSRGeneratorSet,
		csr_generator.NewPrivateKeyGenerator,
		CsrAgentContextProvider,
		smh_security_providers.VirtualMeshCertificateSigningRequestClientProvider,
	)

	return CsrAgentContext{}, nil
}
