# example run of istio test framework

* currently installs istio operator based on `pkg/test/manifests/operator/gloo-mesh-istio.yaml` for every cluster provided in the kube config below.
* then installs 3 echo applications to those clusters and tests that they can talk to each other within the same cluster.

example run
```shell
go test -v github.com/solo-io/gloo-mesh/test/integration/multi-cluster/routing \
  -args --istio.test.kube.config=/users/nick/.kube/mp,/users/nick/.kube/cp-us-east \
  --istio.test.nocleanup=true
```


## todo
* use gm to setup routing between the two clusters and use echo calls to test it works


## cluster setup script using k3d
```sh
#!/bin/bash

network=demo-1

# create docker network if it does not exist
docker network create $network || true

## management plane cluster exposes port 9000 (unused currently)
k3d cluster create mp --image "rancher/k3s:v1.20.2-k3s1"  --k3s-server-arg "--disable=traefik" --network $network
kube_ctx=k3d-mp
k3d kubeconfig get mp > ~/.kube/mp

kubectl label node $kube_ctx-server-0 topology.kubernetes.io/region=us-east-1 --context $kube_ctx
kubectl label node $kube_ctx-server-0 topology.kubernetes.io/zone=us-east-1a --context $kube_ctx

## control plane cluster (us-east) exposes port 9010 (unused currently)
k3d cluster create cp-us-east --image "rancher/k3s:v1.20.2-k3s1"  --k3s-server-arg "--disable=traefik" --network $network
k3d kubeconfig get cp-us-east > ~/.kube/cp-us-east
kube_ctx=k3d-cp-us-east

kubectl label node $kube_ctx-server-0 topology.kubernetes.io/region=us-east-1 --context $kube_ctx
kubectl label node $kube_ctx-server-0 topology.kubernetes.io/zone=us-east-1a --context $kube_ctx

```

### teardown
```shell
#!/bin/bash
network=demo-1

docker network rm $network

k3d cluster delete mp
rm  ~/.kube/mp
  
k3d cluster delete cp-us-east
rm  ~/.kube/cp-us-east
```