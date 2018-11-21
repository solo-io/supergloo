## Installation

### Dependencies

- Go (1.11)
- VM Driver (tested with VirtualBox, KVM)
- Minikube (tested with 0.28.2-0.30.0)
- Helm 2 (tested with 2.11)
- Kubectl (tested with client version 1.12)

> For demo purposes, Supergloo is only supported on local Minikube environments. It will likely support other 
Kubernetes environments in the future. 

### Local Setup

#### 1. Create a new Kubernetes environment in Minikube

`minikube start --vm-driver=virtualbox --memory=8192 --cpus=4 --kubernetes-version=v1.10.0`

> Service meshes require a lot of resources. Swap out virtualbox for your preferred VM driver.

#### 2. Install supergloo cli and supergloo server

`make install-cli supergloo-server`

> When the CLI is first run, it will ensure that Helm is deployed and Supergloo's namespace is initialized.

#### 3. Start the supergloo server locally

`supergloo-server`

> This will stay running and print logs to the console. Open another tab to run the CLI

### Example Workflows

#### Install a new service mesh

Supergloo supports Istio, Consul, and Linkerd2. To install them with default configuration, run the following command:

`supergloo install -m {meshname} -n {namespace} -s`

`{meshname}` should be one of `consul`, `istio`, or `linkerd2`. `{namespace}` is a namespace where the mesh control plane
will be deployed. Supergloo will create this namespace if it doesn't already exist. 

For instance, to deploy `istio` into the `istio-system` namespace, run: 

`supergloo install -m istio -n istio-system -s`


