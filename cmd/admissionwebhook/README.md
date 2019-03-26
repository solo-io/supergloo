# Sidecar injector webhook
This executable implements a `MutationAdmissionWebhook` server that injects pods with AWS App Mesh sidecar proxies. 
The server is called by the Kubernetes API server each time a pod creation event occurs. The server determines 
whether the to-be-created pod need to be patched with the sidecar.

## Detailed flow
Following are the steps the server executes when it receives a `AdmissionReview` request for a pod creation.
1. List all the `Mesh` CRDs in the cluster and select the ones that:
    * are of type matching AWS App Mesh, and
    * have the `EnableAutoInject` flag set to true
2. For each of the selected meshes, check if the pod matches the `InjectionSelector` specified for the given mesh:
    1. if it does not match any mesh, admit the pod without patching it
    2. if it matches multiple meshes return an error, as multiple injection is currently not supported
    3. if it matches a mesh, continue
3. Retrieve the `SidecarPatchConfigMap` specified in the mesh CRD. This config map contains the template for the patch that will be applied to the pod. 
A missing or non-existing `SidecarPatchConfigMap` when `EnableAutoInject` is true will cause an error to be returned.
4. Retrieve the `VirtualNodeLabel` from the mesh CRD. The webhook will look for this label on the pod and use its value as 
the name of the [Virtual Node](https://docs.aws.amazon.com/app-mesh/latest/userguide/virtual_nodes.html) that the pod is associated with. 
The Virtual Node does not have to exist at this point. A missing `VirtualNodeLabel` when `EnableAutoInject` is true will cause an error to be returned.
5. The patch template is rendered with the following values:
    * `MeshName`: `name` of the mesh CRD, used to build the `APPMESH_VIRTUAL_NODE_NAME` env that is set on the sidecar proxy container.
    * `VirtualNodeName`: value of the pod label with key equal to `VirtualNodeLabel`. Also used to build the `APPMESH_VIRTUAL_NODE_NAME` env.
    * `AwsRegion`: `Region` of the `Mesh` CRD; indicates the AWs region where the control plane for the mesh is located. 
    * `AppPort`:  the `containerPort` of the container in the pod. Will be used as value for the `APPMESH_APP_PORTS` env that is set 
    on the `InitContainer` that will be added to the pod specification.
6. Convert the rendered patch to the [JSONPatch](https://tools.ietf.org/html/rfc6902) format expected by the Kubernetes API server.
7. Set the resulting `JSONPatch` on the response and return. The Kubernetes API server will apply the patch to the original pod spec.

    
## Limitations
The sidecar injector webhook currently has the following limitations:
* Pods must match at most one `Mesh`
* Pods must have at most one container
* The container in the input pod must specify a `containerPort`
* `InjectionSelector` must be of type `Label` or `Namespace` (`Upstreams` is currently not supported)
* AWS App Mesh is currently available only in the following AWS regions:
    * us-west-2
    * us-east-1
    * us-east-2
    * eu-west-1