## Getting Started Guide

### Dependencies

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

#### 2. Initialize Helm

`kubectl apply -f hack/install/helm/helm-service-account.yaml`

`helm init --service-account tiller --upgrade`

#### 3. Install supergloo cli

`make install-cli`

#### 4. Install supergloo server

`make supergloo-server`

#### 5. Set up namespace for supergloo-system

`kubectl create namespace supergloo-system`


## Dev Setup Guide

- After cloning, run `make init` to set up pre-commit githook to enforce Go formatting and imports
- If using IntelliJ/IDEA/GoLand, mark directory `api/external` as Resource Root

### Updating API

- To regenerate API from protos, run `go generate ./...`