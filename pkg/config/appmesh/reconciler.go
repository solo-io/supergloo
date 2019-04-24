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

func (r *reconciler) Reconcile(ctx context.Context, mesh *v1.Mesh, desiredResources *translator.ResourceSnapshot) error {

	// Get instance of AppMesh API client
	client, err := r.clientBuilder.GetClientInstance(mesh.GetAwsAppMesh().AwsSecret, mesh.GetAwsAppMesh().Region)
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve AppMesh API client instance")
	}

	existingMesh, err := client.GetMesh(ctx, mesh.Metadata.Name)
	if err != nil {
		return err
	}

	// TODO: implement
	if existingMesh == nil {
		// Mesh does not exist, this is the easy case: just create everything
	} else {
		// 1. List the existing resources of each type
		// 2. For each of the desired resources in the snapshot:
		//   - if it does not exist, create it
		//   - if it does already exist, overwrite the existing one
		// 3. For each of the existing resources, if it has no correspondent desired resource, then delete it
	}

	return nil
}
