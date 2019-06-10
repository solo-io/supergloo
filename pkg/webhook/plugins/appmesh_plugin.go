package plugins

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/supergloo/pkg/constants"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	"github.com/solo-io/supergloo/pkg/registration/appmesh"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/webhook/clients"
	"github.com/solo-io/supergloo/pkg/webhook/patch"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

// This struct represents the data required for rendering the App Mesh patch template
type templateData struct {
	MeshName        string
	VirtualNodeName string
	AppPort         string
	AwsRegion       string
}

type AppMeshInjectionPlugin struct{}

func (AppMeshInjectionPlugin) Name() string {
	return "AppMeshInjectionPlugin"
}

func (AppMeshInjectionPlugin) GetAutoInjectMeshes(ctx context.Context) ([]*v1.Mesh, error) {
	var awsAppMeshes []*v1.Mesh
	meshes, err := clients.GetClientSet().ListMeshes(metav1.NamespaceAll, skclients.ListOpts{Ctx: ctx})
	if err != nil {
		return nil, err
	}

	for _, mesh := range meshes {
		appMesh := mesh.GetAwsAppMesh()

		// Mesh is not app mesh
		if appMesh == nil {
			continue
		}

		if !appMesh.EnableAutoInject {
			continue
		}
		awsAppMeshes = append(awsAppMeshes, mesh)
	}
	return awsAppMeshes, nil
}

func (AppMeshInjectionPlugin) CheckMatch(ctx context.Context, candidatePod *corev1.Pod, meshes []*v1.Mesh) ([]*v1.Mesh, error) {
	logger := contextutils.LoggerFrom(ctx)

	// Check whether the pod has to be injected with a sidecar
	var matchingMeshes []*v1.Mesh
	for _, mesh := range meshes {

		appMesh := mesh.GetAwsAppMesh()

		if appMesh == nil {
			continue
		}

		selector := appMesh.InjectionSelector
		if selector == nil {
			return nil, errors.Errorf("auto-injection enabled but no selector for mesh %s.%s", mesh.Metadata.Namespace, mesh.Metadata.Name)
		}
		logger.Debugf("testing pod against pod selector %v in mesh %s.%s", selector, mesh.Metadata.Namespace, mesh.Metadata.Name)
		podMatchesSelector, err := match(candidatePod, selector)
		if err != nil {
			return nil, err
		}

		if podMatchesSelector {
			logger.Infof("pod %s.%s matches selector %v in mesh %s.%s",
				candidatePod.Namespace, candidatePod.Name, selector, mesh.Metadata.Namespace, mesh.Metadata.Name)
			matchingMeshes = append(matchingMeshes, mesh)
		}
	}
	return matchingMeshes, nil
}

func (AppMeshInjectionPlugin) GetSidecarPatch(ctx context.Context, pod *corev1.Pod, meshes []*v1.Mesh) ([]patch.JSONPatchOperation, error) {
	logger := contextutils.LoggerFrom(ctx)

	// We do not currently support multiple injection for AWS App Mesh
	var mesh *v1.Mesh
	switch len(meshes) {
	case 0:
		return nil, nil
	case 1:
		mesh = meshes[0]
	default:
		return nil, errors.Errorf("pod matches selectors in multiple meshes. Multiple injection is currently not supported")
	}

	appMesh := mesh.GetAwsAppMesh()

	// Skip if not AWS App Mesh
	if appMesh == nil {
		return nil, nil
	}

	// Get the config map containing the sidecar patch to be applied to the pod
	// If nil, use the default config map: <supergloo_ns>.sidecar-injector
	patchConfigMapRef := appMesh.SidecarPatchConfigMap
	if patchConfigMapRef == nil {
		namespace, err := clients.GetClientSet().GetSuperglooNamespace()
		if err != nil {
			return nil, err
		}
		patchConfigMapRef = &core.ResourceRef{
			Namespace: namespace,
			Name:      constants.SidecarInjectorImageName,
		}
	}
	configMap, err := clients.GetClientSet().GetConfigMap(patchConfigMapRef.Namespace, patchConfigMapRef.Name)
	if err != nil {
		return nil, errors.Errorf("failed to retrieve config map [%s]", patchConfigMapRef.String())
	}
	logger.Debugf("found config map [%s]", patchConfigMapRef.String())

	// Get data to use during template rendering from the candidate pod
	templateData, err := getTemplateData(pod, mesh)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve data for patch template rendering")
	}

	podJSONPatch, err := patch.BuildSidecarPatch(pod, configMap, templateData)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to inject sidecar")
	}
	return podJSONPatch, nil
}

// Returns true if the given pod matches the given selector
func match(pod *corev1.Pod, selector *v1.PodSelector) (bool, error) {
	switch s := selector.SelectorType.(type) {
	case *v1.PodSelector_LabelSelector_:
		podLabels := pod.Labels
		labelsMatch := labels.SelectorFromSet(s.LabelSelector.LabelsToMatch).Matches(labels.Set(podLabels))
		return labelsMatch, nil

	case *v1.PodSelector_UpstreamSelector_:
		return false, fmt.Errorf("upstream selectors are currently unsupported by pod auto-injection")
	case *v1.PodSelector_ServiceSelector_:
		return false, fmt.Errorf("service selectors are currently unsupported by pod auto-injection")

	case *v1.PodSelector_NamespaceSelector_:
		for _, ns := range s.NamespaceSelector.Namespaces {
			if ns == pod.Namespace {
				return true, nil
			}
		}
	}
	return false, nil
}

func getTemplateData(pod *corev1.Pod, mesh *v1.Mesh) (templateData, error) {

	awsRegion := mesh.GetAwsAppMesh().Region
	if awsRegion == "" {
		return templateData{}, errors.Errorf("mesh resource is missing required Region field")
	}
	if !strings.Contains(appmesh.AppMeshAvailableRegions, awsRegion) {
		return templateData{}, errors.Errorf("AWS App Mesh is currently not available in [%s]. Supported regions are: %s", awsRegion, appmesh.AppMeshAvailableRegions)
	}

	vnLabel := mesh.GetAwsAppMesh().VirtualNodeLabel
	if vnLabel == "" {
		return templateData{}, errors.Errorf("mesh resource is missing required VirtualNodeLabel field")
	}
	virtualNodeName, ok := pod.Labels[vnLabel]
	if !ok {
		return templateData{}, errors.Errorf("pod is missing required virtual node label %v", vnLabel)
	}

	var ports []string
	for _, container := range pod.Spec.Containers {
		if len(container.Ports) == 0 {
			return templateData{}, errors.Errorf("no containerPorts for container %s. Must specify at least 1 port", container.Name)
		}
		for _, port := range container.Ports {
			ports = append(ports, fmt.Sprintf("%v", port.ContainerPort))
		}
	}

	return templateData{
		MeshName:        mesh.Metadata.Name,
		AppPort:         strings.Join(ports, ","),
		VirtualNodeName: virtualNodeName,
		AwsRegion:       awsRegion,
	}, nil
}
