## Getting Started Guide

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

#### 2. Install supergloo cli

`make install-cli`

> When the CLI is first run, it will ensure that Helm and Supergloo server are deployed to the cluster.

#### 3. Install a mesh

`supergloo install`

This will bring you into an interactive mesh install. 


## Dev Setup Guide

- After cloning, run `make init` to set up pre-commit githook to enforce Go formatting and imports
- If using IntelliJ/IDEA/GoLand, mark directory `api/external` as Resource Root

### Updating API

- To regenerate API from protos, run `go generate ./...`