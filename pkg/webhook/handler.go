package webhook

import (
	"context"
	"fmt"
	"strings"

	"github.com/solo-io/go-utils/errors"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func admit(ctx context.Context, ar v1beta1.AdmissionReview) (*v1beta1.AdmissionResponse, error) {
	logger := contextutils.LoggerFrom(ctx)

	podResource := metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"}
	if ar.Request.Resource != podResource {
		return nil, errors.Errorf("expected resource to be: %s, but found: %v", podResource, ar.Request.Resource)
	}

	pod := corev1.Pod{}
	if _, _, err := Codecs.UniversalDeserializer().Decode(ar.Request.Object.Raw, nil, &pod); err != nil {
		return nil, errors.Wrapf(err, "failed to deserialize raw pod resource")
	}
	logger.Infof("evaluating pod %s.%s for sidecar auto-injection", pod.Namespace, pod.Name)

	// Retrieve all AWS App Mesh resources with EnableAutoInject == true
	awsAppMeshes, err := getAppMeshesWithAutoInjection(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list meshes")
	}

	// Check whether the pod has to be injected with a sidecar
	var matchingMeshes []*v1.Mesh
	for _, mesh := range awsAppMeshes {
		// this is safe, the mesh type has been validated in getAppMeshesWithAutoInjection
		selector := mesh.MeshType.(*v1.Mesh_AwsAppMesh).AwsAppMesh.InjectionSelector
		if selector == nil {
			return nil, errors.Errorf("auto-injection enabled but no selector for mesh %s.%s", mesh.Metadata.Namespace, mesh.Metadata.Name)
		}
		logger.Debugf("testing pod against pod selector %v in mesh %s.%s", selector, mesh.Metadata.Namespace, mesh.Metadata.Name)
		podMatchesSelector, err := match(pod, selector)
		if err != nil {
			return nil, err
		}

		if podMatchesSelector {
			logger.Infof("pod %s.%s matches selector %v in mesh %s.%s",
				pod.Namespace, pod.Name, selector, mesh.Metadata.Namespace, mesh.Metadata.Name)
			matchingMeshes = append(matchingMeshes, mesh)
		}
	}

	reviewResponse := v1beta1.AdmissionResponse{}
	reviewResponse.Allowed = true

	switch len(matchingMeshes) {
	case 0:
		logger.Info("pod does not match any mesh auto-injection selector, admit it without patching")
		break
	case 1:
		matchingMesh := matchingMeshes[0]

		// Get the config map containing the sidecar patch to be applied to the pod
		patchConfigMapRef := matchingMesh.MeshType.(*v1.Mesh_AwsAppMesh).AwsAppMesh.SidecarPatchConfigMap
		if patchConfigMapRef == nil {
			return nil, errors.Errorf("auto-injection enabled SidecarPatchConfigMap is nil for mesh %s.%s",
				matchingMesh.Metadata.Namespace, matchingMesh.Metadata.Name)
		}
		configMap, err := globalClientSet.GetConfigMap(patchConfigMapRef.Namespace, patchConfigMapRef.Name)
		if err != nil {
			return nil, errors.Errorf("failed to retrieve config map [%s]", patchConfigMapRef.String())
		}
		logger.Debugf("found config map [%s]", patchConfigMapRef.String())

		logger.Infof("injecting AWS App Mesh Envoy sidecar")
		podJSONPatch, err := buildSidecarPatch(&pod, configMap, matchingMesh)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to inject sidecar")
		}

		pt := v1beta1.PatchTypeJSONPatch
		reviewResponse.PatchType = &pt
		reviewResponse.Patch = podJSONPatch
		reviewResponse.Result = &metav1.Status{
			Status:  "Success",
			Message: strings.TrimSpace("successfully injected pod with AWS App Mesh Envoy sidecar")}

	default:
		toErrorResponse(ctx, "pod matches selectors in multiple meshes. Multiple injection is currently not supported")
	}

	return &reviewResponse, nil
}

// Returns true if the given pod matches the given selector
func match(pod corev1.Pod, selector *v1.PodSelector) (bool, error) {
	switch s := selector.SelectorType.(type) {
	case *v1.PodSelector_LabelSelector_:
		podLabels := pod.Labels
		labelsMatch := labels.SelectorFromSet(s.LabelSelector.LabelsToMatch).Matches(labels.Set(podLabels))
		return labelsMatch, nil

	case *v1.PodSelector_UpstreamSelector_:
		return false, fmt.Errorf("upstream selectors are currently unsupported by pod auto-injection")

	case *v1.PodSelector_NamespaceSelector_:
		for _, ns := range s.NamespaceSelector.Namespaces {
			if ns == pod.Namespace {
				return true, nil
			}
		}
	}
	return false, nil
}

func getAppMeshesWithAutoInjection(ctx context.Context) ([]*v1.Mesh, error) {
	var awsAppMeshes []*v1.Mesh
	meshes, err := globalClientSet.ListMeshes(metav1.NamespaceAll, clients.ListOpts{Ctx: ctx})
	if err != nil {
		return nil, err
	}

	for _, mesh := range meshes {
		appMesh, isAppMesh := mesh.MeshType.(*v1.Mesh_AwsAppMesh)
		if !isAppMesh {
			continue
		}
		if appMesh.AwsAppMesh == nil {
			contextutils.LoggerFrom(ctx).Warnf("unexpected nil value for AwsAppMesh in mesh %s.%s", mesh.Metadata.Namespace, mesh.Metadata.Name)
			continue
		}

		if !appMesh.AwsAppMesh.EnableAutoInject {
			continue
		}
		awsAppMeshes = append(awsAppMeshes, mesh)
	}
	return awsAppMeshes, nil
}
