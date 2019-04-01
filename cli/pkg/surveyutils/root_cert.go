package surveyutils

import (
	"context"

	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/options"
)

func SurveySetRootCert(ctx context.Context, in *options.SetRootCert) error {
	mesh, err := SurveyMesh("select the mesh for which you wish to set the root cert", ctx)
	if err != nil {
		return err
	}
	tlsSecret, err := SurveyTlsSecret("select the tls secret to use as the new root cert", ctx)
	if err != nil {
		return err
	}
	in.TargetMesh = options.ResourceRefValue(mesh)
	in.TlsSecret = options.ResourceRefValue(tlsSecret)
	return nil
}

func SurveyTlsSecret(prompt string, ctx context.Context) (core.ResourceRef, error) {
	tlsSecretes, err := clients.MustTlsSecretClient().List("", skclients.ListOpts{Ctx: ctx})
	if err != nil {
		return core.ResourceRef{}, err
	}

	return surveyResources("tls secrets", prompt, "<unset custom cert>", tlsSecretes.AsResources())
}
