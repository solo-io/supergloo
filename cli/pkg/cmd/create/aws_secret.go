package create

import (
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/spf13/cobra"
)

func AwsSecretCmd(opts *options.Options) *cobra.Command {
	sOpts := &(opts.Create).AwsSecret
	cmd := &cobra.Command{
		Use:   "aws-secret",
		Short: `Create an AWS secret with the given name`,
		Long:  `Create an AWS secret with the given name`,
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			// make sure the given args are valid
			if err := validateAwsSecretArgs(opts); err != nil {
				fmt.Println(err)
				return
			}
			// gather any missing args that are available through interactive mode
			if err := gatherAwsSecretArgs(args, opts); err != nil {
				fmt.Println(err)
				return
			}
			// create the secret
			if err := createAwsSecret(opts); err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Created aws secret [%v] in namespace [%v]\n", args[0], sOpts.Namespace)
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&sOpts.AccessKey, "access-key", "", "aws access key")
	flags.StringVar(&sOpts.SecretKey, "secret-key", "", "aws secret key")

	flags.StringVar(&sOpts.Namespace, "secretnamespace", "", "namespace in which to store the secret")

	return cmd
}

func validateAwsSecretArgs(opts *options.Options) error {
	sOpts := &(opts.Create).AwsSecret
	// check if we are interactive mode
	if opts.Top.Static && sOpts.Namespace == "" {
		return fmt.Errorf("please provide a namespace for the secret")
	}

	if opts.Top.Static && sOpts.AccessKey == "" {
		return fmt.Errorf("please provide an AWS access key")
	}

	if opts.Top.Static && sOpts.SecretKey == "" {
		return fmt.Errorf("please provide an AWS secret key")
	}

	if sOpts.Namespace != "" {
		if !common.Contains(opts.Cache.Namespaces, sOpts.Namespace) {
			return fmt.Errorf("please provide a valid namespace for the secret. %v does not exist", sOpts.Namespace)
		}
	}
	return nil
}

func gatherAwsSecretArgs(args []string, opts *options.Options) error {
	sOpts := &(opts.Create).AwsSecret

	// apply the args
	sOpts.Name = args[0]

	// check if we are interactive mode
	if opts.Top.Static {
		return nil
	}

	if sOpts.Namespace == "" {
		namespace, err := common.ChooseNamespace(opts, "Select a namespace")
		if err != nil {
			return fmt.Errorf("input error")
		}
		sOpts.Namespace = namespace
	}

	if sOpts.SecretKey == "" {
		secretKey, err := common.GetString("AWS Secret Key")
		if err != nil {
			return fmt.Errorf("input error")
		}
		sOpts.SecretKey = secretKey
	}

	if sOpts.AccessKey == "" {
		accessKey, err := common.GetString("AWS Access Key")
		if err != nil {
			return fmt.Errorf("input error")
		}
		sOpts.AccessKey = accessKey
	}
	return nil
}

func createAwsSecret(opts *options.Options) error {
	sOpts := &(opts.Create).AwsSecret

	secret := &gloov1.Secret{
		Metadata: core.Metadata{
			Name:      sOpts.Name,
			Namespace: sOpts.Namespace,
		},
		Kind: &gloov1.Secret_Aws{
			Aws: &gloov1.AwsSecret{
				// these can be read in from ~/.aws/credentials by default (if user does not provide)
				// see https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/configuring-sdk.html for more details
				AccessKey: sOpts.AccessKey,
				SecretKey: sOpts.SecretKey,
			},
		},
	}

	secretClient, err := common.GetGlooSecretClient()
	if err != nil {
		return err
	}
	_, err = (*secretClient).Write(secret, clients.WriteOpts{})
	if err != nil {
		return err
	}
	return nil
}
