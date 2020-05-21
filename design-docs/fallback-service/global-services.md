# Motivation
We would like to create a fail-over service topology, where requests are routed locally first,
and spill over to a remote cluster if there are health issues.

To create a failover service:

This CRD (name is pending):

```
kind: FailOverService
metadata:
  name: reviews-remote
spec:
  meshWorkloads:
  - reviews-cluster1
  - reviews-cluster2
```

Will translate to:

# Local Cluster

A service entry representing all the remote services.
each endpoint here points to the istio ingress gateway of a remote cluster.

```
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: reviews-remote
spec:
  hosts:
  - reviews-remote.default.global
  location: MESH_INTERNAL
  ports:
  - name: http1
    number: 9080
    protocol: http
  resolution: DNS
  addresses:
  - 240.0.0.2
  endpoints:
  - address: ${CLUSTER2_GW_ADDR}
    ports:
      http1: 32000 # Do not change this port value
      reviews-cluster2
```

Another service entry. this one exists just so we have an envoy cluster to modify.
we need it so `reviews-global.default.global` is resolvable.

```
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: reviews-global
spec:
  hosts:
  - reviews-global.default.global
  location: MESH_INTERNAL
  ports:
  - name: http1
    number: 9080
    protocol: http
  resolution: STATIC
  addresses:
  - 240.0.0.20
```

This envoy filter modifies the above cluster, and turns it into an aggregate cluster. with the local cluster being priority 0, and the remote one being priority 1.
```
apiVersion: networking.istio.io/v1alpha3
kind: EnvoyFilter
metadata:
  name: reviews-global.default.global
spec:
  configPatches:
  - applyTo: CLUSTER
    match:
      context: ANY
      cluster:
        name: "outbound|9080||reviews-global.default.global"
    patch:
      operation: REMOVE
  - applyTo: CLUSTER
    match:
      context: ANY
      cluster:
        name: "outbound|9080||reviews-global.default.global"
    patch:
      operation: ADD
      value: # cluster specification
        name: "outbound|9080||reviews-global.default.global"
        connect_timeout: 1s
        lb_policy: CLUSTER_PROVIDED
        cluster_type:
          name: envoy.clusters.aggregate
          typed_config:
            "@type": type.googleapis.com/udpa.type.v1.TypedStruct
            type_url: type.googleapis.com/envoy.config.cluster.aggregate.v2alpha.ClusterConfig
            value:
              clusters:
              - outbound|9080||reviews.default.svc.cluster.local
              - outbound|9080||reviews-remote.default.global
```

Destination rule to enable mTLS. This also adds health / outlier detection.
```
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: reviews-remote
spec:
  host: reviews-remote.default.global
  # this should probably be in the individual services, instead of the global one (pending test).
  trafficPolicy:
    outlierDetection:
      baseEjectionTime: 120s
      consecutive5xxErrors: 10
      interval: 5s
  trafficPolicy:
    tls:
      mode: ISTIO_MUTUAL
```

# Remote cluster
On the remote cluster we will configure these objects, to redirect incoming traffic to the correct local service (pending final testing):
```
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: reviews-global.default.global-vs
spec:
  exportTo:
  - '*'
  gateways:
  - istio-multicluster-ingressgateway
  hosts:
  - reviews-global.default.global
  http:
  - route:
    - destination:
        host: reviews.default.svc.cluster.local
        port:
          number: 9080
```

```
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: ingress-api-gateway
  namespace: default
spec:
  selector:
    istio: ingressgateway # use istio default ingress gateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - reviews-global.default.global
```

To summarize, this uses an aggregate cluster to rotue traffic to a local instance first, and a failover when the local one fails health checks (active or passive).
Each cluster in an aggeragate cluster has a lower priority, which means that if we have more than 1 remote cluster we need to figure out the order.

Anther option is to only support 2 clusters.
If we want to support more than 2, we can't use SNI routing. We can use regular L7 routing. or alternatively require that
all remote services are in the same name/namespace.