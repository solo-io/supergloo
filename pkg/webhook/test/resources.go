package test

import (
	"github.com/ghodss/yaml"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	. "github.com/onsi/gomega"
)

type patchConfigMap struct {
	AsStruct     *corev1.ConfigMap
	AsJsonString string
}

type pod struct {
	AsStruct     *corev1.Pod
	AsJsonString string
}

func (p *pod) ToRequest() admissionv1beta1.AdmissionReview {
	return admissionv1beta1.AdmissionReview{
		Request: &admissionv1beta1.AdmissionRequest{
			Resource: metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			Object: runtime.RawExtension{
				Raw: []byte(p.AsJsonString),
			},
		},
	}
}

type ResourcesForTest struct {
	OneContOneInitContPatch               patchConfigMap
	NoContainerPatch                      patchConfigMap
	NoInitContainerPatch                  patchConfigMap
	EmptyPatch                            patchConfigMap
	TwoEntryPatch                         patchConfigMap
	MatchingPod                           pod
	MatchingPodWithoutPorts               pod
	NonMatchingPod                        pod
	AppMeshInjectEnabledLabelSelector     *v1.Mesh
	AppMeshInjectEnabledNamespaceSelector *v1.Mesh
	AppMeshInjectDisabled                 *v1.Mesh
	AppMeshNoConfigMap                    *v1.Mesh
	AppMeshNoSelector                     *v1.Mesh
	IstioMesh                             *v1.Mesh
	TemplateData                          interface{}
}

func newPatchConfigMap(decoder runtime.Decoder, configMapYamlString string) patchConfigMap {
	configMapStruct := &corev1.ConfigMap{}

	jsonBytes, err := yaml.YAMLToJSON([]byte(configMapYamlString))
	if err != nil {
		panic("failed to convert configMap YAML to JSON")
	}

	_, _, err = decoder.Decode([]byte(configMapYamlString), nil, configMapStruct)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	return patchConfigMap{
		AsJsonString: string(jsonBytes),
		AsStruct:     configMapStruct,
	}
}

func newPod(decoder runtime.Decoder, podYamlString string) pod {
	podStruct := &corev1.Pod{}

	jsonBytes, err := yaml.YAMLToJSON([]byte(podYamlString))
	if err != nil {
		panic("failed to convert pod YAML to JSON")
	}

	_, _, err = decoder.Decode([]byte(podYamlString), nil, podStruct)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())

	return pod{
		AsJsonString: string(jsonBytes),
		AsStruct:     podStruct,
	}
}

func GetTestResources(decoder runtime.Decoder) *ResourcesForTest {
	return &ResourcesForTest{
		OneContOneInitContPatch:               newPatchConfigMap(decoder, oneContainerOneInitContainerPatch),
		NoContainerPatch:                      newPatchConfigMap(decoder, noContainerPatch),
		NoInitContainerPatch:                  newPatchConfigMap(decoder, noInitContainerPatch),
		EmptyPatch:                            newPatchConfigMap(decoder, emptyPatch),
		TwoEntryPatch:                         newPatchConfigMap(decoder, twoEntryPatch),
		MatchingPod:                           newPod(decoder, matchingPod),
		MatchingPodWithoutPorts:               newPod(decoder, matchingPodWithoutPorts),
		NonMatchingPod:                        newPod(decoder, nonMatchingPod),
		AppMeshInjectEnabledLabelSelector:     appMeshInjectEnabledLabelSelector,
		AppMeshInjectEnabledNamespaceSelector: appMeshInjectEnabledNamespaceSelector,
		AppMeshInjectDisabled:                 appMeshInjectDisabled,
		AppMeshNoConfigMap:                    appMeshNoConfigMap,
		AppMeshNoSelector:                     appMeshNoSelector,
		IstioMesh:                             istioMesh,
		TemplateData:                          getTemplateData(),
	}
}

func getTemplateData() interface{} {
	return struct {
		MeshName        string
		VirtualNodeName string
		AppPort         int32
		AwsRegion       string
	}{
		"test-mesh", "testrunner-vn", 1234, "us-east-1",
	}
}

var appMeshInjectEnabledLabelSelector = &v1.Mesh{
	Metadata: core.Metadata{
		Name:      "test-mesh",
		Namespace: "supergloo-system",
	},
	MeshType: &v1.Mesh_AwsAppMesh{
		AwsAppMesh: &v1.AwsAppMesh{
			Region:           "us-east-1",
			VirtualNodeLabel: "virtual-node",
			EnableAutoInject: true,
			SidecarPatchConfigMap: &core.ResourceRef{
				Name:      "sidecar-injector-webhook-configmap",
				Namespace: "supergloo-system",
			},
			InjectionSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "testrunner"}}}}}}}

var appMeshInjectEnabledNamespaceSelector = &v1.Mesh{
	Metadata: core.Metadata{
		Name:      "test-mesh",
		Namespace: "supergloo-system",
	},
	MeshType: &v1.Mesh_AwsAppMesh{
		AwsAppMesh: &v1.AwsAppMesh{
			Region:           "us-east-1",
			VirtualNodeLabel: "virtual-node",
			EnableAutoInject: true,
			SidecarPatchConfigMap: &core.ResourceRef{
				Name:      "sidecar-injector-webhook-configmap",
				Namespace: "supergloo-system",
			},
			InjectionSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_NamespaceSelector_{
					NamespaceSelector: &v1.PodSelector_NamespaceSelector{
						Namespaces: []string{"my-ns"}}}}}}}

var appMeshInjectDisabled = &v1.Mesh{
	Metadata: core.Metadata{
		Name:      "test-mesh",
		Namespace: "supergloo-system",
	},
	MeshType: &v1.Mesh_AwsAppMesh{
		AwsAppMesh: &v1.AwsAppMesh{
			Region:           "us-east-1",
			VirtualNodeLabel: "virtual-node",
			EnableAutoInject: false,
			SidecarPatchConfigMap: &core.ResourceRef{
				Name:      "sidecar-injector-webhook-configmap",
				Namespace: "supergloo-system",
			},
			InjectionSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "testrunner"}}}}}}}

var appMeshNoConfigMap = &v1.Mesh{
	Metadata: core.Metadata{
		Name:      "test-mesh",
		Namespace: "supergloo-system",
	},
	MeshType: &v1.Mesh_AwsAppMesh{
		AwsAppMesh: &v1.AwsAppMesh{
			Region:           "us-east-1",
			VirtualNodeLabel: "virtual-node",
			EnableAutoInject: true,
			InjectionSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{
							"app": "testrunner"}}}}}}}

var appMeshNoSelector = &v1.Mesh{
	Metadata: core.Metadata{
		Name:      "test-mesh",
		Namespace: "supergloo-system",
	},
	MeshType: &v1.Mesh_AwsAppMesh{
		AwsAppMesh: &v1.AwsAppMesh{
			Region:           "us-east-1",
			VirtualNodeLabel: "virtual-node",
			EnableAutoInject: true,
			SidecarPatchConfigMap: &core.ResourceRef{
				Name:      "sidecar-injector-webhook-configmap",
				Namespace: "supergloo-system"}}}}

var istioMesh = &v1.Mesh{
	Metadata: core.Metadata{
		Name:      "test-mesh",
		Namespace: "supergloo-system",
	},
	MeshType: &v1.Mesh_Istio{
		Istio: &v1.IstioMesh{
			InstallationNamespace: "supergloo-system",
		}}}

var matchingPod = `
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
  namespace: my-ns
  labels:
    app: testrunner
    version: "1"
    virtual-node: testrunner-vn
spec:
  containers:
    - image: soloio/testrunner:latest
      imagePullPolicy: IfNotPresent
      name: testrunner
      ports:
        - containerPort: 1234
`
var matchingPodWithoutPorts = `
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: testrunner
    version: "1"
    virtual-node: testrunner-vn
spec:
  containers:
    - image: soloio/testrunner:latest
      imagePullPolicy: IfNotPresent
      name: testrunner
`

var nonMatchingPod = `
apiVersion: v1
kind: Pod
metadata:
spec:
  containers:
    - image: soloio/testrunner:latest
      imagePullPolicy: IfNotPresent
      name: testrunner
      ports:
        - containerPort: 1234
`

var oneContainerOneInitContainerPatch = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: sidecar-injector-webhook-configmap
data:
  aws-app-mesh-patch.yaml: |
    containers:
      - name: envoy
        image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
        securityContext:
          runAsUser: 1337
        env:
          - name: "APPMESH_VIRTUAL_NODE_NAME"
            value: "mesh/{{- .MeshName -}}/virtualNode/{{- .VirtualNodeName -}}"
          - name: "ENVOY_LOG_LEVEL"
            value: "debug"
          - name: "AWS_REGION"
            value: "{{- .AwsRegion -}}"
    initContainers:
      - name: proxyinit
        image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
        securityContext:
          capabilities:
            add:
              - NET_ADMIN
        env:
          - name: "APPMESH_START_ENABLED"
            value: "1"
          - name: "APPMESH_IGNORE_UID"
            value: "1337"
          - name: "APPMESH_ENVOY_INGRESS_PORT"
            value: "15000"
          - name: "APPMESH_ENVOY_EGRESS_PORT"
            value: "15001"
          - name: "APPMESH_APP_PORTS"
            value: "{{- .AppPort -}}"
          - name: "APPMESH_EGRESS_IGNORED_IP"
            value: "169.254.169.254"
`

var noContainerPatch = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: sidecar-injector-webhook-configmap
data:
  aws-app-mesh-patch.yaml: |
    initContainers:
      - name: proxyinit
        image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
`

var noInitContainerPatch = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: sidecar-injector-webhook-configmap
data:
  aws-app-mesh-patch.yaml: |
    containers:
      - name: envoy
        image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
`

var emptyPatch = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: sidecar-injector-webhook-configmap
data:
`

var twoEntryPatch = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: sidecar-injector-webhook-configmap
data:
  aws-app-mesh-patch.yaml: |
    containers:
      - name: envoy
        image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
  some-other-patch.yaml: |
    containers:
      - name: hello
        image: some-repo/some-image:some-tag
`
