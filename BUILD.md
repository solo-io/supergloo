# Service Mesh Hub Developer Workflow

Using this guide, you should be able to build Service Mesh Hub locally and run the tests. Using this flow, you can make modifications to the code, run the tests, and submit PRs.

## Prerequisites

Before you get started building any of the components of Service Mesh Hub, you should successfully run the following make target:

```shell
make update-deps
```

This will download any necessary tools Service Mesh Hub uses to build or generate code.

## Building the CLI

The quickest way to get started building Service Mesh Hub is to build the `meshctl` cli. You can do that with the following:

```shell
make meshctl
```

After building, you should see the `meshctl` binary in the `./_output` folder. If you run it, you should see a dev snapshot version:

```shell
./_output/meshctl version

Client: {"version":"0.4.11-9-g7f759222-dirty"}
Server: version undefined, could not find any version of service mesh hub running
```

## Bootstrap an end-to-end environment with Kind

Service Mesh Hub uses a CI script to bootstrap a working environment uses Kind which you can also use for local development. The script is in  `ci/setup-kind.sh`. 

The script does the following:

* create two Kind clusters
* Build Service Mesh Hub from current source, build docker images, push them into kind
* Generate helm charts for this current build
* Install Service Mesh Hub using `meshctl` with the current helm charts
* Register both clusters into Service Mesh Hub
* Install Istio onto both clusters
* Install bookinfo demo from Istio

A successful run of that script probably indicates that your dev environment is good to go.

### Prerequisites

Need at *least* the following:

* helm v3.0.0
* kind v0.7.0
* kubernetes/kubectl at least at 1.15 (1.14 may work as well)
* Go 1.14

Allocate at least 12GB to Docker Desktop, based on anecdotal experience.

### Running

To quickly run the script, you can use the `start-local-env` make target:

```shell script
make start-local-env
```

To cleanup, you can run:

```shell script
make destroy-local-env
```


If the script runs succesfully, you're kubeconfig will be correctly updated to point to the new kind cluster (named `management-plane-*`)

You should now be ready to follow the tutorials in the docs, or start making modifications to the code.

### Making changes

If you make changes to the code, you can easily build and push the image changes to Kind with the following `make` targets:

```
make docker
make kind-load-images
```
If you need to push to a specific cluster, you can specify in the `CLUSTER_NAME` env variable:

```shell script
CLUSTER_NAME=foo make kind-laod-images
```

You can then re-start the Service Mesh Hub components and it will pick up the changes from the new images:

```
kubectl delete po --all -n service-mesh-hub
```

### Contribute code back

Please see the [CONTRIBUTING.md](CONTRIBUTING.md) doc.  We have a `gofmt` and `goimports` `pre-commit` hook that you can install with the following `make` target:

```shell script
make init
```

This will add [the following pre-commit hook](./.githooks/pre-commit) to check for formatting before committing to the PR.

## Understanding Service Mesh Hub components

We have three main areas that see active development:

* our CLI tool, `meshctl`. The source code lives in `./cli/`
* `mesh-discovery`, the pod that discovers service meshes and workloads. Lives in `./services/mesh-discovery`
* `mesh-networking`, the pod that handles user-written Service Mesh Hub configuration. Lives in `./services/mesh-networking`

Throughout all of your development, you can also rely on `meshctl check` to see if it can detect any problems with your setup.

## Important `make` Targets

* `make update-deps` - download Go CLI tools used to generate and format code. Should be run before doing local dev.
Typically does not need to be run more than once.
* `make meshctl` - build `meshctl`
* `make package-index-mgmt-plane-helm` - package the main Helm chart for Service Mesh Hub
* `make package-csr-agent-chart` - package the chart for our CSR agent
* `make generated-code` - regenerate all of our generated mocks, clients, DI code, documentation, etc.

## Common Local Dev Gotchas

### Cross-Cluster Communication in KinD

When registering KinD clusters, you'll have to pass the arg `--local-cluster-domain-override host.docker.internal`
to `meshctl cluster register`. This is because of how KinD reaches the host system network.

### Debugging Pods In Goland

When debugging either of our pods in Goland, if you are running the others in KinD, you'll have to add the following
line to your /etc/hosts file:

```
127.0.0.1 host.docker.internal
``` 

This will code running in Goland on your host system that attempts to resolve `host.docker.internal` to be
directed back to localhost.
