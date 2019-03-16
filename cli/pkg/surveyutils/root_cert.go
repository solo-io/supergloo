package surveyutils

import (
	"context"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
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

func SurveyMesh(prompt string, ctx context.Context) (core.ResourceRef, error) {
	meshes, err := helpers.MustMeshClient().List("", clients.ListOpts{Ctx: ctx})
	if err != nil {
		return core.ResourceRef{}, err
	}

	byKey := make(map[string]core.ResourceRef)
	var keys []string
	for _, resource := range meshes {
		ref := resource.Metadata.Ref()
		byKey[ref.Key()] = ref
		keys = append(keys, ref.Key())
	}

	if len(keys) == 0 {
		return core.ResourceRef{}, errors.Errorf("no meshes found. create one first.")
	}

	var key string
	if err := cliutil.ChooseFromList(
		prompt,
		&key,
		keys,
	); err != nil {
		return core.ResourceRef{}, err
	}

	ref, ok := byKey[key]
	if !ok {
		return core.ResourceRef{}, errors.Errorf("internal error: mesh map missing key %v", key)
	}

	return ref, nil
}

func SurveyTlsSecret(prompt string, ctx context.Context) (core.ResourceRef, error) {
	tlsSecretes, err := helpers.MustTlsSecretClient().List("", clients.ListOpts{Ctx: ctx})
	if err != nil {
		return core.ResourceRef{}, err
	}

	byKey := make(map[string]core.ResourceRef)
	var keys []string
	for _, resource := range tlsSecretes {
		ref := resource.Metadata.Ref()
		byKey[ref.Key()] = ref
		keys = append(keys, ref.Key())
	}

	if len(keys) == 0 {
		return core.ResourceRef{}, errors.Errorf("no tlsSecretes found. create one first.")
	}

	var key string
	if err := cliutil.ChooseFromList(
		prompt,
		&key,
		keys,
	); err != nil {
		return core.ResourceRef{}, err
	}

	ref, ok := byKey[key]
	if !ok {
		return core.ResourceRef{}, errors.Errorf("internal error: tlsSecret map missing key %v", key)
	}

	return ref, nil
}
