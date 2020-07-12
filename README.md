# To test the major refactor:

Make sure kind is set up with istio, then:

```bash
# register cluster
go run cmd/cli/main.go cluster register \
    --cluster-name master-cluster \
    --cluster-domain-override=host.docker.internal

# run discovery
go run cmd/mesh-discovery/main.go 
```

# Helpful commands:

Issue curl from productpage to local reviews service:

```
k alpha debug --image=curlimages/curl@sha256:aa45e9d93122a3cfdf8d7de272e2798ea63733eeee6d06bd2ee4f2f8c4027d7c -n bookinfo $(kubectl get pod -n bookinfo | grep productpage | awk '{print $1}') -i -- curl -v http://reviews:9080/reviews/123
```

Get Envoy config dump for productpage sidecar:

```
istioctl proxy-config route $(k get pod -n bookinfo | grep productpage | awk '{print $1}').bookinfo -ojson
```

Get Envoy logs from productpage sidecar:

```
kpf -n bookinfo deployment/productpage-v1 15000&; sleep 1.5 && curl 'localhost:15000/logging?level=debug' -XPOST ; killall kubectl; k logs -n bookinfo deploy/productpage-v1 -c istio-proxy -f
```
