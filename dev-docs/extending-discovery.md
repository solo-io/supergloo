# Working with Gloo Mesh Discovery APIs

If you want to write a custom Kubernetes controller for Gloo Mesh created objects like `workloads`, `traffictargets` etc, you can use the Gloo Mesh APIs to do so.

We also welcome PRs for extending the built-in Gloo Mesh discovery controller from the community.

## Creating necessary K8S RBAC permissions

For your application to use Gloo Mesh discovery APIs, you need to create a service account and bind it to a cluster role as shown below:

    cat <<EOF | kubectl apply -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
    name: gloo-mesh-sa
    namespace: default
    ---
    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
    creationTimestamp: null
    name: gloo-mesh-sa
    namespace: default
    rules:
    - apiGroups: ["discovery.gloo-mesh.solo.io"]
        resources: ["traffictargets", "workloads"]
        verbs: ["create", "get", "list", "watch", "update", "delete", "patch"]
    ---
    kind: ClusterRoleBinding
    apiVersion: rbac.authorization.k8s.io/v1
    metadata:
    name: gloo-mesh-sa
    namespace: default
    subjects:
    - kind: ServiceAccount
        name: gloo-mesh-sa
        namespace: default
    roleRef:
    kind: ClusterRole
    name: gloo-mesh-sa
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
              serviceAccountName: gloo-mesh-sa
              ...


For more information on what the RBAC objects actually do, refer the documentation - [https://kubernetes.io/docs/reference/access-authn-authz/rbac/](https://kubernetes.io/docs/reference/access-authn-authz/rbac/) 


## Working with Gloo Mesh resources

If you want to create a new deployment level abstraction (i.e., workloads) or a service level abstraction (i.e. traffictargets) manually and don't want Gloo Mesh to garbage collect it during the reconciliation loop, you need to exclude the `owner.discovery.mesh.gloo.solo.io: gloo-mesh` label from the metadata when creating the resources.

The label `owner.discovery.mesh.gloo.solo.io: gloo-mesh` signals to the discovery component of Gloo Mesh that it should manage that resource.

 The only metadata information which you should provide in your CRD manifest is the following:
 
     metadata:
       labels:
         cluster.discovery.mesh.gloo.solo.io: cluster1
       name: helloworld-test-default-cluster1
       namespace: gloo-mesh
     ...
