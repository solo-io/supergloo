package get

import (
	"fmt"
	"strings"

	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/clients"
	"github.com/solo-io/supergloo/cli/pkg/constants"
	"github.com/solo-io/supergloo/cli/pkg/util"
	"github.com/spf13/cobra"
)

var test string

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: `Display one or many supergloo resources`,
		Long:  `Display one or many supergloo resources`,
		Args:  cobra.RangeArgs(1, 2),
		Run: func(c *cobra.Command, args []string) {
			err := get(args)
			if err != nil {
				fmt.Println(err)
			}
		},
	}
	// TODO: handle flags
	return cmd
}

func get(args []string) error {
	if argNumber := len(args); argNumber == 1 {
		return getResource(args[0], "")
	} else {
		// Show the resource of the given type with the given name
		return getResource(args[0], args[1])
	}
}

func getResource(resourceType, resourceName string) error {

	sgClient, err := clients.NewClient()
	if err != nil {
		return err
	}

	resourceTypes, err := sgClient.ListResourceTypes()
	if err != nil {
		return err
	}

	// Validate resource type
	if !util.Contains(resourceTypes, resourceType) {
		return errors.Errorf(constants.UnknownResourceTypeMsg, resourceType)
	}

	// Fetch the resource information
	info, err := sgClient.ListResources(resourceType, resourceName)
	if err != nil {
		return err
	}

	// print the information to stdout
	_, err = fmt.Println(info.Headers())
	if err != nil {
		return err
	}
	_, err = fmt.Println(strings.Join(info.Resources(), "\n"))

	return err
}
