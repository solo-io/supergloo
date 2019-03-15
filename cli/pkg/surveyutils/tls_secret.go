package surveyutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"

	"github.com/solo-io/gloo/pkg/cliutil"
)

func SurveyTlsSecret(secret *options.CreateTlsSecret) error {
	if err := cliutil.GetStringInput("path to root-cert file", &secret.RootCaFilename); err != nil {
		return err
	}
	if err := cliutil.GetStringInput("path to ca-cert file", &secret.PrivateKeyFilename); err != nil {
		return err
	}
	if err := cliutil.GetStringInput("path to ca-key file", &secret.CaCertFilename); err != nil {
		return err
	}
	if err := cliutil.GetStringInput("path to cert-chain file", &secret.CertChainFilename); err != nil {
		return err
	}
	return nil
}
