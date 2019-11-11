# mesh-projects

mesh-projects is a multi-purpose repo aimed at handling all multi-cluster, mesh related, operators.

All mesh-projects operators are located in the `services` folder. Current the 3 main projects are:

1) mesh-discovery
    * mesh-discovery is the operator which reads in resources from local and remote clusters, and tries to determine
    if meshes exist. If it determines that one does it will write out a mesh CRD in the local `writeNamespace`.
2) mesh-config
    * mesh-config is the mesh-projects policy operator. It operates on the rbac related sections of the mesh CRD and 
    creates mesh resources to implement the policy.
3) mesh-bridge
    * mesh-bridge is the mesh-projects cross cluster networking operator. It operates on the mesh-bridge crd and creates
    network bridges between different meshes.

#### Multi Cluster Setup

As mentioned earlier all of the above operators/services have multi-cluster features. Enabling these features is simple,
but does require a few steps depending on where the clusters are located. The current supported platforms are EKS and GKE.

* EKS

    EKS is the slightly easier option of the 2 as or right now. In order to enable access to a remote EKS cluster the 
    following steps have to be followed
    1) Create a secret with the aws credentials with which you would like to authenticate. This can be done with the 
    following command `kubectl create secret generic --from-file=credentials=~/.aws/credentials aws-cred`. The secret 
    must have the name `aws-cred`, and be in the same namespace as the pod or it will not be picked up.
    2) Once the credentials secret has been created the pod which needs the creds should be restarted. This isn't strictly
    necessary as the file may be hot-reloaded. But the fastest way if possible is to delete it.
    3) Create a kubeconfig with the cluster you wish to access as the current context. This can be done very easily using
    eksctl. `eksctl utils write-kubeconfig --region <cluster-region> --name <cluster-namne> --auto-kubeconfig`. This command 
    will save the kubeconfig into `~/.kube/eksctl/clusters/<cluster-name>`.
    4) Create a secret containing the above kubeconfig. This can be done using the following script:
    `go run ./hack/kube/remote-kube-config.go -f ~/.kube/eksctl/clusters/<cluster-name> <cluster-id> | k apply -f -`
    5) Now the multi cluster clients should pick up the kube config, and use it. Notice the `<cluster-id>` above. That value
    is how the multi-cluster clients ID the cluster internally. This value can be the same as the `<cluster-name>` or not.
 
 * GKE
 
    GKE follows similar steps to EKS but requires a couple extra steps.
    1) Save the credentials for the service account you wish to authenticate with to a local directory. This can be done
    by navigating to the iam/serviceaccounts section of the gcloud console, and creating a new json key. Once the creds 
    have been downloaded. They should be saved to `~/.google/<creds>.json`.
    2) As above, create a secret with the credentials of the service account.This can be done with the following command 
    `kubectl create secret generic --from-file=credentials=~/.google/<creds>.json gcloud-cred`. The secret must have 
    the name `gcloud-cred`, and be in the same namespace as the pod or it will not be picked up.
    3) Create a kubeconfig with the cluster you wish to access as the current context. This can be done using the gcloud
    command line tool. First create an empty file to house this kube config, for instance: `~/.kube/gcloud/clusters/<cluster-name>`.
    The run: `KUBECONFIG=~/.kube/gcloud/clusters/<cluster-name> gcloud container clusters get-credentials ...`. Fill in
    the rest with the cluster you wish to connect to.
    4) Create a secret containing the above kubeconfig. This can be done using the following script:
    `go run ./hack/kube/remote-kube-config.go -f ~/.kube/gcloud/clusters/<cluster-name> <cluster-id> | k apply -f -`
    5) Now the multi cluster clients should pick up the kube config, and use it. Notice the `<cluster-id>` above. That value
    is how the multi-cluster clients ID the cluster internally. This value can be the same as the `<cluster-name>` or not.
    