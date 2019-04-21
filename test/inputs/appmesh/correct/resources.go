package correct

import (
	appmeshApi "github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/solo-io/supergloo/test/inputs/appmesh"
)

func GetAllResources() appmesh.TestResourceSet {
	resources := make(appmesh.TestResourceSet)

	resources["product-page"] = appmesh.PodsServicesUpstreamsTuple{
		Pods:      []string{productPagePod},
		Services:  []string{productPageSvc},
		Upstreams: []string{productPageUs},
	}

	resources["details"] = appmesh.PodsServicesUpstreamsTuple{
		Pods:      []string{detailsPod},
		Services:  []string{detailsSvc},
		Upstreams: []string{detailsUs},
	}

	resources["reviews-v1"] = appmesh.PodsServicesUpstreamsTuple{
		Pods:      []string{reviewsV1Pod},
		Services:  []string{reviewsV1Svc},
		Upstreams: []string{reviewsV1Us},
	}

	resources["reviews-v2"] = appmesh.PodsServicesUpstreamsTuple{
		Pods:      []string{reviewsV2Pod1, reviewsV2Pod2},
		Services:  []string{reviewsV2Svc},
		Upstreams: []string{reviewsV2Us},
	}

	resources["reviews-v3"] = appmesh.PodsServicesUpstreamsTuple{
		Pods:      []string{reviewsV3Pod},
		Services:  []string{reviewsV3Svc},
		Upstreams: []string{reviewsV3Us},
	}

	resources["ratings"] = appmesh.PodsServicesUpstreamsTuple{
		Pods:      []string{ratingsPod},
		Services:  []string{ratingsSvc},
		Upstreams: []string{ratingsUs},
	}

	resources["other"] = appmesh.PodsServicesUpstreamsTuple{
		Pods:      []string{discoveryPod, appmesh.KubernetesPods},
		Services:  []string{appmesh.KubernetesServices},
		Upstreams: []string{appmesh.KubernetesUpstreams},
	}

	return resources
}

func GetAppMeshRelatedResources() appmesh.TestResourceSet {
	resources := make(appmesh.TestResourceSet)
	for name, tuple := range GetAllResources() {
		if name != "other" {
			resources[name] = tuple
		}
	}
	return resources
}

// Returns the virtual node set as it is before any Routing Rules have been processed or all traffic has been allowed
func GetExpectedInitVirtualNodes() map[string]*appmeshApi.VirtualNodeData {
	return map[string]*appmeshApi.VirtualNodeData{
		productPageHostname: createVirtualNode(productPageVnName, productPageHostname, MeshName, "http", 9080, nil),
		detailsHostname:     createVirtualNode(detailsVnName, detailsHostname, MeshName, "http", 9080, nil),
		reviewsV1Hostname:   createVirtualNode(reviewsV1VnName, reviewsV1Hostname, MeshName, "http", 9080, nil),
		reviewsV2Hostname:   createVirtualNode(reviewsV2VnName, reviewsV2Hostname, MeshName, "http", 9080, nil),
		reviewsV3Hostname:   createVirtualNode(reviewsV3VnName, reviewsV3Hostname, MeshName, "http", 9080, nil),
		ratingsHostname:     createVirtualNode(ratingsVnName, ratingsHostname, MeshName, "http", 9080, nil),
	}
}

// Returns the virtual node set as it is expected to be after allowing all traffic (no Routing Rules)
func GetExpectedVirtualNodesOnlyAllowAll() map[string]*appmeshApi.VirtualNodeData {
	return map[string]*appmeshApi.VirtualNodeData{
		productPageHostname: createVirtualNode(productPageVnName, productPageHostname, MeshName, "http", 9080, allHostsMinus(productPageHostname)),
		detailsHostname:     createVirtualNode(detailsVnName, detailsHostname, MeshName, "http", 9080, allHostsMinus(detailsHostname)),
		reviewsV1Hostname:   createVirtualNode(reviewsV1VnName, reviewsV1Hostname, MeshName, "http", 9080, allHostsMinus(reviewsV1Hostname)),
		reviewsV2Hostname:   createVirtualNode(reviewsV2VnName, reviewsV2Hostname, MeshName, "http", 9080, allHostsMinus(reviewsV2Hostname)),
		reviewsV3Hostname:   createVirtualNode(reviewsV3VnName, reviewsV3Hostname, MeshName, "http", 9080, allHostsMinus(reviewsV3Hostname)),
		ratingsHostname:     createVirtualNode(ratingsVnName, ratingsHostname, MeshName, "http", 9080, allHostsMinus(ratingsHostname)),
	}
}

func createVirtualNode(name, host, mesh, protocol string, port int64, backendHosts []string) *appmeshApi.VirtualNodeData {
	var backends []*appmeshApi.Backend
	for _, vs := range backendHosts {
		vsName := vs
		backends = append(backends, &appmeshApi.Backend{
			VirtualService: &appmeshApi.VirtualServiceBackend{
				VirtualServiceName: &vsName,
			},
		})
	}
	return &appmeshApi.VirtualNodeData{
		MeshName:        &mesh,
		VirtualNodeName: &name,
		Spec: &appmeshApi.VirtualNodeSpec{
			ServiceDiscovery: &appmeshApi.ServiceDiscovery{
				Dns: &appmeshApi.DnsServiceDiscovery{
					Hostname: &host,
				},
			},
			Listeners: []*appmeshApi.Listener{
				{
					PortMapping: &appmeshApi.PortMapping{
						Port:     &port,
						Protocol: &protocol,
					},
				},
			},
			Backends: backends,
		},
	}
}

// Returns a slice of all hosts except the given one
func allHostsMinus(excludedHost string) (out []string) {
	for _, host := range allHostnames {
		if host != excludedHost {
			out = append(out, host)
		}
	}
	return
}

var (
	MeshName            = "test-mesh"
	productPageVnName   = "productpage-vn"
	productPageHostname = "productpage.default.svc.cluster.local"
	detailsVnName       = "details-vn"
	detailsHostname     = "details.default.svc.cluster.local"
	reviewsV1VnName     = "reviews-v1-vn"
	reviewsV1Hostname   = "reviews.default.svc.cluster.local"
	reviewsV2VnName     = "reviews-v2-vn"
	reviewsV2Hostname   = "reviews-v2.default.svc.cluster.local"
	reviewsV3VnName     = "reviews-v3-vn"
	reviewsV3Hostname   = "reviews-v3.default.svc.cluster.local"
	ratingsVnName       = "ratings-vn"
	ratingsHostname     = "ratings.default.svc.cluster.local"
	allHostnames        = []string{productPageHostname, detailsHostname, reviewsV1Hostname, reviewsV2Hostname, reviewsV3Hostname, ratingsHostname}
)

// Pods
var detailsPod = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-19T21:38:58Z"
  generateName: details-v1-858dbf857d-
  labels:
    app: details
    pod-template-hash: 858dbf857d
    version: v1
  name: details-v1-858dbf857d-w2bcj
  namespace: default
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: details-v1-858dbf857d
    uid: 8a2d1f7d-62eb-11e9-989c-08002717ab34
  resourceVersion: "208156"
  selfLink: /api/v1/namespaces/default/pods/details-v1-858dbf857d-w2bcj
  uid: 8a3012f6-62eb-11e9-989c-08002717ab34
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
      name: default-token-kmzll
      readOnly: true
  - env:
    - name: APPMESH_VIRTUAL_NODE_NAME
      value: mesh/test-mesh/virtualNode/` + detailsVnName + `
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  nodeName: minikube
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
  - name: default-token-kmzll
    secret:
      defaultMode: 420
      secretName: default-token-kmzll
`

var productPagePod = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-19T21:38:59Z"
  generateName: productpage-v1-5667d4c7d-
  labels:
    app: productpage
    pod-template-hash: 5667d4c7d
    version: v1
  name: productpage-v1-5667d4c7d-9gptr
  namespace: default
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: productpage-v1-5667d4c7d
    uid: 8ab4baa5-62eb-11e9-989c-08002717ab34
  resourceVersion: "208157"
  selfLink: /api/v1/namespaces/default/pods/productpage-v1-5667d4c7d-9gptr
  uid: 8ab72a0e-62eb-11e9-989c-08002717ab34
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
      name: default-token-kmzll
      readOnly: true
  - env:
    - name: APPMESH_VIRTUAL_NODE_NAME
      value: mesh/test-mesh/virtualNode/` + productPageVnName + `
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  nodeName: minikube
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
  - name: default-token-kmzll
    secret:
      defaultMode: 420
      secretName: default-token-kmzll
  
`

var ratingsPod = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-19T21:38:58Z"
  generateName: ratings-v1-d95d475fd-
  labels:
    app: ratings
    pod-template-hash: d95d475fd
    version: v1
  name: ratings-v1-d95d475fd-7krmm
  namespace: default
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: ratings-v1-d95d475fd
    uid: 8a3c1f41-62eb-11e9-989c-08002717ab34
  resourceVersion: "208164"
  selfLink: /api/v1/namespaces/default/pods/ratings-v1-d95d475fd-7krmm
  uid: 8a3db26e-62eb-11e9-989c-08002717ab34
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
      name: default-token-kmzll
      readOnly: true
  - env:
    - name: APPMESH_VIRTUAL_NODE_NAME
      value: mesh/test-mesh/virtualNode/` + ratingsVnName + `
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  nodeName: minikube
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
  - name: default-token-kmzll
    secret:
      defaultMode: 420
      secretName: default-token-kmzll
`

var reviewsV1Pod = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-19T21:38:58Z"
  generateName: reviews-v1-84598f78fd-
  labels:
    app: reviews
    pod-template-hash: 84598f78fd
    version: v1
  name: reviews-v1-84598f78fd-hqmms
  namespace: default
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: reviews-v1-84598f78fd
    uid: 8a5acc4f-62eb-11e9-989c-08002717ab34
  resourceVersion: "208151"
  selfLink: /api/v1/namespaces/default/pods/reviews-v1-84598f78fd-hqmms
  uid: 8a5c9d9a-62eb-11e9-989c-08002717ab34
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
      name: default-token-kmzll
      readOnly: true
  - env:
    - name: APPMESH_VIRTUAL_NODE_NAME
      value: mesh/test-mesh/virtualNode/` + reviewsV1VnName + `
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  nodeName: minikube
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
  - name: default-token-kmzll
    secret:
      defaultMode: 420
      secretName: default-token-kmzll
`

var reviewsV2Pod1 = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-19T21:38:59Z"
  generateName: reviews-v2-bbc775b79-
  labels:
    app: reviews
    pod-template-hash: bbc775b79
    version: v2
  name: reviews-v2-bbc775b79-5w9xz
  namespace: default
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: reviews-v2-bbc775b79
    uid: 8a734e87-62eb-11e9-989c-08002717ab34
  resourceVersion: "208514"
  selfLink: /api/v1/namespaces/default/pods/reviews-v2-bbc775b79-5w9xz
  uid: 8a776dce-62eb-11e9-989c-08002717ab34
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
      name: default-token-kmzll
      readOnly: true
  - env:
    - name: APPMESH_VIRTUAL_NODE_NAME
      value: mesh/test-mesh/virtualNode/` + reviewsV2VnName + `
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  nodeName: minikube
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
  - name: default-token-kmzll
    secret:
      defaultMode: 420
      secretName: default-token-kmzll
`
var reviewsV2Pod2 = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-19T21:38:59Z"
  generateName: reviews-v2-bbc775b79-
  labels:
    app: reviews
    pod-template-hash: bbc775b79
    version: v2
  name: reviews-v2-bbc775b79-xcgrg
  namespace: default
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: reviews-v2-bbc775b79
    uid: 8a734e87-62eb-11e9-989c-08002717ab34
  resourceVersion: "208511"
  selfLink: /api/v1/namespaces/default/pods/reviews-v2-bbc775b79-xcgrg
  uid: 8a7c04ac-62eb-11e9-989c-08002717ab34
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
      name: default-token-kmzll
      readOnly: true
  - env:
    - name: APPMESH_VIRTUAL_NODE_NAME
      value: mesh/test-mesh/virtualNode/` + reviewsV2VnName + `
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  nodeName: minikube
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
  - name: default-token-kmzll
    secret:
      defaultMode: 420
      secretName: default-token-kmzll
`
var reviewsV3Pod = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-19T21:38:59Z"
  generateName: reviews-v3-55956c49fb-
  labels:
    app: reviews
    pod-template-hash: 55956c49fb
    version: v3
  name: reviews-v3-55956c49fb-xdbgp
  namespace: default
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: reviews-v3-55956c49fb
    uid: 8a866b0c-62eb-11e9-989c-08002717ab34
  resourceVersion: "208154"
  selfLink: /api/v1/namespaces/default/pods/reviews-v3-55956c49fb-xdbgp
  uid: 8a8f7882-62eb-11e9-989c-08002717ab34
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
      name: default-token-kmzll
      readOnly: true
  - env:
    - name: APPMESH_VIRTUAL_NODE_NAME
      value: mesh/test-mesh/virtualNode/` + reviewsV3VnName + `
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
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
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-kmzll
      readOnly: true
  nodeName: minikube
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
  - name: default-token-kmzll
    secret:
      defaultMode: 420
      secretName: default-token-kmzll
`

const discoveryPod = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-18T15:19:25Z"
  generateName: discovery-5f98c7855c-
  labels:
    gloo: discovery
    pod-template-hash: 5f98c7855c
  name: discovery-5f98c7855c-grqlv
  namespace: gloo-system
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: discovery-5f98c7855c
    uid: 59b2b913-61ed-11e9-a50c-08002717ab34
  resourceVersion: "171467"
  selfLink: /api/v1/namespaces/gloo-system/pods/discovery-5f98c7855c-grqlv
  uid: 59b3b33e-61ed-11e9-a50c-08002717ab34
spec:
  containers:
  - env:
    - name: POD_NAMESPACE
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: metadata.namespace
    image: quay.io/solo-io/discovery:0.13.14
    imagePullPolicy: Always
    name: discovery
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: default-token-pmkp9
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: minikube
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
  - name: default-token-pmkp9
    secret:
      defaultMode: 420
      secretName: default-token-pmkp9
`

// Upstreams
const detailsUs = `
metadata:
  labels:
    app: details
    discovered_by: kubernetesplugin
  name: default-details-9080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: details
    serviceName: details
    serviceNamespace: default
    servicePort: 9080
---
metadata:
  labels:
    app: details
    discovered_by: kubernetesplugin
  name: default-details-v1-9080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: details
      version: v1
    serviceName: details
    serviceNamespace: default
    servicePort: 9080
`
const productPageUs = `
metadata:
  labels:
    app: productpage
    discovered_by: kubernetesplugin
  name: default-productpage-9080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: productpage
    serviceName: productpage
    serviceNamespace: default
    servicePort: 9080
---
metadata:
  labels:
    app: productpage
    discovered_by: kubernetesplugin
  name: default-productpage-v1-9080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: productpage
      version: v1
    serviceName: productpage
    serviceNamespace: default
    servicePort: 9080
`
const ratingsUs = `
metadata:
  labels:
    app: ratings
    discovered_by: kubernetesplugin
  name: default-ratings-9080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: ratings
    serviceName: ratings
    serviceNamespace: default
    servicePort: 9080
---
metadata:
  labels:
    app: ratings
    discovered_by: kubernetesplugin
  name: default-ratings-v1-9080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: ratings
      version: v1
    serviceName: ratings
    serviceNamespace: default
    servicePort: 9080
`
const reviewsV1Us = `
metadata:
  labels:
    app: reviews
    discovered_by: kubernetesplugin
  name: default-reviews-9080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: reviews
      version: v1
    serviceName: reviews
    serviceNamespace: default
    servicePort: 9080
`
const reviewsV2Us = `
metadata:
  labels:
    app: reviews
    discovered_by: kubernetesplugin
  name: default-reviews-v2-9080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: reviews
      version: v2
    serviceName: reviews-v2
    serviceNamespace: default
    servicePort: 9080
`
const reviewsV3Us = `
metadata:
  labels:
    app: reviews
    discovered_by: kubernetesplugin
  name: default-reviews-v3-9080
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      app: reviews
      version: v3
    serviceName: reviews-v3
    serviceNamespace: default
    servicePort: 9080
`

// Services
const detailsSvc = `
apiVersion: v1
kind: Service
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"details"},"name":"details","namespace":"default"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"details"}}}
  creationTimestamp: "2019-04-19T21:38:58Z"
  labels:
    app: details
  name: details
  namespace: default
  resourceVersion: "206918"
  selfLink: /api/v1/namespaces/default/services/details
  uid: 8a2a4e5b-62eb-11e9-989c-08002717ab34
spec:
  clusterIP: 10.105.142.32
  ports:
  - name: http
    port: 9080
    protocol: TCP
    targetPort: 9080
  selector:
    app: details
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
`
const productPageSvc = `
apiVersion: v1
kind: Service
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"productpage"},"name":"productpage","namespace":"default"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"productpage"}}}
  creationTimestamp: "2019-04-19T21:38:59Z"
  labels:
    app: productpage
  name: productpage
  namespace: default
  resourceVersion: "207005"
  selfLink: /api/v1/namespaces/default/services/productpage
  uid: 8aa9b286-62eb-11e9-989c-08002717ab34
spec:
  clusterIP: 10.97.192.20
  ports:
  - name: http
    port: 9080
    protocol: TCP
    targetPort: 9080
  selector:
    app: productpage
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
`
const ratingsSvc = `
apiVersion: v1
kind: Service
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"ratings"},"name":"ratings","namespace":"default"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"ratings"}}}
  creationTimestamp: "2019-04-19T21:38:58Z"
  labels:
    app: ratings
  name: ratings
  namespace: default
  resourceVersion: "206933"
  selfLink: /api/v1/namespaces/default/services/ratings
  uid: 8a347133-62eb-11e9-989c-08002717ab34
spec:
  clusterIP: 10.108.83.116
  ports:
  - name: http
    port: 9080
    protocol: TCP
    targetPort: 9080
  selector:
    app: ratings
  sessionAffinity: None
  type: ClusterIP
status:
  loadBalancer: {}
`
const reviewsV1Svc = `
apiVersion: v1
kind: Service
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"reviews"},"name":"reviews","namespace":"default"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"reviews","version":"v1"}}}
  creationTimestamp: "2019-04-19T21:38:58Z"
  labels:
    app: reviews
  name: reviews
  namespace: default
  resourceVersion: "206951"
  selfLink: /api/v1/namespaces/default/services/reviews
  uid: 8a53fe1a-62eb-11e9-989c-08002717ab34
spec:
  clusterIP: 10.104.251.7
  ports:
  - name: http
    port: 9080
    protocol: TCP
    targetPort: 9080
  selector:
    app: reviews
    version: v1
  sessionAffinity: None
  type: ClusterIP
`
const reviewsV2Svc = `
apiVersion: v1
kind: Service
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"reviews"},"name":"reviews-v2","namespace":"default"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"reviews","version":"v2"}}}
  creationTimestamp: "2019-04-19T21:38:59Z"
  labels:
    app: reviews
  name: reviews-v2
  namespace: default
  resourceVersion: "206964"
  selfLink: /api/v1/namespaces/default/services/reviews-v2
  uid: 8a67501d-62eb-11e9-989c-08002717ab34
spec:
  clusterIP: 10.97.100.183
  ports:
  - name: http
    port: 9080
    protocol: TCP
    targetPort: 9080
  selector:
    app: reviews
    version: v2
  sessionAffinity: None
  type: ClusterIP
`
const reviewsV3Svc = `
apiVersion: v1
kind: Service
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Service","metadata":{"annotations":{},"labels":{"app":"reviews"},"name":"reviews-v3","namespace":"default"},"spec":{"ports":[{"name":"http","port":9080}],"selector":{"app":"reviews","version":"v3"}}}
  creationTimestamp: "2019-04-19T21:38:59Z"
  labels:
    app: reviews
  name: reviews-v3
  namespace: default
  resourceVersion: "206978"
  selfLink: /api/v1/namespaces/default/services/reviews-v3
  uid: 8a7b671b-62eb-11e9-989c-08002717ab34
spec:
  clusterIP: 10.111.124.40
  ports:
  - name: http
    port: 9080
    protocol: TCP
    targetPort: 9080
  selector:
    app: reviews
    version: v3
  sessionAffinity: None
  type: ClusterIP
`
