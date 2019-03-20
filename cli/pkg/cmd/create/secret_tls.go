package create

import (
	"io/ioutil"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	"github.com/solo-io/go-utils/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/spf13/cobra"
)

func tlsCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tls",
		Short: `create a tls secret with cert`,
		Long:  `Create a secret with the given name`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveyMetadata("Routing Rule", &opts.Metadata); err != nil {
					return err
				}
				if err := surveyutils.SurveyCreateTlsSecret(&opts.CreateTlsSecret); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			if err := createTlsSecret(opts); err != nil {
				return err
			}
			return nil
		},
	}

	flagutils.AddCreateTlsSecretFlags(cmd.PersistentFlags(), &opts.CreateTlsSecret)

	return cmd
}

func createTlsSecret(opts *options.Options) error {
	if opts.Metadata.Name == "" {
		return errors.Errorf("name cannot be empty, provide with --name flag")
	}

	// read the values
	rootCa, err := ioutil.ReadFile(opts.CreateTlsSecret.RootCaFilename)
	if err != nil {
		return errors.Wrapf(err, "reading rootca file: %v", opts.CreateTlsSecret.RootCaFilename)
	}
	privateKey, err := ioutil.ReadFile(opts.CreateTlsSecret.PrivateKeyFilename)
	if err != nil {
		return errors.Wrapf(err, "reading privatekey file: %v", opts.CreateTlsSecret.PrivateKeyFilename)
	}
	certChain, err := ioutil.ReadFile(opts.CreateTlsSecret.CertChainFilename)
	if err != nil {
		return errors.Wrapf(err, "reading certchain file: %v", opts.CreateTlsSecret.CertChainFilename)
	}
	caCert, err := ioutil.ReadFile(opts.CreateTlsSecret.CaCertFilename)
	if err != nil {
		return errors.Wrapf(err, "reading cacert file: %v", opts.CreateTlsSecret.CaCertFilename)
	}

	secret := &v1.TlsSecret{
		Metadata:  opts.Metadata,
		CertChain: string(certChain),
		CaKey:     string(privateKey),
		RootCert:  string(rootCa),
		CaCert:    string(caCert),
	}

	secretClient := clients.MustTlsSecretClient()

	secret, err = secretClient.Write(secret, skclients.WriteOpts{Ctx: opts.Ctx})
	if err != nil {
		return errors.Wrapf(err, "writing secret to storage")
	}

	helpers.PrintTlsSecrets(v1.TlsSecretList{secret}, opts.OutputType)

	return nil
}
