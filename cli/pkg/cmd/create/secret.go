package create

import (
	"fmt"
	"io/ioutil"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	istiosecret "github.com/solo-io/supergloo/pkg/api/external/istio/encryption/v1"
	"github.com/spf13/cobra"
)

func SecretCmd(opts *options.Options) *cobra.Command {
	sOpts := &(opts.Create).Secret
	cmd := &cobra.Command{
		Use:   "secret",
		Short: `Create a secret with the given name`,
		Long:  `Create a secret with the given name`,
		Args:  cobra.ExactArgs(1),
		Run: func(c *cobra.Command, args []string) {
			// make sure the given args are valid
			if err := validateSecretArgs(opts); err != nil {
				fmt.Println(err)
				return
			}
			// gather any missing args that are available through interactive mode
			if err := gatherSecretArgs(args, opts); err != nil {
				fmt.Println(err)
				return
			}
			// create the secret
			if err := createSecret(opts); err != nil {
				fmt.Println(err)
				return
			}
			fmt.Printf("Created secret [%v] in namespace [%v]\n", args[0], sOpts.Namespace)
		},
	}

	flags := cmd.Flags()

	flags.StringVar(&sOpts.RootCa, "rootca", "", "filename of rootca for secret")
	cmd.MarkFlagRequired("rootca")

	flags.StringVar(&sOpts.PrivateKey, "privatekey", "", "filename of privatekey for secret")
	cmd.MarkFlagRequired("privatekey")

	flags.StringVar(&sOpts.CertChain, "certchain", "", "filename of certchain for secret")
	cmd.MarkFlagRequired("certchain")

	flags.StringVar(&sOpts.Namespace, "secretnamespace", "", "namespace in which to store the secret")

	return cmd
}

func validateSecretArgs(opts *options.Options) error {
	sOpts := &(opts.Create).Secret
	// check if we are interactive mode
	if opts.Top.Static && sOpts.Namespace == "" {
		return fmt.Errorf("please provide a namespace for the secret")
	}
	if sOpts.Namespace != "" {
		if !common.Contains(opts.Cache.Namespaces, sOpts.Namespace) {
			return fmt.Errorf("please provide a valid namespace for the secret. %v does not exist", sOpts.Namespace)
		}
	}
	return nil
}

func gatherSecretArgs(args []string, opts *options.Options) error {
	sOpts := &(opts.Create).Secret

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
	return nil
}

func createSecret(opts *options.Options) error {
	sOpts := &(opts.Create).Secret

	// read the values
	rootCa, err := ioutil.ReadFile(sOpts.RootCa)
	if err != nil {
		return fmt.Errorf("Error while reading rootca file: %v\n%v", sOpts.RootCa, err)
	}
	privateKey, err := ioutil.ReadFile(sOpts.PrivateKey)
	if err != nil {
		return fmt.Errorf("Error while reading private key file: %v\n%v", sOpts.PrivateKey, err)
	}
	certChain, err := ioutil.ReadFile(sOpts.CertChain)
	if err != nil {
		return fmt.Errorf("Error while reading certchain file: %v\n%v", sOpts.CertChain, err)
	}

	secret := &istiosecret.IstioCacertsSecret{
		Metadata: core.Metadata{
			Namespace: sOpts.Namespace,
			Name:      sOpts.Name,
		},
		CertChain: string(certChain),
		RootCert:  string(rootCa),
		CaCert:    string(rootCa),
		CaKey:     string(privateKey),
	}
	secretClient, err := common.GetIstioSecretClient()
	if err != nil {
		return err
	}
	_, err = (*secretClient).Write(secret, clients.WriteOpts{})
	if err != nil {
		return err
	}
	return nil
}
