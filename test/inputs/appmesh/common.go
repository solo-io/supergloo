package appmesh

const KubernetesPods = `
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-01T18:52:21Z"
  generateName: coredns-fb8b8dccf-
  labels:
    k8s-app: kube-dns
    pod-template-hash: fb8b8dccf
  name: coredns-fb8b8dccf-5cdzc
  namespace: kube-system
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: coredns-fb8b8dccf
    uid: 48169fe8-54af-11e9-bf69-08002717ab34
  resourceVersion: "171603"
  selfLink: /api/v1/namespaces/kube-system/pods/coredns-fb8b8dccf-5cdzc
  uid: 481bd73f-54af-11e9-bf69-08002717ab34
spec:
  containers:
  - args:
    - -conf
    - /etc/coredns/Corefile
    image: k8s.gcr.io/coredns:1.3.1
    imagePullPolicy: IfNotPresent
    livenessProbe:
      failureThreshold: 5
      httpGet:
        path: /health
        port: 8080
        scheme: HTTP
      initialDelaySeconds: 60
      periodSeconds: 10
      successThreshold: 1
      timeoutSeconds: 5
    name: coredns
    ports:
    - containerPort: 53
      name: dns
      protocol: UDP
    - containerPort: 53
      name: dns-tcp
      protocol: TCP
    - containerPort: 9153
      name: metrics
      protocol: TCP
    readinessProbe:
      failureThreshold: 3
      httpGet:
        path: /health
        port: 8080
        scheme: HTTP
      periodSeconds: 10
      successThreshold: 1
      timeoutSeconds: 1
    resources:
      limits:
        memory: 170Mi
      requests:
        cpu: 100m
        memory: 70Mi
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        add:
        - NET_BIND_SERVICE
        drop:
        - all
      procMount: Default
      readOnlyRootFilesystem: true
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /etc/coredns
      name: config-volume
      readOnly: true
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: coredns-token-w82b2
      readOnly: true
  dnsPolicy: Default
  enableServiceLinks: true
  nodeName: minikube
  nodeSelector:
    beta.kubernetes.io/os: linux
  priority: 2000000000
  priorityClassName: system-cluster-critical
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: coredns
  serviceAccountName: coredns
  terminationGracePeriodSeconds: 30
  tolerations:
  - key: CriticalAddonsOnly
    operator: Exists
  - effect: NoSchedule
    key: node-role.kubernetes.io/master
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  volumes:
  - configMap:
      defaultMode: 420
      items:
      - key: Corefile
        path: Corefile
      name: coredns
    name: config-volume
  - name: coredns-token-w82b2
    secret:
      defaultMode: 420
      secretName: coredns-token-w82b2
---
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-01T18:52:21Z"
  generateName: coredns-fb8b8dccf-
  labels:
    k8s-app: kube-dns
    pod-template-hash: fb8b8dccf
  name: coredns-fb8b8dccf-tthjt
  namespace: kube-system
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: ReplicaSet
    name: coredns-fb8b8dccf
    uid: 48169fe8-54af-11e9-bf69-08002717ab34
  resourceVersion: "171617"
  selfLink: /api/v1/namespaces/kube-system/pods/coredns-fb8b8dccf-tthjt
  uid: 481b315b-54af-11e9-bf69-08002717ab34
spec:
  containers:
  - args:
    - -conf
    - /etc/coredns/Corefile
    image: k8s.gcr.io/coredns:1.3.1
    imagePullPolicy: IfNotPresent
    livenessProbe:
      failureThreshold: 5
      httpGet:
        path: /health
        port: 8080
        scheme: HTTP
      initialDelaySeconds: 60
      periodSeconds: 10
      successThreshold: 1
      timeoutSeconds: 5
    name: coredns
    ports:
    - containerPort: 53
      name: dns
      protocol: UDP
    - containerPort: 53
      name: dns-tcp
      protocol: TCP
    - containerPort: 9153
      name: metrics
      protocol: TCP
    readinessProbe:
      failureThreshold: 3
      httpGet:
        path: /health
        port: 8080
        scheme: HTTP
      periodSeconds: 10
      successThreshold: 1
      timeoutSeconds: 1
    resources:
      limits:
        memory: 170Mi
      requests:
        cpu: 100m
        memory: 70Mi
    securityContext:
      allowPrivilegeEscalation: false
      capabilities:
        add:
        - NET_BIND_SERVICE
        drop:
        - all
      procMount: Default
      readOnlyRootFilesystem: true
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /etc/coredns
      name: config-volume
      readOnly: true
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: coredns-token-w82b2
      readOnly: true
  dnsPolicy: Default
  enableServiceLinks: true
  nodeName: minikube
  nodeSelector:
    beta.kubernetes.io/os: linux
  priority: 2000000000
  priorityClassName: system-cluster-critical
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: coredns
  serviceAccountName: coredns
  terminationGracePeriodSeconds: 30
  tolerations:
  - key: CriticalAddonsOnly
    operator: Exists
  - effect: NoSchedule
    key: node-role.kubernetes.io/master
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  volumes:
  - configMap:
      defaultMode: 420
      items:
      - key: Corefile
        path: Corefile
      name: coredns
    name: config-volume
  - name: coredns-token-w82b2
    secret:
      defaultMode: 420
      secretName: coredns-token-w82b2
---
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kubernetes.io/config.hash: 18c827a17f0a6b507c2029890cd786ad
    kubernetes.io/config.mirror: 18c827a17f0a6b507c2029890cd786ad
    kubernetes.io/config.seen: "2019-04-01T18:52:05.820918558Z"
    kubernetes.io/config.source: file
  creationTimestamp: "2019-04-01T18:53:26Z"
  labels:
    component: etcd
    tier: control-plane
  name: etcd-minikube
  namespace: kube-system
  resourceVersion: "171398"
  selfLink: /api/v1/namespaces/kube-system/pods/etcd-minikube
  uid: 6eebdffa-54af-11e9-bf69-08002717ab34
spec:
  containers:
  - command:
    - etcd
    - --advertise-client-urls=https://192.168.99.100:2379
    - --cert-file=/var/lib/minikube/certs/etcd/server.crt
    - --client-cert-auth=true
    - --data-dir=/data/minikube
    - --initial-advertise-peer-urls=https://192.168.99.100:2380
    - --initial-cluster=minikube=https://192.168.99.100:2380
    - --key-file=/var/lib/minikube/certs/etcd/server.key
    - --listen-client-urls=https://127.0.0.1:2379,https://192.168.99.100:2379
    - --listen-peer-urls=https://192.168.99.100:2380
    - --name=minikube
    - --peer-cert-file=/var/lib/minikube/certs/etcd/peer.crt
    - --peer-client-cert-auth=true
    - --peer-key-file=/var/lib/minikube/certs/etcd/peer.key
    - --peer-trusted-ca-file=/var/lib/minikube/certs/etcd/ca.crt
    - --snapshot-count=10000
    - --trusted-ca-file=/var/lib/minikube/certs/etcd/ca.crt
    image: k8s.gcr.io/etcd:3.3.10
    imagePullPolicy: IfNotPresent
    livenessProbe:
      exec:
        command:
        - /bin/sh
        - -ec
        - ETCDCTL_API=3 etcdctl --endpoints=https://[127.0.0.1]:2379 --cacert=/var/lib/minikube/certs//etcd/ca.crt
          --cert=/var/lib/minikube/certs//etcd/healthcheck-client.crt --key=/var/lib/minikube/certs//etcd/healthcheck-client.key
          get foo
      failureThreshold: 8
      initialDelaySeconds: 15
      periodSeconds: 10
      successThreshold: 1
      timeoutSeconds: 15
    name: etcd
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /data/minikube
      name: etcd-data
    - mountPath: /var/lib/minikube/certs//etcd
      name: etcd-certs
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  hostNetwork: true
  nodeName: minikube
  priority: 2000000000
  priorityClassName: system-cluster-critical
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    operator: Exists
  volumes:
  - hostPath:
      path: /var/lib/minikube/certs//etcd
      type: DirectoryOrCreate
    name: etcd-certs
  - hostPath:
      path: /data/minikube
      type: DirectoryOrCreate
    name: etcd-data
---
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kubernetes.io/config.hash: 0abcb7a1f0c9c0ebc9ec348ffdfb220c
    kubernetes.io/config.mirror: 0abcb7a1f0c9c0ebc9ec348ffdfb220c
    kubernetes.io/config.seen: "2019-04-01T18:52:05.820914412Z"
    kubernetes.io/config.source: file
  creationTimestamp: "2019-04-01T18:53:14Z"
  labels:
    component: kube-addon-manager
    kubernetes.io/minikube-addons: addon-manager
    version: v9.0
  name: kube-addon-manager-minikube
  namespace: kube-system
  resourceVersion: "171393"
  selfLink: /api/v1/namespaces/kube-system/pods/kube-addon-manager-minikube
  uid: 67c58c0e-54af-11e9-bf69-08002717ab34
spec:
  containers:
  - env:
    - name: KUBECONFIG
      value: /var/lib/minikube/kubeconfig
    image: k8s.gcr.io/kube-addon-manager:v9.0
    imagePullPolicy: IfNotPresent
    name: kube-addon-manager
    resources:
      requests:
        cpu: 5m
        memory: 50Mi
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /etc/kubernetes/
      name: addons
      readOnly: true
    - mountPath: /var/lib/minikube/
      name: kubeconfig
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  hostNetwork: true
  nodeName: minikube
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    operator: Exists
  volumes:
  - hostPath:
      path: /etc/kubernetes/
      type: ""
    name: addons
  - hostPath:
      path: /var/lib/minikube/
      type: ""
    name: kubeconfig
---
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kubernetes.io/config.hash: 023cdc77988402bd2101e9dc50c78f18
    kubernetes.io/config.mirror: 023cdc77988402bd2101e9dc50c78f18
    kubernetes.io/config.seen: "2019-04-01T18:52:05.820919921Z"
    kubernetes.io/config.source: file
  creationTimestamp: "2019-04-01T18:53:33Z"
  labels:
    component: kube-apiserver
    tier: control-plane
  name: kube-apiserver-minikube
  namespace: kube-system
  resourceVersion: "171407"
  selfLink: /api/v1/namespaces/kube-system/pods/kube-apiserver-minikube
  uid: 73183a39-54af-11e9-bf69-08002717ab34
spec:
  containers:
  - command:
    - kube-apiserver
    - --advertise-address=192.168.99.100
    - --allow-privileged=true
    - --authorization-mode=Node,RBAC
    - --client-ca-file=/var/lib/minikube/certs/ca.crt
    - --enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota
    - --enable-bootstrap-token-auth=true
    - --etcd-cafile=/var/lib/minikube/certs/etcd/ca.crt
    - --etcd-certfile=/var/lib/minikube/certs/apiserver-etcd-client.crt
    - --etcd-keyfile=/var/lib/minikube/certs/apiserver-etcd-client.key
    - --etcd-servers=https://127.0.0.1:2379
    - --insecure-port=0
    - --kubelet-client-certificate=/var/lib/minikube/certs/apiserver-kubelet-client.crt
    - --kubelet-client-key=/var/lib/minikube/certs/apiserver-kubelet-client.key
    - --kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname
    - --proxy-client-cert-file=/var/lib/minikube/certs/front-proxy-client.crt
    - --proxy-client-key-file=/var/lib/minikube/certs/front-proxy-client.key
    - --requestheader-allowed-names=front-proxy-client
    - --requestheader-client-ca-file=/var/lib/minikube/certs/front-proxy-ca.crt
    - --requestheader-extra-headers-prefix=X-Remote-Extra-
    - --requestheader-group-headers=X-Remote-Group
    - --requestheader-username-headers=X-Remote-User
    - --secure-port=8443
    - --service-account-key-file=/var/lib/minikube/certs/sa.pub
    - --service-cluster-ip-range=10.96.0.0/12
    - --tls-cert-file=/var/lib/minikube/certs/apiserver.crt
    - --tls-private-key-file=/var/lib/minikube/certs/apiserver.key
    image: k8s.gcr.io/kube-apiserver:v1.14.0
    imagePullPolicy: IfNotPresent
    livenessProbe:
      failureThreshold: 8
      httpGet:
        host: 192.168.99.100
        path: /healthz
        port: 8443
        scheme: HTTPS
      initialDelaySeconds: 15
      periodSeconds: 10
      successThreshold: 1
      timeoutSeconds: 15
    name: kube-apiserver
    resources:
      requests:
        cpu: 250m
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /etc/ssl/certs
      name: ca-certs
      readOnly: true
    - mountPath: /var/lib/minikube/certs/
      name: k8s-certs
      readOnly: true
    - mountPath: /usr/share/ca-certificates
      name: usr-share-ca-certificates
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  hostNetwork: true
  nodeName: minikube
  priority: 2000000000
  priorityClassName: system-cluster-critical
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    operator: Exists
  volumes:
  - hostPath:
      path: /etc/ssl/certs
      type: DirectoryOrCreate
    name: ca-certs
  - hostPath:
      path: /var/lib/minikube/certs/
      type: DirectoryOrCreate
    name: k8s-certs
  - hostPath:
      path: /usr/share/ca-certificates
      type: DirectoryOrCreate
    name: usr-share-ca-certificates
---
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kubernetes.io/config.hash: 2899d819dcdb72186fb15d30a0cc5a71
    kubernetes.io/config.mirror: 2899d819dcdb72186fb15d30a0cc5a71
    kubernetes.io/config.seen: "2019-04-01T18:52:05.82092107Z"
    kubernetes.io/config.source: file
  creationTimestamp: "2019-04-01T18:53:33Z"
  labels:
    component: kube-controller-manager
    tier: control-plane
  name: kube-controller-manager-minikube
  namespace: kube-system
  resourceVersion: "171449"
  selfLink: /api/v1/namespaces/kube-system/pods/kube-controller-manager-minikube
  uid: 7317fa0c-54af-11e9-bf69-08002717ab34
spec:
  containers:
  - command:
    - kube-controller-manager
    - --authentication-kubeconfig=/etc/kubernetes/controller-manager.conf
    - --authorization-kubeconfig=/etc/kubernetes/controller-manager.conf
    - --bind-address=127.0.0.1
    - --client-ca-file=/var/lib/minikube/certs/ca.crt
    - --cluster-signing-cert-file=/var/lib/minikube/certs/ca.crt
    - --cluster-signing-key-file=/var/lib/minikube/certs/ca.key
    - --controllers=*,bootstrapsigner,tokencleaner
    - --kubeconfig=/etc/kubernetes/controller-manager.conf
    - --leader-elect=true
    - --requestheader-client-ca-file=/var/lib/minikube/certs/front-proxy-ca.crt
    - --root-ca-file=/var/lib/minikube/certs/ca.crt
    - --service-account-private-key-file=/var/lib/minikube/certs/sa.key
    - --use-service-account-credentials=true
    image: k8s.gcr.io/kube-controller-manager:v1.14.0
    imagePullPolicy: IfNotPresent
    livenessProbe:
      failureThreshold: 8
      httpGet:
        host: 127.0.0.1
        path: /healthz
        port: 10252
        scheme: HTTP
      initialDelaySeconds: 15
      periodSeconds: 10
      successThreshold: 1
      timeoutSeconds: 15
    name: kube-controller-manager
    resources:
      requests:
        cpu: 200m
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /etc/ssl/certs
      name: ca-certs
      readOnly: true
    - mountPath: /var/lib/minikube/certs/
      name: k8s-certs
      readOnly: true
    - mountPath: /etc/kubernetes/controller-manager.conf
      name: kubeconfig
      readOnly: true
    - mountPath: /usr/share/ca-certificates
      name: usr-share-ca-certificates
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  hostNetwork: true
  nodeName: minikube
  priority: 2000000000
  priorityClassName: system-cluster-critical
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    operator: Exists
  volumes:
  - hostPath:
      path: /etc/ssl/certs
      type: DirectoryOrCreate
    name: ca-certs
  - hostPath:
      path: /var/lib/minikube/certs/
      type: DirectoryOrCreate
    name: k8s-certs
  - hostPath:
      path: /etc/kubernetes/controller-manager.conf
      type: FileOrCreate
    name: kubeconfig
  - hostPath:
      path: /usr/share/ca-certificates
      type: DirectoryOrCreate
    name: usr-share-ca-certificates
---
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2019-04-02T01:16:26Z"
  generateName: kube-proxy-
  labels:
    controller-revision-hash: b7775b676
    k8s-app: kube-proxy
    pod-template-generation: "1"
  name: kube-proxy-dt5lc
  namespace: kube-system
  ownerReferences:
  - apiVersion: apps/v1
    blockOwnerDeletion: true
    controller: true
    kind: DaemonSet
    name: kube-proxy
    uid: 452ee0a4-54af-11e9-bf69-08002717ab34
  resourceVersion: "171446"
  selfLink: /api/v1/namespaces/kube-system/pods/kube-proxy-dt5lc
  uid: efbb0387-54e4-11e9-a9b8-08002717ab34
spec:
  affinity:
    nodeAffinity:
      requiredDuringSchedulingIgnoredDuringExecution:
        nodeSelectorTerms:
        - matchFields:
          - key: metadata.name
            operator: In
            values:
            - minikube
  containers:
  - command:
    - /usr/local/bin/kube-proxy
    - --config=/var/lib/kube-proxy/config.conf
    - --hostname-override=$(NODE_NAME)
    env:
    - name: NODE_NAME
      valueFrom:
        fieldRef:
          apiVersion: v1
          fieldPath: spec.nodeName
    image: k8s.gcr.io/kube-proxy:v1.14.0
    imagePullPolicy: IfNotPresent
    name: kube-proxy
    resources: {}
    securityContext:
      privileged: true
      procMount: Default
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/lib/kube-proxy
      name: kube-proxy
    - mountPath: /run/xtables.lock
      name: xtables-lock
    - mountPath: /lib/modules
      name: lib-modules
      readOnly: true
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: kube-proxy-token-62l9p
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  hostNetwork: true
  nodeName: minikube
  priority: 2000001000
  priorityClassName: system-node-critical
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: kube-proxy
  serviceAccountName: kube-proxy
  terminationGracePeriodSeconds: 30
  tolerations:
  - key: CriticalAddonsOnly
    operator: Exists
  - operator: Exists
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
  - effect: NoSchedule
    key: node.kubernetes.io/disk-pressure
    operator: Exists
  - effect: NoSchedule
    key: node.kubernetes.io/memory-pressure
    operator: Exists
  - effect: NoSchedule
    key: node.kubernetes.io/pid-pressure
    operator: Exists
  - effect: NoSchedule
    key: node.kubernetes.io/unschedulable
    operator: Exists
  - effect: NoSchedule
    key: node.kubernetes.io/network-unavailable
    operator: Exists
  volumes:
  - configMap:
      defaultMode: 420
      name: kube-proxy
    name: kube-proxy
  - hostPath:
      path: /run/xtables.lock
      type: FileOrCreate
    name: xtables-lock
  - hostPath:
      path: /lib/modules
      type: ""
    name: lib-modules
  - name: kube-proxy-token-62l9p
    secret:
      defaultMode: 420
      secretName: kube-proxy-token-62l9p
---
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kubernetes.io/config.hash: 58272442e226c838b193bbba4c44091e
    kubernetes.io/config.mirror: 58272442e226c838b193bbba4c44091e
    kubernetes.io/config.seen: "2019-04-01T18:52:05.820922001Z"
    kubernetes.io/config.source: file
  creationTimestamp: "2019-04-01T18:53:20Z"
  labels:
    component: kube-scheduler
    tier: control-plane
  name: kube-scheduler-minikube
  namespace: kube-system
  resourceVersion: "171410"
  selfLink: /api/v1/namespaces/kube-system/pods/kube-scheduler-minikube
  uid: 6b585359-54af-11e9-bf69-08002717ab34
spec:
  containers:
  - command:
    - kube-scheduler
    - --bind-address=127.0.0.1
    - --kubeconfig=/etc/kubernetes/scheduler.conf
    - --leader-elect=true
    image: k8s.gcr.io/kube-scheduler:v1.14.0
    imagePullPolicy: IfNotPresent
    livenessProbe:
      failureThreshold: 8
      httpGet:
        host: 127.0.0.1
        path: /healthz
        port: 10251
        scheme: HTTP
      initialDelaySeconds: 15
      periodSeconds: 10
      successThreshold: 1
      timeoutSeconds: 15
    name: kube-scheduler
    resources:
      requests:
        cpu: 100m
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /etc/kubernetes/scheduler.conf
      name: kubeconfig
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  hostNetwork: true
  nodeName: minikube
  priority: 2000000000
  priorityClassName: system-cluster-critical
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    operator: Exists
  volumes:
  - hostPath:
      path: /etc/kubernetes/scheduler.conf
      type: FileOrCreate
    name: kubeconfig
---
apiVersion: v1
kind: Pod
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"Pod","metadata":{"annotations":{},"labels":{"addonmanager.kubernetes.io/mode":"Reconcile","integration-test":"storage-provisioner"},"name":"storage-provisioner","namespace":"kube-system"},"spec":{"containers":[{"command":["/storage-provisioner"],"image":"gcr.io/k8s-minikube/storage-provisioner:v1.8.1","imagePullPolicy":"IfNotPresent","name":"storage-provisioner","volumeMounts":[{"mountPath":"/tmp","name":"tmp"}]}],"hostNetwork":true,"serviceAccountName":"storage-provisioner","volumes":[{"hostPath":{"path":"/tmp","type":"Directory"},"name":"tmp"}]}}
  creationTimestamp: "2019-04-01T18:52:24Z"
  labels:
    addonmanager.kubernetes.io/mode: Reconcile
    integration-test: storage-provisioner
  name: storage-provisioner
  namespace: kube-system
  resourceVersion: "171583"
  selfLink: /api/v1/namespaces/kube-system/pods/storage-provisioner
  uid: 4978cb7c-54af-11e9-bf69-08002717ab34
spec:
  containers:
  - command:
    - /storage-provisioner
    image: gcr.io/k8s-minikube/storage-provisioner:v1.8.1
    imagePullPolicy: IfNotPresent
    name: storage-provisioner
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /tmp
      name: tmp
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: storage-provisioner-token-5cxwh
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  hostNetwork: true
  nodeName: minikube
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: storage-provisioner
  serviceAccountName: storage-provisioner
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
  - hostPath:
      path: /tmp
      type: Directory
    name: tmp
  - name: storage-provisioner-token-5cxwh
    secret:
      defaultMode: 420
      secretName: storage-provisioner-token-5cxwh
`

const KubernetesUpstreams = `
metadata:
  labels:
    component: apiserver
    discovered_by: kubernetesplugin
    provider: kubernetes
  name: default-kubernetes-443
  namespace: gloo-system
upstreamSpec:
  kube:
    serviceName: kubernetes
    serviceNamespace: default
    servicePort: 443
---
metadata:
  labels:
    discovered_by: kubernetesplugin
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: KubeDNS
  name: kube-system-kube-dns-53
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      k8s-app: kube-dns
    serviceName: kube-dns
    serviceNamespace: kube-system
    servicePort: 53
---
metadata:
  labels:
    discovered_by: kubernetesplugin
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: KubeDNS
  name: kube-system-kube-dns-9153
  namespace: gloo-system
upstreamSpec:
  kube:
    selector:
      k8s-app: kube-dns
    serviceName: kube-dns
    serviceNamespace: kube-system
    servicePort: 9153
`

const KubernetesServices = `
apiVersion: v1
kind: Service
metadata:
  creationTimestamp: "2019-04-01T18:52:15Z"
  labels:
    component: apiserver
    provider: kubernetes
  name: kubernetes
  namespace: default
  resourceVersion: "155"
  selfLink: /api/v1/namespaces/default/services/kubernetes
  uid: 44248e0d-54af-11e9-bf69-08002717ab34
spec:
  clusterIP: 10.96.0.1
  ports:
  - name: https
    port: 443
    protocol: TCP
    targetPort: 8443
  sessionAffinity: None
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  annotations:
    prometheus.io/port: "9153"
    prometheus.io/scrape: "true"
  creationTimestamp: "2019-04-01T18:52:16Z"
  labels:
    k8s-app: kube-dns
    kubernetes.io/cluster-service: "true"
    kubernetes.io/name: KubeDNS
  name: kube-dns
  namespace: kube-system
  resourceVersion: "212"
  selfLink: /api/v1/namespaces/kube-system/services/kube-dns
  uid: 4500251d-54af-11e9-bf69-08002717ab34
spec:
  clusterIP: 10.96.0.10
  ports:
  - name: dns
    port: 53
    protocol: UDP
    targetPort: 53
  - name: dns-tcp
    port: 53
    protocol: TCP
    targetPort: 53
  - name: metrics
    port: 9153
    protocol: TCP
    targetPort: 9153
  selector:
    k8s-app: kube-dns
  sessionAffinity: None
  type: ClusterIP
`
