---
title: Istio DNS for global routing
menuTitle: Istio DNS for global routing
weight: 50
description: Understanding how to configure Istio DNS for resolvable routing between clusters
---

At the moment, when Service Mesh Hub creates [ServiceEntry](https://istio.io/docs/reference/config/networking/service-entry/) resources for Istio to enable cross-cluster routing and service discovery, these entries and the hostnames they use are not directly routable. 

For example, when you have the Bookinfo sample installed across multiple clusters [as in the guides]({{% versioned_link_path fromRoot="/guides" %}}) you will end up with `ServiceEntry` resources automatically created on `cluster-1` which reference services on `cluster-2`

```shell
kubectl get serviceentry -n istio-system

NAME                                          HOSTS                                           LOCATION        RESOLUTION   AGE
istio-ingressgateway.istio-system.cluster-2   [istio-ingressgateway.istio-system.cluster-2]   MESH_INTERNAL   DNS          163m
ratings.default.cluster-2                     [ratings.default.cluster-2]                     MESH_INTERNAL   DNS          163m
reviews.default.cluster-2                     [reviews.default.cluster-2]                     MESH_INTERNAL   DNS          163m
```

From within a pod on `cluster-1`, you cannot reach any of those `hostnames` directly:

```shell
curl reviews.default.cluster-2                                                            

curl: (6) Could not resolve host: reviews.default.cluster-2  
```

For this to work, we need to configure DNS stubbing and Istio's `coredns` DNS server to serve DNS queries based on the `ServiceEntry`s that have been created. 

## Setting up

Using the `ServiceEntry` hostnames as seen in the previous section, what we really want is Istio's `coredns` server to know that it's responsible for `cluster-2` and `cluster-1` hostnames. 


{{% notice note %}}
You should have the `istiocoredns` component enabled when you install Istio. Check the [Istio installation for multicluster]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}) guide.
{{% /notice %}}

Istio's `coredns` server reads from a configmap named `coredns` in the `istio-system` namespace:

```shell
kubectl get cm/coredns -n istio-system
NAME      DATA   AGE
coredns   1      179m
```

Let's add [Server entries](https://coredns.io/manual/plugins/) with plugins to the `istio-coredns-plugin` to stub out domains like the following:

{{< highlight yaml "hl_lines=11-17" >}}
kubectl --context $CLUSTER apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: istio-system
  labels:
    app: istiocoredns
    release: istio
data:
  Corefile: |
    cluster-2 {
             grpc . 127.0.0.1:8053
          } 	
    cluster-1 {
             grpc . 127.0.0.1:8053
          } 
    .:53 {
          errors
          health 
          grpc global 127.0.0.1:8053
          forward . /etc/resolv.conf {
            except global
          }           
          prometheus :9153
          cache 30
          reload
        }
EOF        
{{< /highlight >}}

Everything else in the configmap can stay the same. 

{{% notice note %}}
Use the domain stubs that match your `ServiceEntry`, not `cluster-1` and `cluster-2`. For example, the hostnames will follow the pattern <service>.<namespace>.<cluster-name>. Use whatever `cluster-name` is for the domain stubs. 
{{% /notice %}}


You may need to restart the `istiocoredns` component (you shouldn't have to, it will eventually pick up this new configuration -- but if you want more immediate results, then restart):

```shell
kubectl delete pod -n istio-system -l app=istiocoredns 
```
At this point, the hostnames from the `ServiceEntry` should be routable **if you use the Istio coredns DNS server**. So for example, an `nslookup` against this server should work:

```shell
nslookup ratings.default.cluster-2 istiocoredns.istio-system.svc.cluster.local
Server:    10.8.62.25
Address 1: 10.8.62.25 istiocoredns.istio-system.svc.cluster.local

Name:      ratings.default.cluster-2
Address 1: 240.0.0.2
```

However, the hostname still won't be routable in the Kubernetes cluster until you add a stub domain to the Kubernetes DNS.

## Adding stub domains to kube-dns

You can add stub domains to your Kubernetes DNS to delagate to the Istio `coredns` for certain domains. How you configure this depends on what you use for DNS within Kubernetes. The following examples are given [based on the Istio documentation](https://istio.io/docs/setup/install/multicluster/gateways/#setup-dns)


{{< tabs >}}
{{< tab name="kube-dns" codelang="shell">}}

ISTIO_COREDNS=$(kubectl get svc -n istio-system istiocoredns -o jsonpath={.spec.clusterIP})

kubectl --context $CLUSTER apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: kube-dns
  namespace: kube-system
data:
  stubDomains: |
    {
      "cluster-2": ["$ISTIO_COREDNS"],
      "cluster-1": ["$ISTIO_COREDNS"],
      "global": ["$ISTIO_COREDNS"]
    }
EOF
{{< /tab >}}

{{< tab name="coredns (<1.4.0)" codelang="shell" >}}

ISTIO_COREDNS=$(kubectl get svc -n istio-system istiocoredns -o jsonpath={.spec.clusterIP})

kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
     cluster-2:53 {
        errors
        cache 30
        proxy . "$ISTIO_COREDNS:53"
    } 
    cluster-1:53 {
        errors
        cache 30
        proxy . "$ISTIO_COREDNS:53"
    }  
    global:53 {
        errors
        cache 30
        proxy . "$ISTIO_COREDNS:53"
    }  
    .:53 {
        errors
        health
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           upstream
           fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        proxy . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
EOF    
{{< /tab >}}

{{< tab name="coredns (==1.4.0)" codelang="shell" >}}

ISTIO_COREDNS=$(kubectl get svc -n istio-system istiocoredns -o jsonpath={.spec.clusterIP})

kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    cluster-2:53 {
        errors
        cache 30
        forward . "$ISTIO_COREDNS:53"
    }  
    cluster-1:53 {
        errors
        cache 30
        forward . "$ISTIO_COREDNS:53"
    }  
    global:53 {
        errors
        cache 30
        forward . "$ISTIO_COREDNS:53"
    }  
    .:53 {
        errors
        health
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           upstream
           fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
EOF
{{< /tab >}}


{{< tab name="coredns (>1.4.0)" codelang="shell" >}}

ISTIO_COREDNS=$(kubectl get svc -n istio-system istiocoredns -o jsonpath={.spec.clusterIP})

kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: coredns
  namespace: kube-system
data:
  Corefile: |
    cluster-2:53 {
        errors
        cache 30
        forward . "$ISTIO_COREDNS:53"
    }  
    cluster-1:53 {
        errors
        cache 30
        forward . "$ISTIO_COREDNS:53"
    }  
    global:53 {
        errors
        cache 30
        forward . "$ISTIO_COREDNS:53"
    }  
    .:53 {
        errors
        health
        ready
        kubernetes cluster.local in-addr.arpa ip6.arpa {
           pods insecure
           upstream
           fallthrough in-addr.arpa ip6.arpa
        }
        prometheus :9153
        forward . /etc/resolv.conf
        cache 30
        loop
        reload
        loadbalance
    }
EOF
{{< /tab >}}
{{< /tabs >}}

At this point the hostnames specified in the `ServiceEntry` created by Service Mesh Hub should be routable within your Kuberentes cluster.
