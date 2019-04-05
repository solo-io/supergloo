package appmesh

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	webhookName            = "sidecar-injector"
	webhookImageName       = "quay.io/solo-io/sidecar-injector"
	webhookImagePullPolicy = "Always"
	resourcesConfigMapName = "sidecar-injection-resources"
	secretKind             = "Secret"
	webhookConfigKind      = "MutatingWebhookConfiguration"
	deploymentKind         = "Deployment"
	serviceKind            = "Service"
	configMapKind          = "ConfigMap"
)

type appMeshSyncer struct {
	reporter   reporter.Reporter
	reconciler AutoInjectionReconciler
	validator  Validator
}

func NewAppMeshRegistrationSyncer(
	reporter reporter.Reporter,
	kube kubernetes.Interface,
	secretClient gloov1.SecretClient,
	installer Installer) v1.RegistrationSyncer {
	return &appMeshSyncer{
		reporter:   reporter,
		reconciler: NewAutoInjectionReconciler(kube, installer),
		validator:  NewAppMeshValidator(kube, secretClient),
	}
}

func (s *appMeshSyncer) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("aws-app-mesh-registration-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v", snap.Hash())
	defer logger.Infof("end sync %v", snap.Hash())
	logger.Debugf("full snapshot: %v", snap)

	var (
		errors               *multierror.Error
		resourceErrors       = reporter.ResourceErrors{}
		autoInjectionEnabled bool
	)

	// Validate the AWS App Mesh meshes
	for _, mesh := range snap.Meshes.List() {
		appMesh := mesh.GetAwsAppMesh()

		if appMesh == nil {
			// Mesh is not of type AwsAppMesh
			continue
		}

		if err := s.validator.Validate(ctx, appMesh); err != nil {
			resourceErrors.AddError(mesh, err)
			continue
		}

		// Note: all meshes share the same auto-injection infrastructure
		if appMesh.EnableAutoInject {
			autoInjectionEnabled = true
		}
	}

	// Write error reports
	errors = multierror.Append(errors, s.reporter.WriteReports(ctx, resourceErrors, nil))

	// All mesh objects are valid, we now have to reconcile the auto-injection components
	if err := s.reconciler.Reconcile(autoInjectionEnabled); err != nil {
		errors = multierror.Append(errors, err)
	}

	return errors.ErrorOrNil()
}
