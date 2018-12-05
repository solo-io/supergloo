package get

import (
	"fmt"
	"strings"

	"github.com/solo-io/supergloo/cli/pkg/cmd/get/info"
	"github.com/solo-io/supergloo/cli/pkg/common"

	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/spf13/cobra"
	"gopkg.in/AlecAivazis/survey.v1"
)

var supportedOutputFormats = []string{"wide", "yaml"}

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: `Display one or many supergloo resources`,
		Long:  `Display one or many supergloo resources`,
		Args:  cobra.RangeArgs(0, 2),
		Run: func(c *cobra.Command, args []string) {
			if err := get(args, opts); err != nil {
				fmt.Println(err)
			}
		},
	}
	getOpts := &opts.Get
	pFlags := cmd.Flags()
	pFlags.StringVarP(&getOpts.Output, "output", "o", "",
		"Output format. Options include: \n"+strings.Join(supportedOutputFormats, "|"))
	return cmd
}

func get(args []string, opts *options.Options) error {

	infoClient, err := info.NewClient()
	if err != nil {
		return err
	}

	if err := ensureParameters(infoClient, opts, args); err != nil {
		return err
	}

	return getResource(infoClient, opts.Get)
}

func ensureParameters(infoClient info.SuperglooInfoClient, opts *options.Options, args []string) error {
	gOpts := &opts.Get

	// Get available resource types
	resourceTypes := infoClient.ListResourceTypes()

	// Argument count is validated by cobra.RangeArgs
	if len(args) == 0 {
		if err := selectResourceInteractive(resourceTypes, opts); err != nil {
			return err
		}
	} else {

		// first arg is the resource type
		gOpts.Type = args[0]
		// second arg is the resource name (optional)
		gOpts.Name = ""
		if len(args) == 2 {
			gOpts.Name = args[1]
		}
	}

	// Validate input resource type
	if !common.Contains(resourceTypes, gOpts.Type) {
		return errors.Errorf(common.UnknownResourceTypeMsg, gOpts.Type)
	}

	// Output format is set by a flag
	if gOpts.Output != "" && !common.Contains(supportedOutputFormats, gOpts.Output) {
		return errors.Errorf(common.UnknownOutputFormat, gOpts.Output, strings.Join(supportedOutputFormats, "|"))
	}

	return nil
}

func getResource(infoClient info.SuperglooInfoClient, gOpts options.Get) error {

	// Fetch the resource information
	err := infoClient.ListResources(gOpts)
	if err != nil {
		return err
	}
	return nil
}

func selectResourceInteractive(resourceTypes []string, opts *options.Options) error {
	chosenResourceType, err := chooseResourceType(resourceTypes)
	if err != nil {
		return err
	}
	opts.Get.Type = chosenResourceType
	return nil
}

func chooseResourceType(resourceTypes []string) (string, error) {

	question := &survey.Select{
		Message: "Select a resource type",
		Options: resourceTypes,
	}

	var choice string
	if err := survey.AskOne(question, &choice, survey.Required); err != nil {
		// this should not error
		fmt.Println("error with input")
		return "", err
	}

	return choice, nil
}
