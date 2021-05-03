# Example run of Istio test framework

* Currently installs Istio operator based on `pkg/test/manifests/operator/gloo-mesh-istio.yaml` for every cluster provided in the kube config below.
* Then installs 3 echo applications to those clusters and tests that they can talk to each other within the same cluster.

Example run
```shell
go test -v github.com/solo-io/gloo-mesh/test/integration/multi-cluster/routing \
  -args --istio.test.kube.config=/Users/nick/.kube/mp,/Users/nick/.kube/cp-us-east \
  --istio.test.nocleanup=true
```


## TODO
* Use GM to setup routing between the two clusters and use echo calls to test it works


## Cluster setup script using k3d
```sh
#!/bin/bash

NETWORK=demo-1

# create docker network if it does not exist
docker network create $NETWORK || true

## Management plane cluster exposes port 9000 (unused currently)
k3d cluster create mp --image "rancher/k3s:v1.20.2-k3s1"  --k3s-server-arg "--disable=traefik" --network $NETWORK
KUBE_CTX=k3d-mp
k3d kubeconfig get mp > ~/.kube/mp

kubectl label node $KUBE_CTX-server-0 topology.kubernetes.io/region=us-east-1 --context $KUBE_CTX
kubectl label node $KUBE_CTX-server-0 topology.kubernetes.io/zone=us-east-1a --context $KUBE_CTX

## Control plane cluster (us-east) exposes port 9010 (unused currently)
k3d cluster create cp-us-east --image "rancher/k3s:v1.20.2-k3s1"  --k3s-server-arg "--disable=traefik" --network $NETWORK
k3d kubeconfig get cp-us-east > ~/.kube/cp-us-east
KUBE_CTX=k3d-cp-us-east

kubectl label node $KUBE_CTX-server-0 topology.kubernetes.io/region=us-east-1 --context $KUBE_CTX
kubectl label node $KUBE_CTX-server-0 topology.kubernetes.io/zone=us-east-1a --context $KUBE_CTX

```

### Teardown
```shell
#!/bin/bash
NETWORK=demo-1

docker network rm $NETWORK

k3d cluster delete mp
rm  ~/.kube/mp
  
k3d cluster delete cp-us-east
rm  ~/.kube/cp-us-east
```