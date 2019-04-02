package patch

import (
	"bytes"
	"html/template"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/solo-io/go-utils/errors"
	corev1 "k8s.io/api/core/v1"
)

// A JSON Patch operation as defined in https://tools.ietf.org/html/rfc6902
type JSONPatchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// This function expects the given config map to contain one entry under it's data element. The entry is expected to
// contain raw YAML representing a PodSpec. The YAML will be rendered as a template using the given data and used to
// build a slice of JSONPatch operations to be applied to the given pod.
//
// Currently only the 'containers' and 'initContainers' elements in the configMap are considered. The patch operations
// that are returned will append them to the correspondent elements in the target pod.
func BuildSidecarPatch(pod *corev1.Pod, configMap *corev1.ConfigMap, templateData interface{}) ([]JSONPatchOperation, error) {
	if len(configMap.Data) != 1 {
		return nil, errors.Errorf("expected exactly 1 entry in config map %s.%s but found %v", configMap.Namespace, configMap.Name, len(configMap.Data))
	}

	// Get the patch in the form of a pod spec
	var patch *corev1.PodSpec
	for _, patchTemplate := range configMap.Data {

		var err error
		patch, err = render(patchTemplate, templateData)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to render sidecar patch template")
		}

		// slice has only one element, but break just in case
		break
	}

	var patches []JSONPatchOperation
	patches = append(patches, toPatchOperation(len(pod.Spec.Containers) == 0, patch.Containers, "/spec/containers/-")...)
	patches = append(patches, toPatchOperation(len(pod.Spec.InitContainers) == 0, patch.InitContainers, "/spec/initContainers/-")...)

	return patches, nil
}

func render(patchAsString string, data interface{}) (*corev1.PodSpec, error) {
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

func toPatchOperation(initialArrayEmpty bool, patchContainers []corev1.Container, path string) (patches []JSONPatchOperation) {
	for i, container := range patchContainers {
		var value interface{} = container
		currPath := path

		// If the array is empty, the first patch operation path must create it by passing a slice as value
		// and a path that does not end with "/-"
		if i == 0 && initialArrayEmpty {
			currPath = strings.TrimSuffix(currPath, "/-")
			value = []corev1.Container{container}
		}

		patches = append(patches, JSONPatchOperation{
			Op:    "add",
			Path:  currPath,
			Value: value,
		})
	}
	return
}
