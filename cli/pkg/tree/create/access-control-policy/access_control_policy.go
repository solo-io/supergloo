package access_control_policy

import (
	"context"
	"fmt"
	"io"
	"strconv"

	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/interactive"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/prompts"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/validate"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	k8s_meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	MatcherSelectorOptionName = "matcher"
	RefSelectorOptionName     = "service account reference"
)

var (
	AllowedMethods = []string{
		core_types.HttpMethodValue_GET.String(),
		core_types.HttpMethodValue_PUT.String(),
		core_types.HttpMethodValue_POST.String(),
		core_types.HttpMethodValue_DELETE.String(),
		core_types.HttpMethodValue_HEAD.String(),
		core_types.HttpMethodValue_CONNECT.String(),
		core_types.HttpMethodValue_OPTIONS.String(),
		core_types.HttpMethodValue_PATCH.String(),
		core_types.HttpMethodValue_TRACE.String(),
	}
	IgnoredNamespaces = []string{"istio-operator", "istio-system", "kube-system", "kube-node-lease", "kube-public", "local-path-storage"}
)

type CreateAccessControlPolicyCmd *cobra.Command

func CreateAccessControlPolicyCommand(
	ctx context.Context,
	out io.Writer,
	opts *options.Options,
	kubeLoader common_config.KubeLoader,
	kubeClientsFactory common.KubeClientsFactory,
	interactivePrompt interactive.InteractivePrompt,
	resourcePrinter resource_printing.ResourcePrinter,
) CreateAccessControlPolicyCmd {
	cmd := cliconstants.CreateAccessControlPolicyCommand
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return createAccessControlPolicy(ctx, out, kubeLoader, kubeClientsFactory, opts, interactivePrompt, resourcePrinter)
	}
	return &cmd
}

func createAccessControlPolicy(
	ctx context.Context,
	out io.Writer,
	kubeLoader common_config.KubeLoader,
	kubeClientsFactory common.KubeClientsFactory,
	opts *options.Options,
	prompt interactive.InteractivePrompt,
	resourcePrinter resource_printing.ResourcePrinter,
) error {
	var err error
	var masterCfg *rest.Config
	var masterKubeClients *common.KubeClients
	var sourceSelector *core_types.IdentitySelector
	var targetSelector *core_types.ServiceSelector
	var allowedPaths []string
	var allowedMethods []core_types.HttpMethodValue
	var allowedPorts []uint32
	if masterCfg, err = kubeLoader.GetRestConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext); err != nil {
		return err
	}
	if masterKubeClients, err = kubeClientsFactory(masterCfg, opts.Root.WriteNamespace); err != nil {
		return err
	}
	if sourceSelector, err = selectSourcesInteractively(ctx, masterKubeClients.ServiceAccountClient, prompt); err != nil {
		return err
	}
	if targetSelector, err = selectTargetsInteractively(ctx, masterKubeClients.MeshServiceClient, prompt); err != nil {
		return err
	}
	if allowedPaths, err = promptAllowedPathsInteractively(prompt); err != nil {
		return err
	}
	if allowedMethods, err = selectAllowedHttpMethodsInteractively(prompt); err != nil {
		return err
	}
	if allowedPorts, err = selectAllowedPortsInteractively(prompt); err != nil {
		return err
	}
	accessControlPolicy := &v1alpha1.AccessControlPolicy{
		TypeMeta: k8s_meta_v1.TypeMeta{
			Kind: "AccessControlPolicy",
		},
		Spec: types.AccessControlPolicySpec{
			SourceSelector:      sourceSelector,
			DestinationSelector: targetSelector,
			AllowedPaths:        allowedPaths,
			AllowedMethods:      allowedMethods,
			AllowedPorts:        allowedPorts,
		},
	}
	if !opts.Create.DryRun {
		return masterKubeClients.AccessControlPolicyClient.Create(ctx, accessControlPolicy)
	} else {
		return resourcePrinter.Print(out, accessControlPolicy, resource_printing.OutputFormat(opts.Create.OutputFormat))
	}
}

func selectSourcesInteractively(
	ctx context.Context,
	serviceAccountClient kubernetes_core.ServiceAccountClient,
	interactivePrompt interactive.InteractivePrompt,
) (*core_types.IdentitySelector, error) {
	var err error
	var identitySelectorType string
	var serviceAccountNames, clusters, namespaces []string
	var serviceAccountNamesToRefs map[string]*core_types.ResourceRef
	sourceSelector := &core_types.IdentitySelector{}
	identitySelectorTypes := []string{MatcherSelectorOptionName, RefSelectorOptionName}
	if identitySelectorType, err = interactivePrompt.SelectValue("Select identity selector type", identitySelectorTypes); err != nil {
		return nil, err
	}
	if identitySelectorType == MatcherSelectorOptionName {
		if namespaces, err = prompts.PromptCommaDelimitedValues(
			"Specify namespaces for selecting source workloads, omit to permit workloads of any namespace", interactivePrompt); err != nil {
			return nil, err
		}
		if clusters, err = prompts.PromptCommaDelimitedValues(
			"Specify clusters for selecting source workloads, omit to permit any cluster", interactivePrompt); err != nil {
			return nil, err
		}
		sourceSelector.IdentitySelectorType = &core_types.IdentitySelector_Matcher_{
			Matcher: &core_types.IdentitySelector_Matcher{
				Namespaces: namespaces,
				Clusters:   clusters,
			},
		}
	} else {
		var selections []string
		if serviceAccountNames, serviceAccountNamesToRefs, err = fetchServiceAccountRefs(ctx, serviceAccountClient); err != nil {
			return nil, err
		}
		if selections, err = interactivePrompt.SelectMultipleValues(
			"Specify service account references for selecting source workloads", serviceAccountNames); err != nil {
			return nil, err
		}
		var serviceAccountRefs []*core_types.ResourceRef
		for _, selection := range selections {
			serviceAccountRefs = append(serviceAccountRefs, serviceAccountNamesToRefs[selection])
		}
		sourceSelector.IdentitySelectorType = &core_types.IdentitySelector_ServiceAccountRefs_{
			ServiceAccountRefs: &core_types.IdentitySelector_ServiceAccountRefs{
				ServiceAccounts: serviceAccountRefs,
			},
		}
	}
	return sourceSelector, nil
}

func selectTargetsInteractively(
	ctx context.Context,
	meshServiceClient zephyr_discovery.MeshServiceClient,
	interactivePrompt interactive.InteractivePrompt,
) (*core_types.ServiceSelector, error) {
	meshServiceNames, meshServiceNamesToRefs, err := fetchMeshServiceRefs(ctx, meshServiceClient)
	if err != nil {
		return nil, err
	}
	targetSelector, err := prompts.SelectServiceSelector(
		"Select the destination Services, omit to select all", meshServiceNames, meshServiceNamesToRefs, interactivePrompt)
	if err != nil {
		return nil, err
	}
	return targetSelector, nil
}

func promptAllowedPathsInteractively(prompt interactive.InteractivePrompt) ([]string, error) {
	var allowedPaths []string
	for {
		value, err := prompt.PromptValue("Specify allowed path, omit to continue.", "")
		if err != nil {
			return nil, err
		}
		if value == "" {
			break
		}
		allowedPaths = append(allowedPaths, value)
	}
	return allowedPaths, nil
}

func selectAllowedHttpMethodsInteractively(prompt interactive.InteractivePrompt) ([]core_types.HttpMethodValue, error) {
	var httpMethods []core_types.HttpMethodValue
	selections, err := prompt.SelectMultipleValues("Select allowed HTTP methods", AllowedMethods)
	if err != nil {
		return nil, err
	}
	for _, selection := range selections {
		value, ok := core_types.HttpMethodValue_value[selection]
		if !ok {
			return nil, eris.Errorf("Unrecognized HTTP enum value: %s", selection)
		}
		httpMethods = append(httpMethods, core_types.HttpMethodValue(value))
	}
	return httpMethods, nil
}

func selectAllowedPortsInteractively(prompt interactive.InteractivePrompt) ([]uint32, error) {
	var ports []uint32
	for {
		value, err := prompt.PromptValueWithValidator(
			"Specify allowed port", "", validate.AllowEmpty(validate.PositiveInteger))
		if err != nil {
			return nil, err
		}
		if value == "" {
			break
		}
		port, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		ports = append(ports, uint32(port))
	}
	return ports, nil
}

// TODO fetch service accounts from all clusters, currently only fetches for cluster hosting mgmt plane
func fetchServiceAccountRefs(
	ctx context.Context,
	serviceAccountClient kubernetes_core.ServiceAccountClient,
) ([]string,
	map[string]*core_types.ResourceRef,
	error) {
	// Ignore irrelevant namespaces
	var ignoredNamespacesSelectors []fields.Selector
	for _, ns := range IgnoredNamespaces {
		ignoredNamespacesSelectors = append(ignoredNamespacesSelectors, fields.OneTermNotEqualSelector("metadata.namespace", ns))
	}
	serviceAccounts, err := serviceAccountClient.List(ctx, &client.ListOptions{
		FieldSelector: fields.AndSelectors(ignoredNamespacesSelectors...),
	})
	if err != nil {
		return nil, nil, err
	}
	var serviceAccountNames []string
	serviceAccountNamesToRef := map[string]*core_types.ResourceRef{}
	for _, serviceAccount := range serviceAccounts.Items {
		serviceAccount := serviceAccount
		serviceAccountName := buildServiceAccountName(&serviceAccount)
		serviceAccountNames = append(serviceAccountNames, serviceAccountName)
		serviceAccountNamesToRef[serviceAccountName] = &core_types.ResourceRef{
			Name:      serviceAccount.GetName(),
			Namespace: serviceAccount.GetNamespace(),
			Cluster:   serviceAccount.GetClusterName(),
		}
	}
	return serviceAccountNames, serviceAccountNamesToRef, nil
}

func buildServiceAccountName(serviceAccount *v1.ServiceAccount) string {
	return fmt.Sprintf("%s.%s.%s", serviceAccount.GetName(), serviceAccount.GetNamespace(), serviceAccount.GetClusterName())
}

func fetchMeshServiceRefs(
	ctx context.Context,
	meshServiceClient zephyr_discovery.MeshServiceClient,
) ([]string,
	map[string]*core_types.ResourceRef,
	error) {
	meshServices, err := meshServiceClient.List(ctx)
	if err != nil {
		return nil, nil, err
	}
	var meshServiceNames []string
	meshServiceNamesToRef := map[string]*core_types.ResourceRef{}
	for _, meshService := range meshServices.Items {
		meshService := meshService
		meshServiceName := buildMeshServiceName(&meshService)
		meshServiceNames = append(meshServiceNames, meshServiceName)
		meshServiceNamesToRef[meshServiceName] = &core_types.ResourceRef{
			Name:      meshService.GetName(),
			Namespace: meshService.GetNamespace(),
			Cluster:   meshService.GetClusterName(),
		}
	}
	return meshServiceNames, meshServiceNamesToRef, nil
}

func buildMeshServiceName(meshService *discovery_v1alpha1.MeshService) string {
	return fmt.Sprintf("%s.%s.%s", meshService.GetName(), meshService.GetNamespace(), meshService.Spec.GetKubeService().GetRef().GetCluster())
}
