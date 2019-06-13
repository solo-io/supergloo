# MeshDiscovery Design

MeshDiscovery attempts to read data from different sources to discover (via heuristics) the presence of control and data plane
components in the target cluster (by default, where the `meshdiscovery` pod runs) in order to create `supergloo.solo.io/v1.Mesh` custom resources.

MeshDiscovery currently supports the following types of meshes:

- Istio

    To detect Istio, MeshDiscovery watches Kubernetes deployments. When it finds a deployment with an image known to be 
    used for Istio's Pilot (the server distributing config to the Istio proxies),   
    it then attempts to discover configuration for that pilot instance, including preexisting configuration for mTLS as well as the presence of 
    the Istio sidecar injector.
    
    Once pilot instances are discovered, MeshDiscovery then searches through the specs for each pod in the cluster for the 
    presence of the Istio sidecar proxy. When the proxy is recognized (based on the container name `istio-proxy`), the 
    pilot instance it's talking to is inferred from arguments to the proxy binary provided in the pod spec. The Istio installation namespace  
    and version are determined from the pilot deployment namespace and container image tag, respectively.

- Linkerd

    Detecting Linkerd is similar to Istio; MeshDiscovery searches for the Linkerd Controller (which ) and determines version 
    based on the image tag. Injected pods are detected based on the presence  of the Linkerd sidecar.

- AWS AppMesh

    MeshDiscovery starts by determining if the target cluster is running on AWS EKS. It does so by searching for the presence of 
    `eks-node` pods that run in the `kube-system` namespace. The image name of these pods can be used to determine the 
    region in which the cluster is running.
    
    When these pods are detected, MeshDiscovery then scans for [AWS-Type Kubernetes secrets](https://gloo.solo.io/v1/github.com/solo-io/gloo/projects/gloo/api/v1/secret.proto.sk/#awssecret) which can be 
    [created with the SupeGloo command line](https://supergloo.solo.io/mesh/register-appmesh/#prep-your-environment). 
    
    These secrets are used to instantiate clients for the AWS AppMesh API. MeshDiscovery will then generate a 
    Mesh custom resource for each discovered Mesh instance in the AppMesh API.


When MeshDiscovery successfully detects a Mesh, its configuration, and its injected sidecars, the data is summarized 
and written as a Mesh custom resource.
