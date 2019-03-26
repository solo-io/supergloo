package webhook

import (
	"bytes"
	"encoding/json"
	"html/template"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/errors"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
)

const appMeshSupportedRegions = "us-west-2, us-east-1, us-east-2, eu-west-1"

// A JSON Patch operation as defined in https://tools.ietf.org/html/rfc6902
type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// This struct represents the data used to execute the patch template
type templateData struct {
	MeshName        string
	VirtualNodeName string
	AppPort         int32
	AwsRegion       string
}

func buildSidecarPatch(pod *corev1.Pod, configMap *corev1.ConfigMap, mesh *v1.Mesh) ([]byte, error) {
	if len(configMap.Data) != 1 {
		return nil, errors.Errorf("expected exactly 1 entry in config map %s.%s but found %v", configMap.Namespace, configMap.Name, len(configMap.Data))
	}
	if len(pod.Spec.Containers) != 1 {
		return nil, errors.Errorf("expected exactly 1 container in pod %s.%s but found %v", pod.Namespace, pod.Name, len(pod.Spec.Containers))
	}
	if len(pod.Spec.Containers[0].Ports) != 1 {
		return nil, errors.Errorf("expected exactly 1 port in container %s but found %v", pod.Spec.Containers[0].Name, len(pod.Spec.Containers[0].Ports))
	}
	awsRegion := mesh.MeshType.(*v1.Mesh_AwsAppMesh).AwsAppMesh.Region
	if awsRegion == "" {
		return nil, errors.Errorf("mesh resource is missing required Region field")
	}
	if !strings.Contains(appMeshSupportedRegions, awsRegion) {
		return nil, errors.Errorf("AWS App Mesh is currently not available in [%s]. Supported regions are: %s", awsRegion, appMeshSupportedRegions)
	}

	vnLabel := mesh.MeshType.(*v1.Mesh_AwsAppMesh).AwsAppMesh.VirtualNodeLabel
	if vnLabel == "" {
		return nil, errors.Errorf("mesh resource is missing required VirtualNodeLabel field")
	}
	virtualNodeName, ok := pod.Labels[vnLabel]
	if !ok {
		return nil, errors.Errorf("pod is missing required virtual node label %v", vnLabel)
	}

	// Get the patch in the form of a pod spec
	var patch *corev1.PodSpec
	for _, patchTemplate := range configMap.Data {

		var err error
		patch, err = render(patchTemplate, templateData{
			MeshName:        mesh.Metadata.Name,
			AppPort:         pod.Spec.Containers[0].Ports[0].ContainerPort,
			VirtualNodeName: virtualNodeName,
			AwsRegion:       awsRegion,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "failed to render sidecar patch template")
		}

		// slice has only one element, but break just in case
		break
	}

	var patches []patchOperation
	patches = append(patches, toPatchOperation(len(pod.Spec.Containers) == 0, patch.Containers, "/spec/containers/-")...)
	patches = append(patches, toPatchOperation(len(pod.Spec.InitContainers) == 0, patch.InitContainers, "/spec/initContainers/-")...)

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		return nil, errors.Errorf("failed to marshal patches to JSON: %v", patches)
	}

	return patchBytes, nil
}

func toPatchOperation(initialArrayEmpty bool, patchContainers []corev1.Container, path string) (patches []patchOperation) {
	for i, container := range patchContainers {
		var value interface{} = container
		currPath := path

		// If the array is empty, the first patch operation path must create it by passing a slice as value
		// and a path that does not end with "/-"
		if i == 0 && initialArrayEmpty {
			currPath = strings.TrimSuffix(currPath, "/-")
			value = []corev1.Container{container}
		}

		patches = append(patches, patchOperation{
			Op:    "add",
			Path:  currPath,
			Value: value,
		})
	}
	return
}

func render(patchAsString string, data templateData) (*corev1.PodSpec, error) {
	patch := &corev1.PodSpec{}
	tmpl, err := template.New("patchTemplate").Parse(patchAsString)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse sidecar patch template")
	}

	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return nil, errors.Wrapf(err, "failed to execute sidecar patch template")
	}

	if err := yaml.Unmarshal([]byte(buf.String()), patch); err != nil {
		return nil, errors.Wrapf(err, "failed to deserialize sidecar patch to PodSpec")
	}
	return patch, nil
}
