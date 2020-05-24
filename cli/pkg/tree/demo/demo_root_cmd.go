package demo

import (
	"github.com/google/wire"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	appmesh_eks "github.com/solo-io/service-mesh-hub/cli/pkg/tree/demo/appmesh-eks"
	istio_multicluster "github.com/solo-io/service-mesh-hub/cli/pkg/tree/demo/istio-multicluster"
	"github.com/spf13/cobra"
)

type DemoCommand *cobra.Command

var DemoSet = wire.NewSet(
	appmesh_eks.AppmeshEksSet,
	istio_multicluster.IstioMulticlusterSet,
	DemoRootCmd,
)

func DemoRootCmd(
	appmeshEksCmd appmesh_eks.AppmeshEksCmd,
	istioMulticlusterCmd istio_multicluster.IstioMulticlusterCmd,
) DemoCommand {
	demo := &cobra.Command{
		Use:   cliconstants.DemoCommand.Use,
		Short: cliconstants.DemoCommand.Short,
		RunE:  common.NonTerminalCommand(cliconstants.DemoCommand.Use),
	}
	demo.AddCommand(
		istioMulticlusterCmd,
		appmeshEksCmd,
	)
	return demo
}
