package surveyutils

import (
	"context"
	"strings"

	"github.com/solo-io/gloo/pkg/cliutil"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/pkg/registration/appmesh"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	namespaceSelector   = "Select all pods in a namespace"
	labelSelector       = "Select pods matching a set of labels"
	useDefaultConfigMap = "Use default"
	useCustomConfigMap  = "Provide my own"
)

// Survey to get AWS access credentials either from stdin or from a credential file.
func SurveyAppmeshRegistration(ctx context.Context, opts *options.Options) error {
	// Set secret with AWS credentials
	if err := surveySecret(ctx, opts); err != nil {
		return err
	}

	// Set AWS region
	if err := cliutil.ChooseFromList(
		"In which AWS region do you want AWS App Mes resources to be created?",
		&opts.RegisterAppMesh.Region,
		strings.Split(appmesh.AppMeshAvailableRegions, ","),
	); err != nil {
		return err
	}

	// Return if auto-injection is disabled
	if enableAutoInjection, err := surveyAutoInjection(); err != nil {
		return err
	} else if !enableAutoInjection {
		opts.RegisterAppMesh.EnableAutoInjection = "false"
		return nil
	}
	opts.RegisterAppMesh.EnableAutoInjection = "true"

	if err := surveyPodSelector(&opts.RegisterAppMesh.PodSelector); err != nil {
		return err
	}

	if err := surveyConfigMap(opts); err != nil {
		return err
	}

	var choice string
	if err := cliutil.GetStringInput("Select the key of the pod label that SuperGloo will use to assign your pods "+
		"to a VirtualNode", &choice); err != nil {
		return err
	}
	opts.RegisterAppMesh.VirtualNodeLabel = choice

	return nil
}

func surveySecret(ctx context.Context, opts *options.Options) error {
	secretClient := clients.MustSecretClient()
	secrets, err := secretClient.List(metav1.NamespaceAll, skclients.ListOpts{Ctx: ctx})
	if err != nil {
		return err
	}

	if len(secrets) == 0 {
		return errors.Errorf("could not find any AWS secret. Create one using 'supergloo create secret aws -i'")
	}

	var awsSecrets = v1.SecretList{}
	for _, s := range secrets {
		if s.GetAws() != nil {
			awsSecrets = append(awsSecrets, s)
		}
	}
	secret, err := surveyResources("secrets", "Choose a secret that contains the AWS credentials to "+
		"connect to AWS App Mesh (hint: you can easily create one with 'supergloo create secret aws -i')",
		"", awsSecrets.AsResources())
	if err != nil {
		return err
	}
	opts.RegisterAppMesh.Secret = options.ResourceRefValue(secret)
	return nil
}

func surveyAutoInjection() (bool, error) {
	var enableAutoInjection string
	if err := cliutil.ChooseFromList(
		"Do you want SuperGloo to auto-inject your pods with the AWS App Mesh sidecar proxies?",
		&enableAutoInjection,
		[]string{"yes", "no"},
	); err != nil {
		return false, err
	}

	if enableAutoInjection == "no" {
		return false, nil
	}
	return true, nil
}

// Upstream selector is currently not supported
func surveyPodSelector(in *options.Selector) error {
	var injectionSelectorType string
	if err := cliutil.ChooseFromList(
		"How you want SuperGloo to select which pods will be auto-injected?",
		&injectionSelectorType,
		[]string{namespaceSelector, labelSelector},
	); err != nil {
		return err
	}

	switch injectionSelectorType {
	case labelSelector:
		m := map[string]string(in.SelectedLabels)
		if err := SurveyMapStringString(&m); err != nil {
			return err
		}
		in.SelectedLabels = options.MapStringStringValue(m)
	case namespaceSelector:
		nss, err := SurveyNamespaces()
		if err != nil {
			return err
		}
		in.SelectedNamespaces = nss
	default:
		return errors.Errorf("%v is not a known selector type", injectionSelectorType)
	}

	return nil
}

func surveyConfigMap(opts *options.Options) error {
	var choice string
	if err := cliutil.ChooseFromList(
		"SuperGloo looks for the patch that will be applied to the pods matching the selection criteria in a config map. "+
			"Do you want to use the default one provided by supergloo or provide your own?",
		&choice,
		[]string{useDefaultConfigMap, useCustomConfigMap},
	); err != nil {
		return err
	}

	if choice == useDefaultConfigMap {
		// No need to set anything here, the registration event loop will generate the default map
		return nil
	}

	configMaps, err := clients.MustKubeClient().CoreV1().ConfigMaps(corev1.NamespaceAll).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(configMaps.Items) == 0 {
		return errors.Errorf("could not find any config map")
	}

	var resRefs []core.ResourceRef
	var resRefsKeys []string
	for _, cm := range configMaps.Items {
		rr := core.ResourceRef{
			Name:      cm.Name,
			Namespace: cm.Namespace,
		}
		resRefs = append(resRefs, rr)
		resRefsKeys = append(resRefsKeys, rr.Key())
	}

	var configMap string
	if err := cliutil.ChooseFromList("Select the config map that you would like to use", &configMap, resRefsKeys); err != nil {
		return err
	}

	var configMapToUse core.ResourceRef
	for _, ref := range resRefs {
		if ref.Key() == configMap {
			configMapToUse = ref
		}
	}

	opts.RegisterAppMesh.ConfigMap = options.ResourceRefValue(configMapToUse)
	return nil
}
