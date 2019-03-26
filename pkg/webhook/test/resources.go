package test

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	. "github.com/onsi/gomega"
)

type patchConfigMap struct {
	AsStruct *corev1.ConfigMap
	AsString string
}

type pod struct {
	AsStruct *corev1.Pod
	AsString string
}

func (p *pod) ToRequest() admissionv1beta1.AdmissionReview {
	return admissionv1beta1.AdmissionReview{
		Request: &admissionv1beta1.AdmissionRequest{
			Resource: metav1.GroupVersionResource{Group: "", Version: "v1", Resource: "pods"},
			Object: runtime.RawExtension{
				Raw: []byte(p.AsString),
			},
		},
	}
}

type ResourcesForTest struct {
	OneContOneInitContPatch patchConfigMap
	NoContainerPatch        patchConfigMap
	NoInitContainerPatch    patchConfigMap
	EmptyPatch              patchConfigMap
	TwoEntryPatch           patchConfigMap
	MatchingPod             pod
	NonMatchingPod          pod
	AppMeshInjectEnabled    *v1.Mesh
	AppMeshInjectDisabled   *v1.Mesh
	AppMeshNoConfigMap      *v1.Mesh
	AppMeshNoSelector       *v1.Mesh
	IstioMesh               *v1.Mesh
}

func newPatchConfigMap(decoder runtime.Decoder, configMap string) patchConfigMap {
	configMapStruct := &corev1.ConfigMap{}
	_, _, err := decoder.Decode([]byte(configMap), nil, configMapStruct)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return patchConfigMap{
		AsString: configMap,
		AsStruct: configMapStruct,
	}
}

func newPod(decoder runtime.Decoder, podString string) pod {
	podStruct := &corev1.Pod{}
	_, _, err := decoder.Decode([]byte(podString), nil, podStruct)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return pod{
		AsString: podString,
		AsStruct: podStruct,
	}
}

func GetTestResources(decoder runtime.Decoder) *ResourcesForTest {
	return &ResourcesForTest{
		OneContOneInitContPatch: newPatchConfigMap(decoder, oneContainerOneInitContainerPatch),
		NoContainerPatch:        newPatchConfigMap(decoder, noContainerPatch),
		NoInitContainerPatch:    newPatchConfigMap(decoder, noInitContainerPatch),
		EmptyPatch:              newPatchConfigMap(decoder, emptyPatch),
		TwoEntryPatch:           newPatchConfigMap(decoder, twoEntryPatch),
		MatchingPod:             newPod(decoder, matchingPod),
		NonMatchingPod:          newPod(decoder, nonMatchingPod),
		AppMeshInjectEnabled:    appMeshInjectEnabled,
		AppMeshInjectDisabled:   appMeshInjectDisabled,
		AppMeshNoConfigMap:      appMeshNoConfigMap,
		AppMeshNoSelector:       appMeshNoSelector,
		IstioMesh:               istioMesh,
	}
}

var appMeshInjectEnabled = &v1.Mesh{
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
