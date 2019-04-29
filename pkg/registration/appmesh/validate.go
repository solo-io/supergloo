package appmesh

import (
	"context"
	"strings"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	cliclients "github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
)

const AppMeshAvailableRegions = "ap-south-1,ap-northeast-2,ap-southeast-1,ap-southeast-2,ap-northeast-1," +
	"us-east-2,us-east-1,us-west-1,us-west-2,eu-west-1,eu-central-1,ca-central-1"

type Validator interface {
	Validate(ctx context.Context, appMesh *v1.AwsAppMesh) error
}

type validator struct {
	kube         kubernetes.Interface
	secretClient gloov1.SecretClient
}

func NewAppMeshValidator(kube kubernetes.Interface, secretClient gloov1.SecretClient) Validator {
	return &validator{kube: kube, secretClient: secretClient}
}

func (v *validator) Validate(ctx context.Context, appMesh *v1.AwsAppMesh) error {

	// Region must be valid
	if appMesh.Region == "" {
		return errors.Errorf("region is required for AWS App Mesh")
	} else if !strings.Contains(AppMeshAvailableRegions, appMesh.Region) {
		return errors.Errorf("invalid AWS region [%s]. AWS App Mesh is currently available in: %s", appMesh.Region, AppMeshAvailableRegions)
	}

	// Check whether secret exists and can be used to access App Mesh
	if err := v.validateSecret(ctx, appMesh); err != nil {
		return err
	}

	// Validate auto-injection configuration only if it is enabled
	if appMesh.EnableAutoInject {
		if err := v.validateAutoInjectionConfig(ctx, appMesh); err != nil {
			return errors.Wrapf(err, "invalid auto-injection configuration")
		}
	}

	return nil
}

func (v *validator) validateSecret(ctx context.Context, appMesh *v1.AwsAppMesh) error {
	secretRef := appMesh.AwsSecret
	if secretRef == nil {
		return errors.Errorf("an AWS secret is required for supergloo to be able to access AWS App Mesh")
	}
	secret, err := v.secretClient.Read(secretRef.Namespace, secretRef.Name, clients.ReadOpts{Ctx: ctx})
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve secret %s.%s", secretRef.Namespace, secretRef.Name)
	}
	awsSecret := secret.GetAws()
	if awsSecret == nil {
		return errors.Errorf("expected secret of kind Secret_Aws, but found %s", secret.GetKind())
	}

	if err := verifyCredentials(awsSecret, appMesh.Region); err != nil {
		return errors.Wrapf(err, "failed to verify AWS credentials stored in secret %s against App Mesh in region %s",
			secret.Metadata.String(), appMesh.Region)
	}
	return nil
}

func (v *validator) validateAutoInjectionConfig(ctx context.Context, appMesh *v1.AwsAppMesh) error {

	// Validate selector for pod sidecar injection
	podSelector := appMesh.InjectionSelector
	if podSelector == nil {
		return errors.Errorf("InjectionSelector is required when EnableAutoInject==true")
	}
	if podSelector.GetUpstreamSelector() != nil {
		return errors.Errorf("upstream injection selectors are currently not supported")
	}

	// If this is nil, the webhook will default to our standard map: <supergloo_namespace>.sidecar-injector
	if configMapRef := appMesh.SidecarPatchConfigMap; configMapRef != nil {
		_, err := v.kube.CoreV1().ConfigMaps(configMapRef.Namespace).Get(configMapRef.Name, metav1.GetOptions{})
		if err != nil {
			return errors.Wrapf(err, "failed to find SidecarPatchConfigMap")
		}
	}

	if vnLabel := appMesh.VirtualNodeLabel; vnLabel == "" {
		return errors.Errorf("VirtualNodeLabel is required when EnableAutoInject==true")
	} else if errs := validation.IsQualifiedName(vnLabel); len(errs) > 0 {
		return errors.Errorf("invalid VirtualNodeLabel format: %s", errs)
	}

	return nil
}

// TODO: deduplicate from cli and move to client set
func verifyCredentials(secret *gloov1.AwsSecret, region string) error {

	appmeshClient, err := cliclients.NewAppmeshClient(secret.AccessKey, secret.SecretKey, region)
	if err != nil {
		return err
	}

	// Check if we can connect to appmesh with the provided credentials
	if _, err := appmeshClient.ListMeshes(nil); err != nil {
		return errors.Wrapf(err, "unable to access the AWS App Mesh service using the provided credentials. "+
			"Make sure they are associated with an IAM user that has permissions to access and modify AWS App "+
			"Mesh resources. Underlying error is")
	}
	return nil
}
