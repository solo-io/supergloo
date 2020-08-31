#!/bin/bash -ex

cluster=$0

K="kubectl --context=kind-${cluster}"

echo "installing osm to ${cluster}..."
ROLLOUT="${K} rollout status deployment --timeout 300s"

# install in permissive mode for testing
osm install

${ROLLOUT} -n osm-system osm-controller

for i in bookstore bookthief bookwarehouse bookbuyer; do ${K} create ns $i; done

for i in bookstore bookthief bookwarehouse bookbuyer; do osm namespace add $i; done

${K} apply -f - <<EOF
##################################################################################################
# Bookbuyer service
##################################################################################################

# Deploy bookbuyer Namespace
apiVersion: v1
kind: Namespace
metadata:
  name: bookbuyer
---
# Deploy bookbuyer Service Account
apiVersion: v1
kind: ServiceAccount
metadata:
  name: bookbuyer
  namespace: bookbuyer
---
# Deploy bookbuyer Service
apiVersion: v1
kind: Service
metadata:
  name: bookbuyer
  namespace: bookbuyer
  labels:
    app: bookbuyer
spec:
  ports:
  - port: 9999
    name: dummy-unused-port
  selector:
    app: bookbuyer
---
# Deploy bookbuyer Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bookbuyer
  namespace: bookbuyer
spec:
  replicas: 1
  selector:
    matchLabels:
      app: bookbuyer
      version: v1
  template:
    metadata:
      labels:
        app: bookbuyer
        version: v1
    spec:
      serviceAccountName: bookbuyer
      containers:
      - name: bookbuyer
        image: openservicemesh/bookbuyer:v0.2.0
        imagePullPolicy: Always
        command: ["/bookbuyer"]
        env:
        - name: "BOOKSTORE_NAMESPACE"
          value: bookstore
        - name: "BOOKSTORE_SVC"
          value: bookstore
---

##################################################################################################
# Bookstore v1 service
##################################################################################################

# Deploy bookstore Namespace
apiVersion: v1
kind: Namespace
metadata:
  name: bookstore
---
# Create a top level service just for the bookstore domain
apiVersion: v1
kind: Service
metadata:
  name: bookstore
  namespace: bookstore
  labels:
    app: bookstore
  annotations:
    # annotation for traffic target discovery
    "discovery.smh.solo.io/enabled": "true"
spec:
  ports:
  - port: 80
    name: bookstore-port
  selector:
    app: bookstore
---
# Deploy bookstore-v1 Service
apiVersion: v1
kind: Service
metadata:
  name: bookstore-v1
  namespace: bookstore
  labels:
    app: bookstore
    version: v1
spec:
  ports:
  - port: 80
    name: bookstore-port
  selector:
    app: bookstore-v1
---
# Deploy bookstore-v1 Service Account
apiVersion: v1
kind: ServiceAccount
metadata:
  name: bookstore-v1
  namespace: bookstore
---
# Deploy bookstore-v1 Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bookstore-v1
  namespace: bookstore
spec:
  replicas: 1
  selector:
    matchLabels:
      app: bookstore-v1
  template:
    metadata:
      labels:
        app: bookstore-v1
    spec:
      serviceAccountName: bookstore-v1
      containers:
      - name: bookstore
        image: openservicemesh/bookstore:v0.2.0
        imagePullPolicy: Always
        ports:
        - containerPort: 80
          name: web
        command: ["/bookstore"]
        args: ["--path", "./", "--port", "80"]
        env:
        - name: BOOKWAREHOUSE_NAMESPACE
          value: bookwarehouse
        - name: IDENTITY
          value: bookstore-v1
---

##################################################################################################
# Booktheif service
##################################################################################################

# Deploy bookthief Namespace
apiVersion: v1
kind: Namespace
metadata:
  name: bookthief
---
# Deploy bookthief ServiceAccount
apiVersion: v1
kind: ServiceAccount
metadata:
  name: bookthief
  namespace: bookthief
---
# Deploy bookthief Service
apiVersion: v1
kind: Service
metadata:
  name: bookthief
  namespace: bookthief
  labels:
    app: bookthief
spec:
  ports:
  - port: 9999
    name: dummy-unused-port
  selector:
    app: bookthief
---
# Deploy bookthief Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bookthief
  namespace: bookthief
spec:
  replicas: 1
  selector:
    matchLabels:
      app: bookthief
  template:
    metadata:
      labels:
        app: bookthief
        version: v1
    spec:
      serviceAccountName: bookthief
      containers:
      - name: bookthief
        image: openservicemesh/bookthief:v0.2.0
        imagePullPolicy: Always
        command: ["/bookthief"]
        env:
        - name: "BOOKSTORE_NAMESPACE"
          value: bookstore
        - name: "BOOKSTORE_SVC"
          value: bookstore
        - name: "BOOKTHIEF_EXPECTED_RESPONSE_CODE"
          value: "503"
      # Include curl container for e2e testing, allows sending traffic mediated by the proxy sidecar
      - name: curl
        image: curlimages/curl@sha256:aa45e9d93122a3cfdf8d7de272e2798ea63733eeee6d06bd2ee4f2f8c4027d7c
        command:
        - "sleep"
        - "10h"
      volumes:
      - name: tmp
        emptyDir: {}

---

##################################################################################################
# Bookwarehouse service
##################################################################################################

# Deploy bookwarehouse Namespace
apiVersion: v1
kind: Namespace
metadata:
  name: bookwarehouse
---
# Deploy bookwarehouse Service Account
apiVersion: v1
kind: ServiceAccount
metadata:
  name: bookwarehouse
  namespace: bookwarehouse
---
# Deploy bookwarehouse Service
apiVersion: v1
kind: Service
metadata:
  name: bookwarehouse
  namespace: bookwarehouse
  labels:
    app: bookwarehouse
spec:
  ports:
  - port: 80
    name: bookwarehouse-port
  selector:
    app: bookwarehouse
---
# Deploy bookwarehouse Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bookwarehouse
  namespace: bookwarehouse
spec:
  replicas: 1
  selector:
    matchLabels:
      app: bookwarehouse
  template:
    metadata:
      labels:
        app: bookwarehouse
        version: v1
    spec:
      serviceAccountName: bookwarehouse
      containers:
      - name: bookwarehouse
        image: openservicemesh/bookwarehouse:v0.2.0
        imagePullPolicy: Always
        command: ["/bookwarehouse"]

---

##################################################################################################
# Bookwarehouse service
##################################################################################################

# Deploy bookstore-v2 Service
apiVersion: v1
kind: Service
metadata:
  name: bookstore-v2
  namespace: bookstore
  labels:
    app: bookstore-v2
spec:
  ports:
  - port: 80
    name: bookstore-port
  selector:
    app: bookstore-v2
---
# Deploy bookstore-v2 Service Account
apiVersion: v1
kind: ServiceAccount
metadata:
  name: bookstore-v2
  namespace: bookstore
---
# Deploy bookstore-v2 Deployment
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bookstore-v2
  namespace: bookstore
spec:
  replicas: 1
  selector:
    matchLabels:
      app: bookstore-v2
  template:
    metadata:
      labels:
        app: bookstore-v2
    spec:
      serviceAccountName: bookstore-v2
      containers:
      - name: bookstore
        image: openservicemesh/bookstore:v0.2.0
        imagePullPolicy: Always
        ports:
        - containerPort: 80
          name: web
        command: ["/bookstore"]
        args: ["--path", "./", "--port", "80"]
        env:
        - name: BOOKWAREHOUSE_NAMESPACE
          value: bookwarehouse
        - name: IDENTITY
          value: bookstore-v2
---
EOF

${ROLLOUT} -n bookstore bookstore-v1
${ROLLOUT} -n bookstore bookstore-v2
${ROLLOUT} -n bookthief bookthief
${ROLLOUT} -n bookwarehouse bookwarehouse
${ROLLOUT} -n bookbuyer bookbuyer