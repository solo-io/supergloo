package get

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	"github.com/solo-io/supergloo/pkg2/constants"
	"github.com/spf13/cobra"
	k8s "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getResourcesCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resources",
		Short:   `Displays resources that can be displayed`,
		Aliases: []string{"r", "options"},
		Args:    cobra.ExactArgs(0),
		RunE: func(c *cobra.Command, args []string) error {
			crdClient, err := common.GetKubeCrdClient()
			if err != nil {
				return err
			}
			crdList, err := (*crdClient).List(k8s.ListOptions{})
			if err != nil {
				return fmt.Errorf("Error retrieving supergloo resource types. Cause: %v \n", err)
			}
			table := tablewriter.NewWriter(os.Stdout)
			table.SetBorder(false)
			table.SetHeader([]string{"", "resource", "plural", "short names"})
			index := 1
			for _, crd := range crdList.Items {
				if strings.Contains(crd.Name, common.SuperglooGroupName) {
					// TODO (EItanya) think of a better way to deal with this
					nameSpec := crd.Spec.Names
					if nameSpec.Singular != "install" {
						table.Append([]string{strconv.Itoa(index), nameSpec.Singular, nameSpec.Plural, strings.Join(nameSpec.ShortNames, ",")})
						index++
					}
				}
			}
			table.Render()
			return nil
		},
	}
	getOpts := &opts.Get
	flags := cmd.Flags()
	flags.StringVarP(&getOpts.Output, "output", "o", "",
		"Output format. Options include: \n"+strings.Join(supportedOutputFormats, "|"))

	flags.StringVarP(&getOpts.Namespace, "namespace", "n", constants.SuperglooNamespace, "namespace to search")
	return cmd
}
