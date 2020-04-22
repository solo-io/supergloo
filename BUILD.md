# Service Mesh Hub Developer Workflow

We have two main areas that see active development:

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

## Run Simple End-To-End Workflow

This guide will work towards getting the script `ci/setup-kind.sh` running. A successful run of that script
probably indicates that your dev environment is good to go.

### Prerequisites

Need at *least* the following:

* helm v3.0.0
* kind v0.7.0
* kubernetes/kubectl at least at 1.15 (1.14 may work as well)
* Go 1.14

Allocate at least 12GB to Docker Desktop, based on anecdotal experience.

### Running

The script will walk through a basic setup for getting Service Mesh Hub installed working in a multicluster context.
You may want to take a look at the different `make` and `meshctl` invocations that the script uses
to build everything and set it all up.
