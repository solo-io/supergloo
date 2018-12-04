package utils

import (
	"k8s.io/client-go/rest"
)

func DeployColorTeller(cfg *rest.Config, namespace string) error {
	return DeployFromYaml(cfg, namespace, ColorTellerYaml)
}

const ColorTellerYaml = `# source: https://github.com/awslabs/aws-app-mesh-examples/tree/master/examples/apps/colorapp/kubernetes
apiVersion: v1
kind: Service
metadata:
  name: colorgateway
  labels:
    app: colorgateway
spec:
  ports:
  - port: 9080
    name: http
  selector:
    app: colorgateway
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: colorgateway
spec:
  replicas: 1
  selector:
    matchLabels:
      app: colorgateway
      version: v1
  template:
    metadata:
      labels:
        app: colorgateway
        version: v1
    spec:
      containers:
        - name: colorgateway
          image: soloio/appmesh-example-colorgateway:latest
          ports:
            - containerPort: 9080
          env:
            - name: "SERVER_PORT"
              value: "9080"
            - name: "COLOR_TELLER_ENDPOINT"
              value: "colorteller.default.svc.cluster.local:9080"
        - name: envoy
          image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
          securityContext:
            runAsUser: 1337
          env:
            - name: "APPMESH_VIRTUAL_NODE_UID"
              value: "mesh/defaultmesh/virtualNode/colorgateway-vn"
            - name: "ENVOY_LOG_LEVEL"
              value: "debug"
            - name: "AWS_REGION"
              value: "us-west-2"
      initContainers:
        - name: proxyinit
          image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager
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
              value: "9080"
            - name: "APPMESH_EGRESS_IGNORED_IP"
              value: "169.254.169.254"
---
apiVersion: v1
kind: Service
metadata:
  name: colorteller
  labels:
    app: colorteller
spec:
  ports:
  - port: 9080
    name: http
  selector:
    app: colorteller-none
---

# black
apiVersion: v1
kind: Service
metadata:
  name: colorteller-black
  labels:
    app: colorteller
    version: black
spec:
  ports:
  - port: 9080
    name: http
  selector:
    app: colorteller
    version: black
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: colorteller-black
spec:
  replicas: 1
  selector:
    matchLabels:
      app: colorteller
      version: black
  template:
    metadata:
      labels:
        app: colorteller
        version: black
    spec:
      containers:
        - name: colorteller
          image: soloio/appmesh-example-colorteller:latest
          ports:
            - containerPort: 9080
          env:
            - name: "SERVER_PORT"
              value: "9080"
            - name: "COLOR"
              value: "black"
        - name: envoy
          image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
          securityContext:
            runAsUser: 1337
          env:
            - name: "APPMESH_VIRTUAL_NODE_UID"
              value: "mesh/defaultmesh/virtualNode/colorteller-black-vn"
            - name: "ENVOY_LOG_LEVEL"
              value: "debug"
            - name: "AWS_REGION"
              value: "us-west-2"
      initContainers:
        - name: proxyinit
          image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager
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
              value: "9080"
            - name: "APPMESH_EGRESS_IGNORED_IP"
              value: "169.254.169.254"
---

# blue
apiVersion: v1
kind: Service
metadata:
  name: colorteller-blue
  labels:
    app: colorteller
    version: blue
spec:
  ports:
  - port: 9080
    name: http
  selector:
    app: colorteller
    version: blue
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: colorteller-blue
spec:
  replicas: 1
  selector:
    matchLabels:
      app: colorteller
      version: blue
  template:
    metadata:
      labels:
        app: colorteller
        version: blue
    spec:
      containers:
        - name: colorteller
          image: soloio/appmesh-example-colorteller:latest
          ports:
            - containerPort: 9080
          env:
            - name: "SERVER_PORT"
              value: "9080"
            - name: "COLOR"
              value: "blue"
        - name: envoy
          image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
          securityContext:
            runAsUser: 1337
          env:
            - name: "APPMESH_VIRTUAL_NODE_UID"
              value: "mesh/defaultmesh/virtualNode/colorteller-blue-vn"
            - name: "ENVOY_LOG_LEVEL"
              value: "debug"
            - name: "AWS_REGION"
              value: "us-west-2"
      initContainers:
        - name: proxyinit
          image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager
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
              value: "9080"
            - name: "APPMESH_EGRESS_IGNORED_IP"
              value: "169.254.169.254"
---

# red
apiVersion: v1
kind: Service
metadata:
  name: colorteller-red
  labels:
    app: colorteller
    version: red
spec:
  ports:
  - port: 9080
    name: http
  selector:
    app: colorteller
    version: red
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: colorteller-red
spec:
  replicas: 1
  selector:
    matchLabels:
      app: colorteller
      version: red
  template:
    metadata:
      labels:
        app: colorteller
        version: red
    spec:
      containers:
        - name: colorteller
          image: soloio/appmesh-example-colorteller:latest
          ports:
            - containerPort: 9080
          env:
            - name: "SERVER_PORT"
              value: "9080"
            - name: "COLOR"
              value: "red"
        - name: envoy
          image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-envoy:v1.8.0.2-beta
          securityContext:
            runAsUser: 1337
          env:
            - name: "APPMESH_VIRTUAL_NODE_UID"
              value: "mesh/defaultmesh/virtualNode/colorteller-red-vn"
            - name: "ENVOY_LOG_LEVEL"
              value: "debug"
            - name: "AWS_REGION"
              value: "us-west-2"
      initContainers:
        - name: proxyinit
          image: 111345817488.dkr.ecr.us-west-2.amazonaws.com/aws-appmesh-proxy-route-manager
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
              value: "9080"
            - name: "APPMESH_EGRESS_IGNORED_IP"
              value: "169.254.169.254"
---

`
