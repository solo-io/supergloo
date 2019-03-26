package test

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var AppMeshInjectEnabled = &v1.Mesh{
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

var AppMeshInjectDisabled = &v1.Mesh{
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

var AppMeshNoConfigMap = &v1.Mesh{
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

var AppMeshNoSelector = &v1.Mesh{
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

var IstioMesh = &v1.Mesh{
	Metadata: core.Metadata{
		Name:      "test-mesh",
		Namespace: "supergloo-system",
	},
	MeshType: &v1.Mesh_Istio{
		Istio: &v1.IstioMesh{
			InstallationNamespace: "supergloo-system",
		}}}

var MatchingPod = `
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

var NonMatchingPod = `
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

var ConfigMap = `
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

var NoContainerPatch = `
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

var NoInitContainerPatch = `
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

var EmptyPatch = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: sidecar-injector-webhook-configmap
data:
`

var TwoEntryPatch = `
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
