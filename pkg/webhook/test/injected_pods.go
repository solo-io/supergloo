package test

import (
	"strings"

	"github.com/ghodss/yaml"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/protoutils"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	kubev1 "k8s.io/api/core/v1"
)

var strPodList = []string{productPage, details, ratingsv1, reviewsv1, reviewsv2, reviewsv3}

func MustGetInjectedPodList() v1.PodList {
	list, err := getInjectedPodList()
	Expect(err).NotTo(HaveOccurred())
	return list
}

func getInjectedPodList() (v1.PodList, error) {
	var podList v1.PodList
	for _, v := range strPodList {
		var podObj kubev1.Pod
		err := yaml.Unmarshal([]byte(v), &podObj)
		if err != nil {
			return nil, err
		}
		customPod := kubernetes.FromKubePod(&podObj)
		podList = append(podList, customPod)
	}
	return podList, nil
}

func MustGetUpstreamList() gloov1.UpstreamList {
	list, err := getUpstreamList()
	Expect(err).NotTo(HaveOccurred())
	return list
}

func getUpstreamList() (gloov1.UpstreamList, error) {
	splitUpstreams := strings.Split(upstreams, "---")
	var upstreamList gloov1.UpstreamList
	for _, v := range splitUpstreams {
		var us gloov1.Upstream
		err := protoutils.UnmarshalYaml([]byte(v), &us)
		if err != nil {
			return nil, err
		}
		upstreamList = append(upstreamList, &us)
	}

	return upstreamList, nil
}

const reviewsv3 = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-08T14:47:55Z"
  generateName: reviews-v3-748456d47b-
  labels:
    app: reviews
    pod-template-hash: 748456d47b
    version: v3
    vn: reviews-v3
  name: reviews-v3-748456d47b-swxdd
  namespace: namespace-with-inject
  ownerReferences:
    - apiVersion: apps/v1
      blockOwnerDeletion: true
      controller: true
      kind: ReplicaSet
      name: reviews-v3-748456d47b
      uid: 4af2cedc-5a0d-11e9-9275-123a211826c2
  resourceVersion: "4410630"
  selfLink: /api/v1/namespaces/namespace-with-inject/pods/reviews-v3-748456d47b-swxdd
  uid: 4af4e6f1-5a0d-11e9-9275-123a211826c2
spec:
  containers:
    - image: istio/examples-bookinfo-reviews-v3:1.8.0
      imagePullPolicy: IfNotPresent
      name: reviews
      ports:
        - containerPort: 9080
          protocol: TCP
      resources: {}
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
      volumeMounts:
        - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
          name: default-token-28nrg
          readOnly: true
    - env:
        - name: APPMESH_VIRTUAL_NODE_NAME
          value: mesh/appmesh/virtualNode/reviews-v3
        - name: ENVOY_LOG_LEVEL
          value: debug
        - name: AWS_REGION
          value: us-east-1
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imagePullPolicy: IfNotPresent
      name: envoy
      resources: {}
      securityContext:
        procMount: Default
        runAsUser: 1337
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  dnsPolicy: ClusterFirst
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
          value: "9080"
        - name: APPMESH_EGRESS_IGNORED_IP
          value: 169.254.169.254
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imagePullPolicy: Always
      name: proxyinit
      resources: {}
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
        procMount: Default
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  nodeName: ip-192-168-174-17.ec2.internal
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
    - effect: NoExecute
      key: node.kubernetes.io/not-ready
      operator: Exists
      tolerationSeconds: 300
    - effect: NoExecute
      key: node.kubernetes.io/unreachable
      operator: Exists
      tolerationSeconds: 300
  volumes:
    - name: default-token-28nrg
      secret:
        defaultMode: 420
        secretName: default-token-28nrg
status:
  conditions:
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:57Z"
      status: "True"
      type: Initialized
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:58Z"
      status: "True"
      type: Ready
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:58Z"
      status: "True"
      type: ContainersReady
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:55Z"
      status: "True"
      type: PodScheduled
  containerStatuses:
    - containerID: docker://d95d6e62b6debb6025623bf6ca7c52ea7c75d8065b796441aad0b6e0f4f8f86d
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy@sha256:8ed0bc5cbb92dcea18ae02f75becf5646960b76239f28d83716166047f023f32
      lastState: {}
      name: envoy
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:58Z"
    - containerID: docker://bb3964b6fe1ce06ff1b914fe38807fed3218d3294686a5cc8cea0dfc01f7a0e1
      image: istio/examples-bookinfo-reviews-v3:1.8.0
      imageID: docker-pullable://istio/examples-bookinfo-reviews-v3@sha256:8c0385f0ca799e655d8770b52cb4618ba54e8966a0734ab1aeb6e8b14e171a3b
      lastState: {}
      name: reviews
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:58Z"
  hostIP: 192.168.174.17
  initContainerStatuses:
    - containerID: docker://802226d3cd85441f9215157e1ecc58990a8a400bd49540a21026c65ef94c2e90
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager@sha256:a055da31668a5dc6e68da49c4a8217726d8437e2a94ce6bb6a15abfdcbb1e925
      lastState: {}
      name: proxyinit
      ready: true
      restartCount: 0
      state:
        terminated:
          containerID: docker://802226d3cd85441f9215157e1ecc58990a8a400bd49540a21026c65ef94c2e90
          exitCode: 0
          finishedAt: "2019-04-08T14:47:57Z"
          reason: Completed
          startedAt: "2019-04-08T14:47:57Z"
  phase: Running
  podIP: 192.168.134.153
  qosClass: BestEffort
  startTime: "2019-04-08T14:47:55Z"
`

const reviewsv2 = `

apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-08T14:47:54Z"
  generateName: reviews-v2-cbd94c99b-
  labels:
    app: reviews
    pod-template-hash: cbd94c99b
    version: v2
    vn: reviews-v2
  name: reviews-v2-cbd94c99b-xcbgv
  namespace: namespace-with-inject
  ownerReferences:
    - apiVersion: apps/v1
      blockOwnerDeletion: true
      controller: true
      kind: ReplicaSet
      name: reviews-v2-cbd94c99b
      uid: 4ae31eda-5a0d-11e9-9275-123a211826c2
  resourceVersion: "4410663"
  selfLink: /api/v1/namespaces/namespace-with-inject/pods/reviews-v2-cbd94c99b-xcbgv
  uid: 4ae583ca-5a0d-11e9-9275-123a211826c2
spec:
  containers:
    - image: istio/examples-bookinfo-reviews-v2:1.8.0
      imagePullPolicy: IfNotPresent
      name: reviews
      ports:
        - containerPort: 9080
          protocol: TCP
      resources: {}
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
      volumeMounts:
        - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
          name: default-token-28nrg
          readOnly: true
    - env:
        - name: APPMESH_VIRTUAL_NODE_NAME
          value: mesh/appmesh/virtualNode/reviews-v2
        - name: ENVOY_LOG_LEVEL
          value: debug
        - name: AWS_REGION
          value: us-east-1
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imagePullPolicy: IfNotPresent
      name: envoy
      resources: {}
      securityContext:
        procMount: Default
        runAsUser: 1337
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  dnsPolicy: ClusterFirst
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
          value: "9080"
        - name: APPMESH_EGRESS_IGNORED_IP
          value: 169.254.169.254
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imagePullPolicy: Always
      name: proxyinit
      resources: {}
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
        procMount: Default
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  nodeName: ip-192-168-225-224.ec2.internal
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
    - effect: NoExecute
      key: node.kubernetes.io/not-ready
      operator: Exists
      tolerationSeconds: 300
    - effect: NoExecute
      key: node.kubernetes.io/unreachable
      operator: Exists
      tolerationSeconds: 300
  volumes:
    - name: default-token-28nrg
      secret:
        defaultMode: 420
        secretName: default-token-28nrg
status:
  conditions:
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:58Z"
      status: "True"
      type: Initialized
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:59Z"
      status: "True"
      type: Ready
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:59Z"
      status: "True"
      type: ContainersReady
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:54Z"
      status: "True"
      type: PodScheduled
  containerStatuses:
    - containerID: docker://67947822e3ce0fa5313441ff527cda4eb1e3c821693038bcb28f507e450c3b1d
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy@sha256:8ed0bc5cbb92dcea18ae02f75becf5646960b76239f28d83716166047f023f32
      lastState: {}
      name: envoy
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:58Z"
    - containerID: docker://258c7ede4284751a9bcdda6aced7761478431a2b2518d52e41516e7e8b5c8ca2
      image: istio/examples-bookinfo-reviews-v2:1.8.0
      imageID: docker-pullable://istio/examples-bookinfo-reviews-v2@sha256:d2483dcb235b27309680177726e4e86905d66e47facaf1d57ed590b2bf95c8ad
      lastState: {}
      name: reviews
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:58Z"
  hostIP: 192.168.225.224
  initContainerStatuses:
    - containerID: docker://a02d2264b299f0de08841497e392a3b3d66ab6d080c888118cf82564f3476e80
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager@sha256:a055da31668a5dc6e68da49c4a8217726d8437e2a94ce6bb6a15abfdcbb1e925
      lastState: {}
      name: proxyinit
      ready: true
      restartCount: 0
      state:
        terminated:
          containerID: docker://a02d2264b299f0de08841497e392a3b3d66ab6d080c888118cf82564f3476e80
          exitCode: 0
          finishedAt: "2019-04-08T14:47:57Z"
          reason: Completed
          startedAt: "2019-04-08T14:47:57Z"
  phase: Running
  podIP: 192.168.215.76
  qosClass: BestEffort
  startTime: "2019-04-08T14:47:54Z"

`

const reviewsv1 = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-08T14:47:54Z"
  generateName: reviews-v1-85b7d84c56-
  labels:
    app: reviews
    pod-template-hash: 85b7d84c56
    version: v1
    vn: reviews-v1
  name: reviews-v1-85b7d84c56-86jdf
  namespace: namespace-with-inject
  ownerReferences:
    - apiVersion: apps/v1
      blockOwnerDeletion: true
      controller: true
      kind: ReplicaSet
      name: reviews-v1-85b7d84c56
      uid: 4ad4200b-5a0d-11e9-9275-123a211826c2
  resourceVersion: "4410659"
  selfLink: /api/v1/namespaces/namespace-with-inject/pods/reviews-v1-85b7d84c56-86jdf
  uid: 4ad639ab-5a0d-11e9-9275-123a211826c2
spec:
  containers:
    - image: istio/examples-bookinfo-reviews-v1:1.8.0
      imagePullPolicy: IfNotPresent
      name: reviews
      ports:
        - containerPort: 9080
          protocol: TCP
      resources: {}
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
      volumeMounts:
        - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
          name: default-token-28nrg
          readOnly: true
    - env:
        - name: APPMESH_VIRTUAL_NODE_NAME
          value: mesh/appmesh/virtualNode/reviews-v1
        - name: ENVOY_LOG_LEVEL
          value: debug
        - name: AWS_REGION
          value: us-east-1
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imagePullPolicy: IfNotPresent
      name: envoy
      resources: {}
      securityContext:
        procMount: Default
        runAsUser: 1337
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  dnsPolicy: ClusterFirst
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
          value: "9080"
        - name: APPMESH_EGRESS_IGNORED_IP
          value: 169.254.169.254
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imagePullPolicy: Always
      name: proxyinit
      resources: {}
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
        procMount: Default
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  nodeName: ip-192-168-79-145.ec2.internal
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
    - effect: NoExecute
      key: node.kubernetes.io/not-ready
      operator: Exists
      tolerationSeconds: 300
    - effect: NoExecute
      key: node.kubernetes.io/unreachable
      operator: Exists
      tolerationSeconds: 300
  volumes:
    - name: default-token-28nrg
      secret:
        defaultMode: 420
        secretName: default-token-28nrg
status:
  conditions:
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:58Z"
      status: "True"
      type: Initialized
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:59Z"
      status: "True"
      type: Ready
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:59Z"
      status: "True"
      type: ContainersReady
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:54Z"
      status: "True"
      type: PodScheduled
  containerStatuses:
    - containerID: docker://da0197bcc63253cfe51c86c545168f75b61918194360b18436d6b8dea40a00ce
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy@sha256:8ed0bc5cbb92dcea18ae02f75becf5646960b76239f28d83716166047f023f32
      lastState: {}
      name: envoy
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:59Z"
    - containerID: docker://8d40c5fd36def8e99967a02d41ee8a9955034d1fdaa8aca025e616443f9207bc
      image: istio/examples-bookinfo-reviews-v1:1.8.0
      imageID: docker-pullable://istio/examples-bookinfo-reviews-v1@sha256:920d46b3c526376b28b90d0e895ca7682d36132e6338301fcbcd567ef81bde05
      lastState: {}
      name: reviews
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:58Z"
  hostIP: 192.168.79.145
  initContainerStatuses:
    - containerID: docker://c5882ae3424c1b8cf717f6029dcd49bd1913f5b9933aea3ce53d255da2813b3a
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager@sha256:a055da31668a5dc6e68da49c4a8217726d8437e2a94ce6bb6a15abfdcbb1e925
      lastState: {}
      name: proxyinit
      ready: true
      restartCount: 0
      state:
        terminated:
          containerID: docker://c5882ae3424c1b8cf717f6029dcd49bd1913f5b9933aea3ce53d255da2813b3a
          exitCode: 0
          finishedAt: "2019-04-08T14:47:57Z"
          reason: Completed
          startedAt: "2019-04-08T14:47:57Z"
  phase: Running
  podIP: 192.168.88.125
  qosClass: BestEffort
  startTime: "2019-04-08T14:47:54Z"
`

const ratingsv1 = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-08T14:47:54Z"
  generateName: ratings-v1-7c9949d479-
  labels:
    app: ratings
    pod-template-hash: 7c9949d479
    version: v1
    vn: ratings-v1
  name: ratings-v1-7c9949d479-85tgt
  namespace: namespace-with-inject
  ownerReferences:
    - apiVersion: apps/v1
      blockOwnerDeletion: true
      controller: true
      kind: ReplicaSet
      name: ratings-v1-7c9949d479
      uid: 4ab4b5c2-5a0d-11e9-9275-123a211826c2
  resourceVersion: "4410620"
  selfLink: /api/v1/namespaces/namespace-with-inject/pods/ratings-v1-7c9949d479-85tgt
  uid: 4ab72238-5a0d-11e9-9275-123a211826c2
spec:
  containers:
    - image: istio/examples-bookinfo-ratings-v1:1.8.0
      imagePullPolicy: IfNotPresent
      name: ratings
      ports:
        - containerPort: 9080
          protocol: TCP
      resources: {}
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
      volumeMounts:
        - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
          name: default-token-28nrg
          readOnly: true
    - env:
        - name: APPMESH_VIRTUAL_NODE_NAME
          value: mesh/appmesh/virtualNode/ratings-v1
        - name: ENVOY_LOG_LEVEL
          value: debug
        - name: AWS_REGION
          value: us-east-1
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imagePullPolicy: IfNotPresent
      name: envoy
      resources: {}
      securityContext:
        procMount: Default
        runAsUser: 1337
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  dnsPolicy: ClusterFirst
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
          value: "9080"
        - name: APPMESH_EGRESS_IGNORED_IP
          value: 169.254.169.254
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imagePullPolicy: Always
      name: proxyinit
      resources: {}
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
        procMount: Default
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  nodeName: ip-192-168-225-224.ec2.internal
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
    - effect: NoExecute
      key: node.kubernetes.io/not-ready
      operator: Exists
      tolerationSeconds: 300
    - effect: NoExecute
      key: node.kubernetes.io/unreachable
      operator: Exists
      tolerationSeconds: 300
  volumes:
    - name: default-token-28nrg
      secret:
        defaultMode: 420
        secretName: default-token-28nrg
status:
  conditions:
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:57Z"
      status: "True"
      type: Initialized
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:58Z"
      status: "True"
      type: Ready
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:58Z"
      status: "True"
      type: ContainersReady
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:54Z"
      status: "True"
      type: PodScheduled
  containerStatuses:
    - containerID: docker://4618651a448c68c5a490a875ac20ec10d3ebce4491f3c3ead68d2ba3d065c9d2
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy@sha256:8ed0bc5cbb92dcea18ae02f75becf5646960b76239f28d83716166047f023f32
      lastState: {}
      name: envoy
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:57Z"
    - containerID: docker://1e3732bea90cd727a0b28e47696717f3937219114b3812c0a0d2944969bfee6d
      image: istio/examples-bookinfo-ratings-v1:1.8.0
      imageID: docker-pullable://istio/examples-bookinfo-ratings-v1@sha256:f84302e53e62a8a12510c7f8f437a7a34949be3dfb37f4eb9d054a76436fa4d7
      lastState: {}
      name: ratings
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:57Z"
  hostIP: 192.168.225.224
  initContainerStatuses:
    - containerID: docker://218aeedb79bbec6265098aac5146a376ec4e36fd724c2f018d9d607e27f5ce11
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager@sha256:a055da31668a5dc6e68da49c4a8217726d8437e2a94ce6bb6a15abfdcbb1e925
      lastState: {}
      name: proxyinit
      ready: true
      restartCount: 0
      state:
        terminated:
          containerID: docker://218aeedb79bbec6265098aac5146a376ec4e36fd724c2f018d9d607e27f5ce11
          exitCode: 0
          finishedAt: "2019-04-08T14:47:57Z"
          reason: Completed
          startedAt: "2019-04-08T14:47:56Z"
  phase: Running
  podIP: 192.168.246.234
  qosClass: BestEffort
  startTime: "2019-04-08T14:47:54Z"

`

const productPage = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-08T14:47:55Z"
  generateName: productpage-v1-8d69b45c-
  labels:
    app: productpage
    pod-template-hash: 8d69b45c
    version: v1
    vn: productpage-v1
  name: productpage-v1-8d69b45c-kxmzk
  namespace: namespace-with-inject
  ownerReferences:
    - apiVersion: apps/v1
      blockOwnerDeletion: true
      controller: true
      kind: ReplicaSet
      name: productpage-v1-8d69b45c
      uid: 4b127d3b-5a0d-11e9-9275-123a211826c2
  resourceVersion: "4410645"
  selfLink: /api/v1/namespaces/namespace-with-inject/pods/productpage-v1-8d69b45c-kxmzk
  uid: 4b14c5a1-5a0d-11e9-9275-123a211826c2
spec:
  containers:
    - image: istio/examples-bookinfo-productpage-v1:1.8.0
      imagePullPolicy: IfNotPresent
      name: productpage
      ports:
        - containerPort: 9080
          protocol: TCP
      resources: {}
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
      volumeMounts:
        - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
          name: default-token-28nrg
          readOnly: true
    - env:
        - name: APPMESH_VIRTUAL_NODE_NAME
          value: mesh/appmesh/virtualNode/productpage-v1
        - name: ENVOY_LOG_LEVEL
          value: debug
        - name: AWS_REGION
          value: us-east-1
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imagePullPolicy: IfNotPresent
      name: envoy
      resources: {}
      securityContext:
        procMount: Default
        runAsUser: 1337
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  dnsPolicy: ClusterFirst
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
          value: "9080"
        - name: APPMESH_EGRESS_IGNORED_IP
          value: 169.254.169.254
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imagePullPolicy: Always
      name: proxyinit
      resources: {}
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
        procMount: Default
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  nodeName: ip-192-168-79-145.ec2.internal
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
    - effect: NoExecute
      key: node.kubernetes.io/not-ready
      operator: Exists
      tolerationSeconds: 300
    - effect: NoExecute
      key: node.kubernetes.io/unreachable
      operator: Exists
      tolerationSeconds: 300
  volumes:
    - name: default-token-28nrg
      secret:
        defaultMode: 420
        secretName: default-token-28nrg
status:
  conditions:
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:58Z"
      status: "True"
      type: Initialized
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:59Z"
      status: "True"
      type: Ready
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:59Z"
      status: "True"
      type: ContainersReady
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:55Z"
      status: "True"
      type: PodScheduled
  containerStatuses:
    - containerID: docker://84ab898aac4231713e501eabb158d57ea39e0cbbfbc02b9b513b3c06da3d979c
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy@sha256:8ed0bc5cbb92dcea18ae02f75becf5646960b76239f28d83716166047f023f32
      lastState: {}
      name: envoy
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:59Z"
    - containerID: docker://2e0cafb94859acbb1e44cea7bf7e8eb0268469f97fbe48fb57c836eb88a18728
      image: istio/examples-bookinfo-productpage-v1:1.8.0
      imageID: docker-pullable://istio/examples-bookinfo-productpage-v1@sha256:ed65a39f8b3ec5a7c7973c8e0861b89465998a0617bc0d0c76ce0a97080694a9
      lastState: {}
      name: productpage
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:58Z"
  hostIP: 192.168.79.145
  initContainerStatuses:
    - containerID: docker://52b9b2f7de0219a57432bbcadb01d0e8f43907c59513a88fd0c9eb578f82e6f9
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager@sha256:a055da31668a5dc6e68da49c4a8217726d8437e2a94ce6bb6a15abfdcbb1e925
      lastState: {}
      name: proxyinit
      ready: true
      restartCount: 0
      state:
        terminated:
          containerID: docker://52b9b2f7de0219a57432bbcadb01d0e8f43907c59513a88fd0c9eb578f82e6f9
          exitCode: 0
          finishedAt: "2019-04-08T14:47:57Z"
          reason: Completed
          startedAt: "2019-04-08T14:47:57Z"
  phase: Running
  podIP: 192.168.97.46
  qosClass: BestEffort
  startTime: "2019-04-08T14:47:55Z"
`

const details = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-08T14:47:54Z"
  generateName: details-v1-876bf485f-
  labels:
    app: details
    pod-template-hash: 876bf485f
    version: v1
    vn: details-v1
  name: details-v1-876bf485f-c4ptc
  namespace: namespace-with-inject
  ownerReferences:
    - apiVersion: apps/v1
      blockOwnerDeletion: true
      controller: true
      kind: ReplicaSet
      name: details-v1-876bf485f
      uid: 4a94dafd-5a0d-11e9-9275-123a211826c2
  resourceVersion: "4410678"
  selfLink: /api/v1/namespaces/namespace-with-inject/pods/details-v1-876bf485f-c4ptc
  uid: 4a9790b4-5a0d-11e9-9275-123a211826c2
spec:
  containers:
    - image: istio/examples-bookinfo-details-v1:1.8.0
      imagePullPolicy: IfNotPresent
      name: details
      ports:
        - containerPort: 9080
          protocol: TCP
      resources: {}
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
      volumeMounts:
        - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
          name: default-token-28nrg
          readOnly: true
    - env:
        - name: APPMESH_VIRTUAL_NODE_NAME
          value: mesh/appmesh/virtualNode/details-v1
        - name: ENVOY_LOG_LEVEL
          value: debug
        - name: AWS_REGION
          value: us-east-1
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imagePullPolicy: IfNotPresent
      name: envoy
      resources: {}
      securityContext:
        procMount: Default
        runAsUser: 1337
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  dnsPolicy: ClusterFirst
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
          value: "9080"
        - name: APPMESH_EGRESS_IGNORED_IP
          value: 169.254.169.254
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imagePullPolicy: Always
      name: proxyinit
      resources: {}
      securityContext:
        capabilities:
          add:
            - NET_ADMIN
        procMount: Default
      terminationMessagePath: /dev/termination-log
      terminationMessagePolicy: File
  nodeName: ip-192-168-79-145.ec2.internal
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
    - effect: NoExecute
      key: node.kubernetes.io/not-ready
      operator: Exists
      tolerationSeconds: 300
    - effect: NoExecute
      key: node.kubernetes.io/unreachable
      operator: Exists
      tolerationSeconds: 300
  volumes:
    - name: default-token-28nrg
      secret:
        defaultMode: 420
        secretName: default-token-28nrg
status:
  conditions:
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:58Z"
      status: "True"
      type: Initialized
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:59Z"
      status: "True"
      type: Ready
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:59Z"
      status: "True"
      type: ContainersReady
    - lastProbeTime: null
      lastTransitionTime: "2019-04-08T14:47:54Z"
      status: "True"
      type: PodScheduled
  containerStatuses:
    - containerID: docker://f960236e77475a4973b513dae17607b2470902059867224c9586c3279d7cf4ca
      image: istio/examples-bookinfo-details-v1:1.8.0
      imageID: docker-pullable://istio/examples-bookinfo-details-v1@sha256:73e1de909cd387cf377bb51ddd90167d0f44cf0659746d1d0e50276e8f1c9df3
      lastState: {}
      name: details
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:58Z"
    - containerID: docker://b7b72c21c6b149bd25a323eba545fff1d5fb649bf8d24c755138e565520889a2
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy@sha256:8ed0bc5cbb92dcea18ae02f75becf5646960b76239f28d83716166047f023f32
      lastState: {}
      name: envoy
      ready: true
      restartCount: 0
      state:
        running:
          startedAt: "2019-04-08T14:47:59Z"
  hostIP: 192.168.79.145
  initContainerStatuses:
    - containerID: docker://d4c45d93a1e80e154aafce95b43260d6ac28d24f51374b486ba0b3898d3718c3
      image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager:latest
      imageID: docker-pullable://111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager@sha256:a055da31668a5dc6e68da49c4a8217726d8437e2a94ce6bb6a15abfdcbb1e925
      lastState: {}
      name: proxyinit
      ready: true
      restartCount: 0
      state:
        terminated:
          containerID: docker://d4c45d93a1e80e154aafce95b43260d6ac28d24f51374b486ba0b3898d3718c3
          exitCode: 0
          finishedAt: "2019-04-08T14:47:57Z"
          reason: Completed
          startedAt: "2019-04-08T14:47:56Z"
  phase: Running
  podIP: 192.168.80.139
  qosClass: BestEffort
  startTime: "2019-04-08T14:47:54Z"
`

const upstreams = `
discoveryMetadata: {}
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"details"},"name":"details","namespace":"namespace-with-inject"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"details"}}}
  labels:
    app: details
    discovered_by: kubernetesplugin
  name: namespace-with-inject-details-9080
  namespace: supergloo-system
  resourceVersion: "5713091"
status: {}
upstreamSpec:
  kube:
    selector:
      app: details
    serviceName: details
    serviceNamespace: namespace-with-inject
    servicePort: 9080

---
discoveryMetadata: {}
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"details"},"name":"details","namespace":"namespace-with-inject"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"details"}}}
  labels:
    app: details
    discovered_by: kubernetesplugin
  name: namespace-with-inject-details-v1-details-v1-9080
  namespace: supergloo-system
  resourceVersion: "5713113"
status: {}
upstreamSpec:
  kube:
    selector:
      app: details
      version: v1
      vn: details-v1
    serviceName: details
    serviceNamespace: namespace-with-inject
    servicePort: 9080

---
discoveryMetadata: {}
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"productpage"},"name":"productpage","namespace":"namespace-with-inject"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"productpage"}}}
  labels:
    app: productpage
    discovered_by: kubernetesplugin
  name: namespace-with-inject-productpage-9080
  namespace: supergloo-system
  resourceVersion: "5713217"
status: {}
upstreamSpec:
  kube:
    selector:
      app: productpage
    serviceName: productpage
    serviceNamespace: namespace-with-inject
    servicePort: 9080

---
discoveryMetadata: {}
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"productpage"},"name":"productpage","namespace":"namespace-with-inject"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"productpage"}}}
  labels:
    app: productpage
    discovered_by: kubernetesplugin
  name: namespace-with-inject-productpage-v1-productpage-v1-9080
  namespace: supergloo-system
  resourceVersion: "5713244"
status: {}
upstreamSpec:
  kube:
    selector:
      app: productpage
      version: v1
      vn: productpage-v1
    serviceName: productpage
    serviceNamespace: namespace-with-inject
    servicePort: 9080

---
discoveryMetadata: {}
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"ratings"},"name":"ratings","namespace":"namespace-with-inject"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"ratings"}}}
  labels:
    app: ratings
    discovered_by: kubernetesplugin
  name: namespace-with-inject-ratings-9080
  namespace: supergloo-system
  resourceVersion: "5713124"
status: {}
upstreamSpec:
  kube:
    selector:
      app: ratings
    serviceName: ratings
    serviceNamespace: namespace-with-inject
    servicePort: 9080

---
discoveryMetadata: {}
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"ratings"},"name":"ratings","namespace":"namespace-with-inject"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"ratings"}}}
  labels:
    app: ratings
    discovered_by: kubernetesplugin
  name: namespace-with-inject-ratings-v1-ratings-v1-9080
  namespace: supergloo-system
  resourceVersion: "5713151"
status: {}
upstreamSpec:
  kube:
    selector:
      app: ratings
      version: v1
      vn: ratings-v1
    serviceName: ratings
    serviceNamespace: namespace-with-inject
    servicePort: 9080

---
discoveryMetadata: {}
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"reviews"},"name":"reviews","namespace":"namespace-with-inject"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"reviews","version":"v1","vn":"reviews-v1"}}}
  labels:
    app: reviews
    discovered_by: kubernetesplugin
  name: namespace-with-inject-reviews-9080
  namespace: supergloo-system
  resourceVersion: "5713253"
status: {}
upstreamSpec:
  kube:
    selector:
      app: reviews
      version: v1
      vn: reviews-v1
    serviceName: reviews
    serviceNamespace: namespace-with-inject
    servicePort: 9080

---
discoveryMetadata: {}
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"reviews"},"name":"reviews-v2","namespace":"namespace-with-inject"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"reviews","version":"v2","vn":"reviews-v2"}}}
  labels:
    app: reviews
    discovered_by: kubernetesplugin
  name: namespace-with-inject-reviews-v2-9080
  namespace: supergloo-system
  resourceVersion: "5713282"
status: {}
upstreamSpec:
  kube:
    selector:
      app: reviews
      version: v2
      vn: reviews-v2
    serviceName: reviews-v2
    serviceNamespace: namespace-with-inject
    servicePort: 9080

---
discoveryMetadata: {}
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"reviews"},"name":"reviews-v3","namespace":"namespace-with-inject"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"reviews","version":"v3","vn":"reviews-v3"}}}
  labels:
    app: reviews
    discovered_by: kubernetesplugin
  name: namespace-with-inject-reviews-v3-9080
  namespace: supergloo-system
  resourceVersion: "5713309"
status: {}
upstreamSpec:
  kube:
    selector:
      app: reviews
      version: v3
      vn: reviews-v3
    serviceName: reviews-v3
    serviceNamespace: namespace-with-inject
    servicePort: 9080

---
discoveryMetadata: {}
metadata:
  annotations: {}
  labels:
    app: sidecar-injector
    discovered_by: kubernetesplugin
  name: supergloo-system-sidecar-injector-443
  namespace: supergloo-system
  resourceVersion: "5712958"
status: {}
upstreamSpec:
  kube:
    selector:
      app: sidecar-injector
    serviceName: sidecar-injector
    serviceNamespace: supergloo-system
    servicePort: 443
`
