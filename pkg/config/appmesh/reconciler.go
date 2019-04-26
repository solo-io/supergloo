package appmesh

import (
	"context"

	translator "github.com/solo-io/supergloo/pkg/translator/appmesh"

	"github.com/pkg/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

type Reconciler interface {
	Reconcile(ctx context.Context, mesh *v1.Mesh, desiredResources *translator.ResourceSnapshot) error
}

func NewReconciler(builder ClientBuilder) Reconciler {
	return &reconciler{
		clientBuilder: builder,
	}
}

type reconciler struct {
	clientBuilder ClientBuilder
}

func (r *reconciler) Reconcile(ctx context.Context, mesh *v1.Mesh, snapshot *translator.ResourceSnapshot) error {

	// Get instance of AppMesh API client
	client, err := r.clientBuilder.GetClientInstance(mesh.GetAwsAppMesh().AwsSecret, mesh.GetAwsAppMesh().Region)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve AppMesh API client instance")
	}

	existingMesh, err := client.GetMesh(ctx, mesh.Metadata.Name)
	if err != nil {
		return err
	}

	helper := newHelper(ctx, client, snapshot)
	if existingMesh == nil {
		if err := helper.createAll(); err != nil {
			return err
		}
	} else {
		if err := helper.reconcile(); err != nil {
			return err
		}
	}
	return nil
}
