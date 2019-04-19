package appmesh

import (
	"strings"

	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/ghodss/yaml"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/protoutils"
	"github.com/solo-io/supergloo/api/custom/kubepod"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Initialize AppMesh configuration object", func() {

	Describe("getting pod info", func() {

		It("does return nil info if the pod does not have a sidecar", func() {
			info, err := getPodInfo("appmesh", getPod(noSidecar))
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(BeNil())
		})

		It("retrieves the correct info for a pod that belong to the given mesh", func() {
			info, err := getPodInfo("appmesh", getPod(withSidecar))
			Expect(err).NotTo(HaveOccurred())
			Expect(info).NotTo(BeNil())
			Expect(info.virtualNodeName).To(BeEquivalentTo("my-app-vn"))
			Expect(info.ports).To(HaveLen(2))
			Expect(info.ports).To(ConsistOf(uint32(8080), uint32(8081)))
		})

		It("fails if the APPMESH_VIRTUAL_NODE_NAME env has an incorrect format", func() {
			_, err := getPodInfo("appmesh", getPod(incorrectVirtualNodeEnvFormat))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unexpected format for APPMESH_VIRTUAL_NODE_NAME env for pod"))
		})

		It("fails if the APPMESH_APP_PORTS env has an incorrect format", func() {
			_, err := getPodInfo("appmesh", getPod(badPortsFormat))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to parse [8080,8081,] (value of APPMESH_APP_PORTS env) to int array"))
		})

		It("fails if the APPMESH_APP_PORTS env is missing", func() {
			_, err := getPodInfo("appmesh", getPod(noPortsEnv))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("could not find APPMESH_APP_PORTS env on any initContainer for pod"))
		})
	})

	It("correctly retrieves information for upstreams related to AppMesh pods", func() {
		podInfo, podList, err := getPodsForMesh("appmesh", getPodList(noSidecar, withSidecar))
		Expect(err).NotTo(HaveOccurred())

		usInfo, usList := getUpstreamsForMesh(getUpstreamList(upstreamsForInjectedPod, upstreamForOtherPod), podInfo, podList)
		Expect(usList).To(HaveLen(4))
		Expect(usInfo).To(HaveLen(4))
		Expect(podInfo).To(HaveLen(1))
		for _, info := range podInfo {
			Expect(info.upstreams).To(HaveLen(4))
		}
	})

	It("correctly groups pod info by virtual node name", func() {
		appMeshPods := make(awsAppMeshPodInfo, 0)
		addPod(appMeshPods, "1", "vn-1")
		addPod(appMeshPods, "2", "vn-2")
		addPod(appMeshPods, "3", "vn-3")
		addPod(appMeshPods, "4", "vn-1")
		addPod(appMeshPods, "5", "vn-3")

		groupedByVn := groupByVirtualNodeName(appMeshPods)

		Expect(groupedByVn).To(HaveLen(3))
		Expect(groupedByVn["vn-1"]).To(HaveLen(2))
		Expect(groupedByVn["vn-2"]).To(HaveLen(1))
		Expect(groupedByVn["vn-3"]).To(HaveLen(2))
	})

	Describe("initializing virtual nodes", func() {

		It("correctly creates a virtual node for the pod when all required information is available", func() {
			podInfo, podList, err := getPodsForMesh("appmesh", getPodList(noSidecar, withSidecar))
			Expect(err).NotTo(HaveOccurred())
			// this is needed since it sets the upstreams on the podInfo entries
			_, _ = getUpstreamsForMesh(getUpstreamList(upstreamsForInjectedPod, upstreamForOtherPod), podInfo, podList)

			virtualNodes, err := initializeVirtualNodes("appmesh", groupByVirtualNodeName(podInfo))
			Expect(err).NotTo(HaveOccurred())
			Expect(virtualNodes).To(HaveLen(1))
			var host string
			var virtualNode *appmesh.VirtualNodeData
			for h, vn := range virtualNodes {
				host = h
				virtualNode = vn
			}
			Expect(host).To(BeEquivalentTo("my-app.test-ns.svc.cluster.local"))
			Expect(*virtualNode.MeshName).To(BeEquivalentTo("appmesh"))
			Expect(*virtualNode.VirtualNodeName).To(BeEquivalentTo("my-app-vn"))
			Expect(*virtualNode.Spec.ServiceDiscovery.Dns.Hostname).To(BeEquivalentTo("my-app.test-ns.svc.cluster.local"))
			Expect(virtualNode.Spec.Listeners).To(HaveLen(2))
		})

		It("fails if we can't determine the DNS hostname for the virtual node because there are no correspondent upstreams", func() {
			podInfo, podList, err := getPodsForMesh("appmesh", getPodList(noSidecar, withSidecar))
			Expect(err).NotTo(HaveOccurred())
			// this is needed since it sets the upstreams on the podInfo entries
			_, _ = getUpstreamsForMesh(getUpstreamList(upstreamForOtherPod), podInfo, podList)

			_, err = initializeVirtualNodes("appmesh", groupByVirtualNodeName(podInfo))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("no services match its pods"))
		})

		It("fails if we can't determine the DNS hostname for the virtual node because there are multiple upstreams matching its pods", func() {
			podInfo, podList, err := getPodsForMesh("appmesh", getPodList(noSidecar, withSidecar))
			Expect(err).NotTo(HaveOccurred())
			// this is needed since it sets the upstreams on the podInfo entries
			_, _ = getUpstreamsForMesh(getUpstreamList(upstreamsForInjectedPod, upstreamForOtherPod, extraUpstreamForInjectedPod), podInfo, podList)

			_, err = initializeVirtualNodes("appmesh", groupByVirtualNodeName(podInfo))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("multiple services match its pods: " +
				"my-app.test-ns.svc.cluster.local,another-svc-matching-vn.test-ns.svc.cluster.local"))
		})

		It("fails if the upstreams matching its pods are missing a mapping for one of the ports specified via the APPMESH_APP_PORTS env ", func() {
			podInfo, podList, err := getPodsForMesh("appmesh", getPodList(noSidecar, withSidecarExtraPort))
			Expect(err).NotTo(HaveOccurred())
			// this is needed since it sets the upstreams on the podInfo entries
			_, _ = getUpstreamsForMesh(getUpstreamList(upstreamsForInjectedPod, upstreamForOtherPod), podInfo, podList)

			_, err = initializeVirtualNodes("appmesh", groupByVirtualNodeName(podInfo))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("does not define mapping for port 9999"))
		})

	})
})

func addPod(info awsAppMeshPodInfo, podName, vnName string) {
	pod := &v1.Pod{
		Pod: kubepod.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: podName},
		},
	}
	info[pod] = &podInfo{virtualNodeName: vnName}
}

func getPod(podYaml string) *corev1.Pod {
	var podObj corev1.Pod
	err := yaml.Unmarshal([]byte(podYaml), &podObj)
	if err != nil {
		panic(err) // should never happen
	}
	return &podObj
}

func getPodList(podYamls ...string) v1.PodList {
	var podList v1.PodList
	for _, yml := range podYamls {
		podList = append(podList, kubernetes.FromKube(getPod(yml)))
	}
	return podList
}

func getUpstreamList(upstreamYamls ...string) gloov1.UpstreamList {
	var upstreamList gloov1.UpstreamList
	for _, yml := range upstreamYamls {
		for _, v := range strings.Split(yml, "---") {
			var us gloov1.Upstream
			err := protoutils.UnmarshalYaml([]byte(v), &us)
			if err != nil {
				panic(err) // should never happen
			}
			upstreamList = append(upstreamList, &us)
		}
	}
	return upstreamList
}

const noSidecar = `
apiVersion: v1
kind: Pod
metadata:
  name: no-sidecar
  namespace: test-ns
  labels:
    app: no-sidecar
spec:
  containers:
    - image: busybox
      imagePullPolicy: IfNotPresent
      name: no-sidecar
      ports:
        - containerPort: 1234
`

const withSidecar = `
apiVersion: v1
kind: Pod
metadata:
  name: my-app
  namespace: test-ns
  labels:
    app: my-app
    version: v1
spec:
  containers:
    - image: busybox
      imagePullPolicy: IfNotPresent
      name: my-app
      ports:
        - containerPort: 8080
    - image: busybox
      imagePullPolicy: IfNotPresent
      name: my-other-app
      ports:
        - containerPort: 8081
    - env:
        - name: APPMESH_VIRTUAL_NODE_NAME
          value: mesh/appmesh/virtualNode/my-app-vn
        - name: ENVOY_LOG_LEVEL
          value: debug
        - name: AWS_REGION
          value: us-east-1
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imagePullPolicy: IfNotPresent
      name: envoy
      securityContext:
        procMount: Default
        runAsUser: 1337
  initContainers:
    - env:
        - name: APPMESH_START_ENABLED
          value: "1"
        - name: APPMESH_IGNORE_UID
          value: "1337"
        - name: APPMESH_ENVOY_INGRESS_PORT
          value: "15000"
        - name: APPMESH_ENVOY_EGRESS_PORT
          value: "15001"
        - name: APPMESH_APP_PORTS
          value: "8080,8081"
        - name: APPMESH_EGRESS_IGNORED_IP
          value: 169.254.169.254
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imagePullPolicy: Always
      name: proxyinit
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
`

const withSidecarExtraPort = `
apiVersion: v1
kind: Pod
metadata:
  name: my-app
  namespace: test-ns
  labels:
    app: my-app
    version: v1
spec:
  containers:
    - image: busybox
      imagePullPolicy: IfNotPresent
      name: my-app
      ports:
        - containerPort: 8080
    - image: busybox
      imagePullPolicy: IfNotPresent
      name: my-other-app
      ports:
        - containerPort: 8081
    - env:
        - name: APPMESH_VIRTUAL_NODE_NAME
          value: mesh/appmesh/virtualNode/my-app-vn
        - name: ENVOY_LOG_LEVEL
          value: debug
        - name: AWS_REGION
          value: us-east-1
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imagePullPolicy: IfNotPresent
      name: envoy
      securityContext:
        procMount: Default
        runAsUser: 1337
  initContainers:
    - env:
        - name: APPMESH_START_ENABLED
          value: "1"
        - name: APPMESH_IGNORE_UID
          value: "1337"
        - name: APPMESH_ENVOY_INGRESS_PORT
          value: "15000"
        - name: APPMESH_ENVOY_EGRESS_PORT
          value: "15001"
        - name: APPMESH_APP_PORTS
          value: "8080,8081,9999"
        - name: APPMESH_EGRESS_IGNORED_IP
          value: 169.254.169.254
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imagePullPolicy: Always
      name: proxyinit
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
`

const incorrectVirtualNodeEnvFormat = `
apiVersion: v1
kind: Pod
metadata:
  name: my-app
  namespace: test-ns
  labels:
    app: my-app
    version: v1
spec:
  containers:
    - image: busybox
      imagePullPolicy: IfNotPresent
      name: my-app
      ports:
        - containerPort: 8080
    - env:
        - name: APPMESH_VIRTUAL_NODE_NAME
          value: appmesh/virtualNode/my-app-vn
        - name: ENVOY_LOG_LEVEL
          value: debug
        - name: AWS_REGION
          value: us-east-1
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imagePullPolicy: IfNotPresent
      name: envoy
      securityContext:
        procMount: Default
        runAsUser: 1337
  initContainers:
    - env:
        - name: APPMESH_START_ENABLED
          value: "1"
        - name: APPMESH_IGNORE_UID
          value: "1337"
        - name: APPMESH_ENVOY_INGRESS_PORT
          value: "15000"
        - name: APPMESH_ENVOY_EGRESS_PORT
          value: "15001"
        - name: APPMESH_APP_PORTS
          value: "8080,8081"
        - name: APPMESH_EGRESS_IGNORED_IP
          value: 169.254.169.254
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imagePullPolicy: Always
      name: proxyinit
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
`

const badPortsFormat = `
apiVersion: v1
kind: Pod
metadata:
  name: my-app
  namespace: test-ns
  labels:
    app: my-app
    version: v1
spec:
  containers:
    - image: busybox
      imagePullPolicy: IfNotPresent
      name: my-app
      ports:
        - containerPort: 8080
    - env:
        - name: APPMESH_VIRTUAL_NODE_NAME
          value: mesh/appmesh/virtualNode/my-app-vn
        - name: ENVOY_LOG_LEVEL
          value: debug
        - name: AWS_REGION
          value: us-east-1
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imagePullPolicy: IfNotPresent
      name: envoy
      securityContext:
        procMount: Default
        runAsUser: 1337
  initContainers:
    - env:
        - name: APPMESH_START_ENABLED
          value: "1"
        - name: APPMESH_IGNORE_UID
          value: "1337"
        - name: APPMESH_ENVOY_INGRESS_PORT
          value: "15000"
        - name: APPMESH_ENVOY_EGRESS_PORT
          value: "15001"
        - name: APPMESH_APP_PORTS
          value: "8080,8081,"
        - name: APPMESH_EGRESS_IGNORED_IP
          value: 169.254.169.254
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imagePullPolicy: Always
      name: proxyinit
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
`

const noPortsEnv = `
apiVersion: v1
kind: Pod
metadata:
  name: my-app
  namespace: test-ns
  labels:
    app: my-app
    version: v1
spec:
  containers:
    - image: busybox
      imagePullPolicy: IfNotPresent
      name: my-app
      ports:
        - containerPort: 8080
    - env:
        - name: APPMESH_VIRTUAL_NODE_NAME
          value: mesh/appmesh/virtualNode/my-app-vn
        - name: ENVOY_LOG_LEVEL
          value: debug
        - name: AWS_REGION
          value: us-east-1
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imagePullPolicy: IfNotPresent
      name: envoy
      securityContext:
        procMount: Default
        runAsUser: 1337
  initContainers:
    - env:
        - name: APPMESH_START_ENABLED
          value: "1"
        - name: APPMESH_IGNORE_UID
          value: "1337"
        - name: APPMESH_ENVOY_INGRESS_PORT
          value: "15000"
        - name: APPMESH_ENVOY_EGRESS_PORT
          value: "15001"
        - name: APPMESH_EGRESS_IGNORED_IP
          value: 169.254.169.254
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imagePullPolicy: Always
      name: proxyinit
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
`

const upstreamsForInjectedPod = `
metadata:
  labels:
    app: my-app
    discovered_by: kubernetesplugin
  name: test-ns-my-app-8080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: my-app
    serviceName: my-app
    serviceNamespace: test-ns
    servicePort: 8080
---
metadata:
  labels:
    app: my-app
    discovered_by: kubernetesplugin
  name: test-ns-my-app-8081
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: my-app
    serviceName: my-app
    serviceNamespace: test-ns
    servicePort: 8081
---
metadata:
  labels:
    app: my-app
    discovered_by: kubernetesplugin
  name: test-ns-my-app-v1-8080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: my-app
      version: v1
    serviceName: my-app
    serviceNamespace: test-ns
    servicePort: 8080
---
metadata:
  labels:
    app: my-app
    discovered_by: kubernetesplugin
  name: test-ns-my-app-v1-8081
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: my-app
      version: v1
    serviceName: my-app
    serviceNamespace: test-ns
    servicePort: 8081
`

const extraUpstreamForInjectedPod = `
metadata:
  labels:
    app: my-app
    discovered_by: kubernetesplugin
  name: test-ns-my-app-8080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: my-app
    serviceName: another-svc-matching-vn
    serviceNamespace: test-ns
    servicePort: 8080`

const upstreamForOtherPod = `
metadata:
  labels:
    app: no-sidecar
    discovered_by: kubernetesplugin
  name: default-no-sidecar-app-1234
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: no-sidecar
    serviceName: no-sidecar-app
    serviceNamespace: test-ns
    servicePort: 1234`
