// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	mc_wire "github.com/solo-io/service-mesh-hub/pkg/common/compute-target/wire"
	csr_generator "github.com/solo-io/service-mesh-hub/pkg/common/csr-generator"
	"github.com/solo-io/service-mesh-hub/pkg/common/csr/certgen"
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
		smh_security.VirtualMeshCertificateSigningRequestClientProvider,
	)

	return CsrAgentContext{}, nil
}
