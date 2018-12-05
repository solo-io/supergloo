# Installation

In this guide, we’ll walk you through how to install SuperGloo onto your Kubernetes cluster and some basic functionality you can use once installed.

## Dependencies

First, you’ll need a Kubernetes cluster running 1.9 or later, and a functioning [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/) command (tested with client version 1.12) on your local machine. 

To run Kubernetes on your local machine, we suggest [Minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/) — running version 0.24.1 or later. (tested with 0.28.2-0.30.0)

You also need to install the [Helm client](https://docs.helm.sh/using_helm/#install-helm) (tested with 2.11).

## Install the CLI

If this is your first time running SuperGloo, you’ll need to download the command-line interface (CLI) onto your local machine. You’ll use this CLI to interact with SuperGloo, including installing it onto your Kubernetes cluster.

To install the CLI, run:

```
curl -sL https://run.solo.io/supergloo/install | sh
```

Alternatively, you can download the CLI directly via the [SuperGloo releases page](https://github.com/solo-io/supergloo/releases).

Next, add SupetGloo to your path with:

```
export PATH=$PATH:$HOME/.supergloo/bin
```

Verify the CLI is installed and running correctly with:

```
supergloo --version
```


## Install SuperGloo onto the cluster

```
supergloo init
```



# Explore SuperGloo

## Install a new service mesh

Supergloo supports Istio, Consul, and Linkerd2. To install them with default configuration, run the following command:

```
supergloo install -m {meshname} -n {namespace} -s
```

`{meshname}` should be one of `consul`, `istio`, or `linkerd2`. <BR>
`{namespace}` is a namespace where the mesh control plane will be deployed. <BR> 
Supergloo will create this namespace if it doesn't already exist. 

For instance, to deploy `istio` into the `istio-system` namespace, run: 

```
supergloo install -m istio -n istio-system -s
```


