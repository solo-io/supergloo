// +build wireinject

package wire

import (
	"context"

	"github.com/google/wire"
	kubernetes_core "github.com/solo-io/mesh-projects/pkg/clients/kubernetes/core"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	"github.com/solo-io/mesh-projects/pkg/security/certgen"
	mc_wire "github.com/solo-io/mesh-projects/services/common/multicluster/wire"
	csr_agent_controller "github.com/solo-io/mesh-projects/services/csr-agent/pkg/controller"
)

func InitializeCsrAgent(ctx context.Context) (CsrAgentContext, error) {
	wire.Build(
		mc_wire.ClusterProviderSet,
		certgen.NewSigner,
		kubernetes_core.NewSecretsClient,
		csr_agent_controller.NewCertClient,
		zephyr_security.NewMeshGroupCertificateSigningRequestClient,
		csr_agent_controller.NewCsrAgentIstioProcessor,
		csr_agent_controller.CsrAgentPredicateProvider,
		csr_agent_controller.NewCsrAgentEventHandler,
		csr_agent_controller.CsrControllerProviderLocal,
		CsrAgentContextProvider,
	)

	return CsrAgentContext{}, nil
}
