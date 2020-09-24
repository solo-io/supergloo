# Working with SMH Discovery APIs

If you want to write a custom Kubernetes controller for SMH created objects like `workloads`, `traffictargets` etc, you can use the SMH APIs to do so.

We also welcome PRs for extending the built-in SMH discovery controller from the community.

## Creating necessary K8S RBAC permissions

For your application to use SMH discovery APIs, you need to create a service account and bind it to a cluster role as shown below:

    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
    name: smh-sa
    namespace: default
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
    creationTimestamp: null
    name: smh-sa
    namespace: default
    rules:
    - apiGroups: ["discovery.smh.solo.io"]
        resources: ["traffictargets", "workloads"]
        verbs: ["create", "get", "list", "watch", "update", "delete", "patch"]
    ---
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
    name: smh-sa
    namespace: default
    subjects:
    - kind: ServiceAccount
        name: smh-sa
        namespace: default
    roleRef:
    kind: ClusterRole
    name: smh-sa
    apiGroup: rbac.authorization.k8s.io
    EOF

You can then use the serviceAccount in your deployment pod spec, like so:

    template:
        metadata:
          labels:
            app: helloworld
        spec:
          containers:
            - name: <your-container-image>
              serviceAccountName: smh-sa
              ...


For more information on what the RBAC objects actually do, refer the documentation - [https://kubernetes.io/docs/reference/access-authn-authz/rbac/](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) 


## Working with SMH resources

If you want to create a new deployment level abstraction (i.e., workloads) or a service level abstraction (i.e. traffictargets) manually and don't want SMH to garbage collect it during the reconciliation loop, you need to exclude the `owner.discovery.smh.solo.io: service-mesh-hub` label from the metadata when creating the resources.

The label `owner.discovery.smh.solo.io: service-mesh-hub` signals to the discovery component of SMH that it should manage that resource.

 The only metadata information which you should provide in your CRD manifest is the following:
 
     metadata:
       labels:
         cluster.discovery.smh.solo.io: cluster1
       name: helloworld-test-default-cluster1
       namespace: service-mesh-hub
     ...