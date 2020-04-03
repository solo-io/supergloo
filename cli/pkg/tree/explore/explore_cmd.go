package explore

import (
	"fmt"
	"strings"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/cliconstants"
	"github.com/solo-io/mesh-projects/cli/pkg/tree/explore/exploration"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/sets"
)

type ExploreCommand *cobra.Command

var (
	ExploreSet = wire.NewSet(
		ExploreCmd,
	)

	validResources = sets.NewString("service", "workload")
)

func ExploreCmd() ExploreCommand {
	exploreCommand := cliconstants.ExploreCommand(validResources.List())
	cmd := &cobra.Command{
		Use:   exploreCommand.Use,
		Short: exploreCommand.Short,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			resourceType := args[0]
			if !validResources.Has(resourceType) {
				return eris.Errorf("Unsupported resource type: %s - expected one of [%s]", resourceType, strings.Join(validResources.List(), ", "))
			}
			resourceName, err := exploration.ParseResourceName(args[1])
			if err != nil {
				return err
			}
			fmt.Println(resourceName)
			return nil
		},
	}

	return cmd
}
