
---
title: "destination_rule.proto"
---

## Package : `istio.networking.v1alpha3`



<a name="top"></a>

<a name="API Reference for destination_rule.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## destination_rule.proto
Copyright 2018 Istio Authors

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

## Table of Contents
  - [ClientTLSSettings](#istio.networking.v1alpha3.ClientTLSSettings)
  - [ConnectionPoolSettings](#istio.networking.v1alpha3.ConnectionPoolSettings)
  - [ConnectionPoolSettings.HTTPSettings](#istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings)
  - [ConnectionPoolSettings.TCPSettings](#istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings)
  - [ConnectionPoolSettings.TCPSettings.TcpKeepalive](#istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings.TcpKeepalive)
  - [DestinationRule](#istio.networking.v1alpha3.DestinationRule)
  - [LoadBalancerSettings](#istio.networking.v1alpha3.LoadBalancerSettings)
  - [LoadBalancerSettings.ConsistentHashLB](#istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB)
  - [LoadBalancerSettings.ConsistentHashLB.HTTPCookie](#istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB.HTTPCookie)
  - [LocalityLoadBalancerSetting](#istio.networking.v1alpha3.LocalityLoadBalancerSetting)
  - [LocalityLoadBalancerSetting.Distribute](#istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute)
  - [LocalityLoadBalancerSetting.Distribute.ToEntry](#istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute.ToEntry)
  - [LocalityLoadBalancerSetting.Failover](#istio.networking.v1alpha3.LocalityLoadBalancerSetting.Failover)
  - [OutlierDetection](#istio.networking.v1alpha3.OutlierDetection)
  - [Subset](#istio.networking.v1alpha3.Subset)
  - [Subset.LabelsEntry](#istio.networking.v1alpha3.Subset.LabelsEntry)
  - [TrafficPolicy](#istio.networking.v1alpha3.TrafficPolicy)
  - [TrafficPolicy.PortTrafficPolicy](#istio.networking.v1alpha3.TrafficPolicy.PortTrafficPolicy)

  - [ClientTLSSettings.TLSmode](#istio.networking.v1alpha3.ClientTLSSettings.TLSmode)
  - [ConnectionPoolSettings.HTTPSettings.H2UpgradePolicy](#istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings.H2UpgradePolicy)
  - [LoadBalancerSettings.SimpleLB](#istio.networking.v1alpha3.LoadBalancerSettings.SimpleLB)






<a name="istio.networking.v1alpha3.ClientTLSSettings"></a>

### ClientTLSSettings
SSL/TLS related settings for upstream connections. See Envoy's [TLS context](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/auth/cert.proto.html) for more details. These settings are common to both HTTP and TCP upstreams.<br>For example, the following rule configures a client to use mutual TLS for connections to upstream database cluster.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: DestinationRule metadata:   name: db-mtls spec:   host: mydbserver.prod.svc.cluster.local   trafficPolicy:     tls:       mode: MUTUAL       clientCertificate: /etc/certs/myclientcert.pem       privateKey: /etc/certs/client_private_key.pem       caCertificates: /etc/certs/rootcacerts.pem ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: DestinationRule metadata:   name: db-mtls spec:   host: mydbserver.prod.svc.cluster.local   trafficPolicy:     tls:       mode: MUTUAL       clientCertificate: /etc/certs/myclientcert.pem       privateKey: /etc/certs/client_private_key.pem       caCertificates: /etc/certs/rootcacerts.pem ``` {{</tab>}} {{</tabs>}}<br>The following rule configures a client to use TLS when talking to a foreign service whose domain matches *.foo.com.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: DestinationRule metadata:   name: tls-foo spec:   host: "*.foo.com"   trafficPolicy:     tls:       mode: SIMPLE ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: DestinationRule metadata:   name: tls-foo spec:   host: "*.foo.com"   trafficPolicy:     tls:       mode: SIMPLE ``` {{</tab>}} {{</tabs>}}<br>The following rule configures a client to use Istio mutual TLS when talking to rating services.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: DestinationRule metadata:   name: ratings-istio-mtls spec:   host: ratings.prod.svc.cluster.local   trafficPolicy:     tls:       mode: ISTIO_MUTUAL ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: DestinationRule metadata:   name: ratings-istio-mtls spec:   host: ratings.prod.svc.cluster.local   trafficPolicy:     tls:       mode: ISTIO_MUTUAL ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mode | [istio.networking.v1alpha3.ClientTLSSettings.TLSmode]({{< ref "destination_rule.md#istio.networking.v1alpha3.ClientTLSSettings.TLSmode" >}}) |  | Indicates whether connections to this port should be secured using TLS. The value of this field determines how TLS is enforced. |
  | clientCertificate | string |  | REQUIRED if mode is `MUTUAL`. The path to the file holding the client-side TLS certificate to use. Should be empty if mode is `ISTIO_MUTUAL`. |
  | privateKey | string |  | REQUIRED if mode is `MUTUAL`. The path to the file holding the client's private key. Should be empty if mode is `ISTIO_MUTUAL`. |
  | caCertificates | string |  | OPTIONAL: The path to the file containing certificate authority certificates to use in verifying a presented server certificate. If omitted, the proxy will not verify the server's certificate. Should be empty if mode is `ISTIO_MUTUAL`. |
  | credentialName | string |  | The name of the secret that holds the TLS certs for the client including the CA certificates. Secret must exist in the same namespace with the proxy using the certificates. The secret (of type `generic`)should contain the following keys and values: `key: <privateKey>`, `cert: <serverCert>`, `cacert: <CACertificate>`. Secret of type tls for client certificates along with ca.crt key for CA certificates is also supported. Only one of client certificates and CA certificate or credentialName can be specified.<br>**NOTE:** This field is currently applicable only at gateways. Sidecars will continue to use the certificate paths. |
  | subjectAltNames | []string | repeated | A list of alternate names to verify the subject identity in the certificate. If specified, the proxy will verify that the server certificate's subject alt name matches one of the specified values. If specified, this list overrides the value of subject_alt_names from the ServiceEntry. |
  | sni | string |  | SNI string to present to the server during TLS handshake. |
  





<a name="istio.networking.v1alpha3.ConnectionPoolSettings"></a>

### ConnectionPoolSettings
Connection pool settings for an upstream host. The settings apply to each individual host in the upstream service.  See Envoy's [circuit breaker](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/circuit_breaking) for more details. Connection pool settings can be applied at the TCP level as well as at HTTP level.<br>For example, the following rule sets a limit of 100 connections to redis service called myredissrv with a connect timeout of 30ms<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: DestinationRule metadata:   name: bookinfo-redis spec:   host: myredissrv.prod.svc.cluster.local   trafficPolicy:     connectionPool:       tcp:         maxConnections: 100         connectTimeout: 30ms         tcpKeepalive:           time: 7200s           interval: 75s ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: DestinationRule metadata:   name: bookinfo-redis spec:   host: myredissrv.prod.svc.cluster.local   trafficPolicy:     connectionPool:       tcp:         maxConnections: 100         connectTimeout: 30ms         tcpKeepalive:           time: 7200s           interval: 75s ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tcp | [istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings]({{< ref "destination_rule.md#istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings" >}}) |  | Settings common to both HTTP and TCP upstream connections. |
  | http | [istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings]({{< ref "destination_rule.md#istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings" >}}) |  | HTTP connection pool settings. |
  





<a name="istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings"></a>

### ConnectionPoolSettings.HTTPSettings
Settings applicable to HTTP1.1/HTTP2/GRPC connections.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| http1MaxPendingRequests | int32 |  | Maximum number of pending HTTP requests to a destination. Default 2^32-1. |
  | http2MaxRequests | int32 |  | Maximum number of requests to a backend. Default 2^32-1. |
  | maxRequestsPerConnection | int32 |  | Maximum number of requests per connection to a backend. Setting this parameter to 1 disables keep alive. Default 0, meaning "unlimited", up to 2^29. |
  | maxRetries | int32 |  | Maximum number of retries that can be outstanding to all hosts in a cluster at a given time. Defaults to 2^32-1. |
  | idleTimeout | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | The idle timeout for upstream connection pool connections. The idle timeout is defined as the period in which there are no active requests. If not set, the default is 1 hour. When the idle timeout is reached the connection will be closed. Note that request based timeouts mean that HTTP/2 PINGs will not keep the connection alive. Applies to both HTTP1.1 and HTTP2 connections. |
  | h2UpgradePolicy | [istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings.H2UpgradePolicy]({{< ref "destination_rule.md#istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings.H2UpgradePolicy" >}}) |  | Specify if http1.1 connection should be upgraded to http2 for the associated destination. |
  





<a name="istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings"></a>

### ConnectionPoolSettings.TCPSettings
Settings common to both HTTP and TCP upstream connections.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxConnections | int32 |  | Maximum number of HTTP1 /TCP connections to a destination host. Default 2^32-1. |
  | connectTimeout | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | TCP connection timeout. format: 1h/1m/1s/1ms. MUST BE >=1ms. Default is 10s. |
  | tcpKeepalive | [istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings.TcpKeepalive]({{< ref "destination_rule.md#istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings.TcpKeepalive" >}}) |  | If set then set SO_KEEPALIVE on the socket to enable TCP Keepalives. |
  





<a name="istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings.TcpKeepalive"></a>

### ConnectionPoolSettings.TCPSettings.TcpKeepalive
TCP keepalive.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| probes | uint32 |  | Maximum number of keepalive probes to send without response before deciding the connection is dead. Default is to use the OS level configuration (unless overridden, Linux defaults to 9.) |
  | time | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | The time duration a connection needs to be idle before keep-alive probes start being sent. Default is to use the OS level configuration (unless overridden, Linux defaults to 7200s (ie 2 hours.) |
  | interval | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | The time duration between keep-alive probes. Default is to use the OS level configuration (unless overridden, Linux defaults to 75s.) |
  





<a name="istio.networking.v1alpha3.DestinationRule"></a>

### DestinationRule
DestinationRule defines policies that apply to traffic intended for a service after routing has occurred.<br><!-- crd generation tags +cue-gen:DestinationRule:groupName:networking.istio.io +cue-gen:DestinationRule:version:v1alpha3 +cue-gen:DestinationRule:storageVersion +cue-gen:DestinationRule:annotations:helm.sh/resource-policy=keep +cue-gen:DestinationRule:labels:app=istio-pilot,chart=istio,heritage=Tiller,release=istio +cue-gen:DestinationRule:subresource:status +cue-gen:DestinationRule:scope:Namespaced +cue-gen:DestinationRule:resource:categories=istio-io,networking-istio-io,shortNames=dr +cue-gen:DestinationRule:printerColumn:name=Host,type=string,JSONPath=.spec.host,description="The name of a service from the service registry" +cue-gen:DestinationRule:printerColumn:name=Age,type=date,JSONPath=.metadata.creationTimestamp,description="CreationTimestamp is a timestamp  representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations.  Clients may not set this value. It is represented in RFC3339 form and is in UTC. Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata" --><br><!-- go code generation tags +kubetype-gen +kubetype-gen:groupVersion=networking.istio.io/v1alpha3 +genclient +k8s:deepcopy-gen=true -->


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | string |  | The name of a service from the service registry. Service names are looked up from the platform's service registry (e.g., Kubernetes services, Consul services, etc.) and from the hosts declared by [ServiceEntries](https://istio.io/docs/reference/config/networking/service-entry/#ServiceEntry). Rules defined for services that do not exist in the service registry will be ignored.<br>*Note for Kubernetes users*: When short names are used (e.g. "reviews" instead of "reviews.default.svc.cluster.local"), Istio will interpret the short name based on the namespace of the rule, not the service. A rule in the "default" namespace containing a host "reviews" will be interpreted as "reviews.default.svc.cluster.local", irrespective of the actual namespace associated with the reviews service. _To avoid potential misconfigurations, it is recommended to always use fully qualified domain names over short names._<br>Note that the host field applies to both HTTP and TCP services. |
  | trafficPolicy | [istio.networking.v1alpha3.TrafficPolicy]({{< ref "destination_rule.md#istio.networking.v1alpha3.TrafficPolicy" >}}) |  | Traffic policies to apply (load balancing policy, connection pool sizes, outlier detection). |
  | subsets | [][istio.networking.v1alpha3.Subset]({{< ref "destination_rule.md#istio.networking.v1alpha3.Subset" >}}) | repeated | One or more named sets that represent individual versions of a service. Traffic policies can be overridden at subset level. |
  | exportTo | []string | repeated | A list of namespaces to which this destination rule is exported. The resolution of a destination rule to apply to a service occurs in the context of a hierarchy of namespaces. Exporting a destination rule allows it to be included in the resolution hierarchy for services in other namespaces. This feature provides a mechanism for service owners and mesh administrators to control the visibility of destination rules across namespace boundaries.<br>If no namespaces are specified then the destination rule is exported to all namespaces by default.<br>The value "." is reserved and defines an export to the same namespace that the destination rule is declared in. Similarly, the value "*" is reserved and defines an export to all namespaces.<br>NOTE: in the current release, the `exportTo` value is restricted to "." or "*" (i.e., the current namespace or all namespaces). |
  





<a name="istio.networking.v1alpha3.LoadBalancerSettings"></a>

### LoadBalancerSettings
Load balancing policies to apply for a specific destination. See Envoy's load balancing [documentation](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/load_balancing) for more details.<br>For example, the following rule uses a round robin load balancing policy for all traffic going to the ratings service.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: DestinationRule metadata:   name: bookinfo-ratings spec:   host: ratings.prod.svc.cluster.local   trafficPolicy:     loadBalancer:       simple: ROUND_ROBIN ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: DestinationRule metadata:   name: bookinfo-ratings spec:   host: ratings.prod.svc.cluster.local   trafficPolicy:     loadBalancer:       simple: ROUND_ROBIN ``` {{</tab>}} {{</tabs>}}<br>The following example sets up sticky sessions for the ratings service hashing-based load balancer for the same ratings service using the the User cookie as the hash key.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml  apiVersion: networking.istio.io/v1alpha3  kind: DestinationRule  metadata:    name: bookinfo-ratings  spec:    host: ratings.prod.svc.cluster.local    trafficPolicy:      loadBalancer:        consistentHash:          httpCookie:            name: user            ttl: 0s ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml  apiVersion: networking.istio.io/v1beta1  kind: DestinationRule  metadata:    name: bookinfo-ratings  spec:    host: ratings.prod.svc.cluster.local    trafficPolicy:      loadBalancer:        consistentHash:          httpCookie:            name: user            ttl: 0s ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| simple | [istio.networking.v1alpha3.LoadBalancerSettings.SimpleLB]({{< ref "destination_rule.md#istio.networking.v1alpha3.LoadBalancerSettings.SimpleLB" >}}) |  |  |
  | consistentHash | [istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB]({{< ref "destination_rule.md#istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB" >}}) |  |  |
  | localityLbSetting | [istio.networking.v1alpha3.LocalityLoadBalancerSetting]({{< ref "destination_rule.md#istio.networking.v1alpha3.LocalityLoadBalancerSetting" >}}) |  | Locality load balancer settings, this will override mesh wide settings in entirety, meaning no merging would be performed between this object and the object one in MeshConfig |
  





<a name="istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB"></a>

### LoadBalancerSettings.ConsistentHashLB
Consistent Hash-based load balancing can be used to provide soft session affinity based on HTTP headers, cookies or other properties. This load balancing policy is applicable only for HTTP connections. The affinity to a particular destination host will be lost when one or more hosts are added/removed from the destination service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| httpHeaderName | string |  | Hash based on a specific HTTP header. |
  | httpCookie | [istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB.HTTPCookie]({{< ref "destination_rule.md#istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB.HTTPCookie" >}}) |  | Hash based on HTTP cookie. |
  | useSourceIp | bool |  | Hash based on the source IP address. |
  | httpQueryParameterName | string |  | Hash based on a specific HTTP query parameter. |
  | minimumRingSize | uint64 |  | The minimum number of virtual nodes to use for the hash ring. Defaults to 1024. Larger ring sizes result in more granular load distributions. If the number of hosts in the load balancing pool is larger than the ring size, each host will be assigned a single virtual node. |
  





<a name="istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB.HTTPCookie"></a>

### LoadBalancerSettings.ConsistentHashLB.HTTPCookie
Describes a HTTP cookie that will be used as the hash key for the Consistent Hash load balancer. If the cookie is not present, it will be generated.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name of the cookie. |
  | path | string |  | Path to set for the cookie. |
  | ttl | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | Lifetime of the cookie. |
  





<a name="istio.networking.v1alpha3.LocalityLoadBalancerSetting"></a>

### LocalityLoadBalancerSetting
Locality-weighted load balancing allows administrators to control the distribution of traffic to endpoints based on the localities of where the traffic originates and where it will terminate. These localities are specified using arbitrary labels that designate a hierarchy of localities in {region}/{zone}/{sub-zone} form. For additional detail refer to [Locality Weight](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/locality_weight) The following example shows how to setup locality weights mesh-wide.<br>Given a mesh with workloads and their service deployed to "us-west/zone1/*" and "us-west/zone2/*". This example specifies that when traffic accessing a service originates from workloads in "us-west/zone1/*", 80% of the traffic will be sent to endpoints in "us-west/zone1/*", i.e the same zone, and the remaining 20% will go to endpoints in "us-west/zone2/*". This setup is intended to favor routing traffic to endpoints in the same locality. A similar setting is specified for traffic originating in "us-west/zone2/*".<br>```yaml   distribute:     - from: us-west/zone1/*       to:         "us-west/zone1/*": 80         "us-west/zone2/*": 20     - from: us-west/zone2/*       to:         "us-west/zone1/*": 20         "us-west/zone2/*": 80 ```<br>If the goal of the operator is not to distribute load across zones and regions but rather to restrict the regionality of failover to meet other operational requirements an operator can set a 'failover' policy instead of a 'distribute' policy.<br>The following example sets up a locality failover policy for regions. Assume a service resides in zones within us-east, us-west & eu-west this example specifies that when endpoints within us-east become unhealthy traffic should failover to endpoints in any zone or sub-zone within eu-west and similarly us-west should failover to us-east.<br>```yaml  failover:    - from: us-east      to: eu-west    - from: us-west      to: us-east ``` Locality load balancing settings.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| distribute | [][istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute]({{< ref "destination_rule.md#istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute" >}}) | repeated | Optional: only one of distribute or failover can be set. Explicitly specify loadbalancing weight across different zones and geographical locations. Refer to [Locality weighted load balancing](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/locality_weight) If empty, the locality weight is set according to the endpoints number within it. |
  | failover | [][istio.networking.v1alpha3.LocalityLoadBalancerSetting.Failover]({{< ref "destination_rule.md#istio.networking.v1alpha3.LocalityLoadBalancerSetting.Failover" >}}) | repeated | Optional: only failover or distribute can be set. Explicitly specify the region traffic will land on when endpoints in local region becomes unhealthy. Should be used together with OutlierDetection to detect unhealthy endpoints. Note: if no OutlierDetection specified, this will not take effect. |
  | enabled | [google.protobuf.BoolValue]({{< ref "wrappers.md#google.protobuf.BoolValue" >}}) |  | enable locality load balancing, this is DestinationRule-level and will override mesh wide settings in entirety. e.g. true means that turn on locality load balancing for this DestinationRule no matter what mesh wide settings is. |
  





<a name="istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute"></a>

### LocalityLoadBalancerSetting.Distribute
Describes how traffic originating in the 'from' zone or sub-zone is distributed over a set of 'to' zones. Syntax for specifying a zone is {region}/{zone}/{sub-zone} and terminal wildcards are allowed on any segment of the specification. Examples:<br>`*` - matches all localities<br>`us-west/*` - all zones and sub-zones within the us-west region<br>`us-west/zone-1/*` - all sub-zones within us-west/zone-1


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | string |  | Originating locality, '/' separated, e.g. 'region/zone/sub_zone'. |
  | to | [][istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute.ToEntry]({{< ref "destination_rule.md#istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute.ToEntry" >}}) | repeated | Map of upstream localities to traffic distribution weights. The sum of all weights should be 100. Any locality not present will receive no traffic. |
  





<a name="istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute.ToEntry"></a>

### LocalityLoadBalancerSetting.Distribute.ToEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | uint32 |  |  |
  





<a name="istio.networking.v1alpha3.LocalityLoadBalancerSetting.Failover"></a>

### LocalityLoadBalancerSetting.Failover
Specify the traffic failover policy across regions. Since zone and sub-zone failover is supported by default this only needs to be specified for regions when the operator needs to constrain traffic failover so that the default behavior of failing over to any endpoint globally does not apply. This is useful when failing over traffic across regions would not improve service health or may need to be restricted for other reasons like regulatory controls.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | string |  | Originating region. |
  | to | string |  | Destination region the traffic will fail over to when endpoints in the 'from' region becomes unhealthy. |
  





<a name="istio.networking.v1alpha3.OutlierDetection"></a>

### OutlierDetection
A Circuit breaker implementation that tracks the status of each individual host in the upstream service.  Applicable to both HTTP and TCP services.  For HTTP services, hosts that continually return 5xx errors for API calls are ejected from the pool for a pre-defined period of time. For TCP services, connection timeouts or connection failures to a given host counts as an error when measuring the consecutive errors metric. See Envoy's [outlier detection](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/outlier) for more details.<br>The following rule sets a connection pool size of 100 HTTP1 connections with no more than 10 req/connection to the "reviews" service. In addition, it sets a limit of 1000 concurrent HTTP2 requests and configures upstream hosts to be scanned every 5 mins so that any host that fails 7 consecutive times with a 502, 503, or 504 error code will be ejected for 15 minutes.<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: DestinationRule metadata:   name: reviews-cb-policy spec:   host: reviews.prod.svc.cluster.local   trafficPolicy:     connectionPool:       tcp:         maxConnections: 100       http:         http2MaxRequests: 1000         maxRequestsPerConnection: 10     outlierDetection:       consecutiveErrors: 7       interval: 5m       baseEjectionTime: 15m ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: DestinationRule metadata:   name: reviews-cb-policy spec:   host: reviews.prod.svc.cluster.local   trafficPolicy:     connectionPool:       tcp:         maxConnections: 100       http:         http2MaxRequests: 1000         maxRequestsPerConnection: 10     outlierDetection:       consecutiveErrors: 7       interval: 5m       baseEjectionTime: 15m ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| consecutiveErrors | int32 |  | Number of errors before a host is ejected from the connection pool. Defaults to 5. When the upstream host is accessed over HTTP, a 502, 503, or 504 return code qualifies as an error. When the upstream host is accessed over an opaque TCP connection, connect timeouts and connection error/failure events qualify as an error. $hide_from_docs |
  | consecutiveGatewayErrors | [google.protobuf.UInt32Value]({{< ref "wrappers.md#google.protobuf.UInt32Value" >}}) |  | Number of gateway errors before a host is ejected from the connection pool. When the upstream host is accessed over HTTP, a 502, 503, or 504 return code qualifies as a gateway error. When the upstream host is accessed over an opaque TCP connection, connect timeouts and connection error/failure events qualify as a gateway error. This feature is disabled by default or when set to the value 0.<br>Note that consecutive_gateway_errors and consecutive_5xx_errors can be used separately or together. Because the errors counted by consecutive_gateway_errors are also included in consecutive_5xx_errors, if the value of consecutive_gateway_errors is greater than or equal to the value of consecutive_5xx_errors, consecutive_gateway_errors will have no effect. |
  | consecutive5xxErrors | [google.protobuf.UInt32Value]({{< ref "wrappers.md#google.protobuf.UInt32Value" >}}) |  | Number of 5xx errors before a host is ejected from the connection pool. When the upstream host is accessed over an opaque TCP connection, connect timeouts, connection error/failure and request failure events qualify as a 5xx error. This feature defaults to 5 but can be disabled by setting the value to 0.<br>Note that consecutive_gateway_errors and consecutive_5xx_errors can be used separately or together. Because the errors counted by consecutive_gateway_errors are also included in consecutive_5xx_errors, if the value of consecutive_gateway_errors is greater than or equal to the value of consecutive_5xx_errors, consecutive_gateway_errors will have no effect. |
  | interval | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | Time interval between ejection sweep analysis. format: 1h/1m/1s/1ms. MUST BE >=1ms. Default is 10s. |
  | baseEjectionTime | [google.protobuf.Duration]({{< ref "duration.md#google.protobuf.Duration" >}}) |  | Minimum ejection duration. A host will remain ejected for a period equal to the product of minimum ejection duration and the number of times the host has been ejected. This technique allows the system to automatically increase the ejection period for unhealthy upstream servers. format: 1h/1m/1s/1ms. MUST BE >=1ms. Default is 30s. |
  | maxEjectionPercent | int32 |  | Maximum % of hosts in the load balancing pool for the upstream service that can be ejected. Defaults to 10%. |
  | minHealthPercent | int32 |  | Outlier detection will be enabled as long as the associated load balancing pool has at least min_health_percent hosts in healthy mode. When the percentage of healthy hosts in the load balancing pool drops below this threshold, outlier detection will be disabled and the proxy will load balance across all hosts in the pool (healthy and unhealthy). The threshold can be disabled by setting it to 0%. The default is 0% as it's not typically applicable in k8s environments with few pods per service. |
  





<a name="istio.networking.v1alpha3.Subset"></a>

### Subset
A subset of endpoints of a service. Subsets can be used for scenarios like A/B testing, or routing to a specific version of a service. Refer to [VirtualService](https://istio.io/docs/reference/config/networking/virtual-service/#VirtualService) documentation for examples of using subsets in these scenarios. In addition, traffic policies defined at the service-level can be overridden at a subset-level. The following rule uses a round robin load balancing policy for all traffic going to a subset named testversion that is composed of endpoints (e.g., pods) with labels (version:v3).<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: DestinationRule metadata:   name: bookinfo-ratings spec:   host: ratings.prod.svc.cluster.local   trafficPolicy:     loadBalancer:       simple: LEAST_CONN   subsets:   - name: testversion     labels:       version: v3     trafficPolicy:       loadBalancer:         simple: ROUND_ROBIN ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: DestinationRule metadata:   name: bookinfo-ratings spec:   host: ratings.prod.svc.cluster.local   trafficPolicy:     loadBalancer:       simple: LEAST_CONN   subsets:   - name: testversion     labels:       version: v3     trafficPolicy:       loadBalancer:         simple: ROUND_ROBIN ``` {{</tab>}} {{</tabs>}}<br>**Note:** Policies specified for subsets will not take effect until a route rule explicitly sends traffic to this subset.<br>One or more labels are typically required to identify the subset destination, however, when the corresponding DestinationRule represents a host that supports multiple SNI hosts (e.g., an egress gateway), a subset without labels may be meaningful. In this case a traffic policy with [ClientTLSSettings](#ClientTLSSettings) can be used to identify a specific SNI host corresponding to the named subset.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name of the subset. The service name and the subset name can be used for traffic splitting in a route rule. |
  | labels | [][istio.networking.v1alpha3.Subset.LabelsEntry]({{< ref "destination_rule.md#istio.networking.v1alpha3.Subset.LabelsEntry" >}}) | repeated | Labels apply a filter over the endpoints of a service in the service registry. See route rules for examples of usage. |
  | trafficPolicy | [istio.networking.v1alpha3.TrafficPolicy]({{< ref "destination_rule.md#istio.networking.v1alpha3.TrafficPolicy" >}}) |  | Traffic policies that apply to this subset. Subsets inherit the traffic policies specified at the DestinationRule level. Settings specified at the subset level will override the corresponding settings specified at the DestinationRule level. |
  





<a name="istio.networking.v1alpha3.Subset.LabelsEntry"></a>

### Subset.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.TrafficPolicy"></a>

### TrafficPolicy
Traffic policies to apply for a specific destination, across all destination ports. See DestinationRule for examples.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| loadBalancer | [istio.networking.v1alpha3.LoadBalancerSettings]({{< ref "destination_rule.md#istio.networking.v1alpha3.LoadBalancerSettings" >}}) |  | Settings controlling the load balancer algorithms. |
  | connectionPool | [istio.networking.v1alpha3.ConnectionPoolSettings]({{< ref "destination_rule.md#istio.networking.v1alpha3.ConnectionPoolSettings" >}}) |  | Settings controlling the volume of connections to an upstream service |
  | outlierDetection | [istio.networking.v1alpha3.OutlierDetection]({{< ref "destination_rule.md#istio.networking.v1alpha3.OutlierDetection" >}}) |  | Settings controlling eviction of unhealthy hosts from the load balancing pool |
  | tls | [istio.networking.v1alpha3.ClientTLSSettings]({{< ref "destination_rule.md#istio.networking.v1alpha3.ClientTLSSettings" >}}) |  | TLS related settings for connections to the upstream service. |
  | portLevelSettings | [][istio.networking.v1alpha3.TrafficPolicy.PortTrafficPolicy]({{< ref "destination_rule.md#istio.networking.v1alpha3.TrafficPolicy.PortTrafficPolicy" >}}) | repeated | Traffic policies specific to individual ports. Note that port level settings will override the destination-level settings. Traffic settings specified at the destination-level will not be inherited when overridden by port-level settings, i.e. default values will be applied to fields omitted in port-level traffic policies. |
  





<a name="istio.networking.v1alpha3.TrafficPolicy.PortTrafficPolicy"></a>

### TrafficPolicy.PortTrafficPolicy
Traffic policies that apply to specific ports of the service


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [istio.networking.v1alpha3.PortSelector]({{< ref "virtual_service.md#istio.networking.v1alpha3.PortSelector" >}}) |  | Specifies the number of a port on the destination service on which this policy is being applied. |
  | loadBalancer | [istio.networking.v1alpha3.LoadBalancerSettings]({{< ref "destination_rule.md#istio.networking.v1alpha3.LoadBalancerSettings" >}}) |  | Settings controlling the load balancer algorithms. |
  | connectionPool | [istio.networking.v1alpha3.ConnectionPoolSettings]({{< ref "destination_rule.md#istio.networking.v1alpha3.ConnectionPoolSettings" >}}) |  | Settings controlling the volume of connections to an upstream service |
  | outlierDetection | [istio.networking.v1alpha3.OutlierDetection]({{< ref "destination_rule.md#istio.networking.v1alpha3.OutlierDetection" >}}) |  | Settings controlling eviction of unhealthy hosts from the load balancing pool |
  | tls | [istio.networking.v1alpha3.ClientTLSSettings]({{< ref "destination_rule.md#istio.networking.v1alpha3.ClientTLSSettings" >}}) |  | TLS related settings for connections to the upstream service. |
  




 <!-- end messages -->


<a name="istio.networking.v1alpha3.ClientTLSSettings.TLSmode"></a>

### ClientTLSSettings.TLSmode
TLS connection mode

| Name | Number | Description |
| ---- | ------ | ----------- |
| DISABLE | 0 | Do not setup a TLS connection to the upstream endpoint. |
| SIMPLE | 1 | Originate a TLS connection to the upstream endpoint. |
| MUTUAL | 2 | Secure connections to the upstream using mutual TLS by presenting client certificates for authentication. |
| ISTIO_MUTUAL | 3 | Secure connections to the upstream using mutual TLS by presenting client certificates for authentication. Compared to Mutual mode, this mode uses certificates generated automatically by Istio for mTLS authentication. When this mode is used, all other fields in `ClientTLSSettings` should be empty. |



<a name="istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings.H2UpgradePolicy"></a>

### ConnectionPoolSettings.HTTPSettings.H2UpgradePolicy
Policy for upgrading http1.1 connections to http2.

| Name | Number | Description |
| ---- | ------ | ----------- |
| DEFAULT | 0 | Use the global default. |
| DO_NOT_UPGRADE | 1 | Do not upgrade the connection to http2. This opt-out option overrides the default. |
| UPGRADE | 2 | Upgrade the connection to http2. This opt-in option overrides the default. |



<a name="istio.networking.v1alpha3.LoadBalancerSettings.SimpleLB"></a>

### LoadBalancerSettings.SimpleLB
Standard load balancing algorithms that require no tuning.

| Name | Number | Description |
| ---- | ------ | ----------- |
| ROUND_ROBIN | 0 | Round Robin policy. Default |
| LEAST_CONN | 1 | The least request load balancer uses an O(1) algorithm which selects two random healthy hosts and picks the host which has fewer active requests. |
| RANDOM | 2 | The random load balancer selects a random healthy host. The random load balancer generally performs better than round robin if no health checking policy is configured. |
| PASSTHROUGH | 3 | This option will forward the connection to the original IP address requested by the caller without doing any form of load balancing. This option must be used with care. It is meant for advanced use cases. Refer to Original Destination load balancer in Envoy for further details. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

