
---

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



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mode | [istio.networking.v1alpha3.ClientTLSSettings.TLSmode]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.ClientTLSSettings.TLSmode" >}}) |  | Indicates whether connections to this port should be secured using TLS. The value of this field determines how TLS is enforced. |
  | clientCertificate | string |  | REQUIRED if mode is `MUTUAL`. The path to the file holding the client-side TLS certificate to use. Should be empty if mode is `ISTIO_MUTUAL`. |
  | privateKey | string |  | REQUIRED if mode is `MUTUAL`. The path to the file holding the client's private key. Should be empty if mode is `ISTIO_MUTUAL`. |
  | caCertificates | string |  | OPTIONAL: The path to the file containing certificate authority certificates to use in verifying a presented server certificate. If omitted, the proxy will not verify the server's certificate. Should be empty if mode is `ISTIO_MUTUAL`. |
  | credentialName | string |  | The name of the secret that holds the TLS certs for the client including the CA certificates. Secret must exist in the same namespace with the proxy using the certificates. The secret (of type `generic`)should contain the following keys and values: `key: <privateKey>`, `cert: <serverCert>`, `cacert: <CACertificate>`. Secret of type tls for client certificates along with ca.crt key for CA certificates is also supported. Only one of client certificates and CA certificate or credentialName can be specified.<br>**NOTE:** This field is currently applicable only at gateways. Sidecars will continue to use the certificate paths. |
  | subjectAltNames | []string | repeated | A list of alternate names to verify the subject identity in the certificate. If specified, the proxy will verify that the server certificate's subject alt name matches one of the specified values. If specified, this list overrides the value of subject_alt_names from the ServiceEntry. |
  | sni | string |  | SNI string to present to the server during TLS handshake. |
  





<a name="istio.networking.v1alpha3.ConnectionPoolSettings"></a>

### ConnectionPoolSettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| tcp | [istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings" >}}) |  | Settings common to both HTTP and TCP upstream connections. |
  | http | [istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings" >}}) |  | HTTP connection pool settings. |
  





<a name="istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings"></a>

### ConnectionPoolSettings.HTTPSettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| http1MaxPendingRequests | int32 |  | Maximum number of pending HTTP requests to a destination. Default 2^32-1. |
  | http2MaxRequests | int32 |  | Maximum number of requests to a backend. Default 2^32-1. |
  | maxRequestsPerConnection | int32 |  | Maximum number of requests per connection to a backend. Setting this parameter to 1 disables keep alive. Default 0, meaning "unlimited", up to 2^29. |
  | maxRetries | int32 |  | Maximum number of retries that can be outstanding to all hosts in a cluster at a given time. Defaults to 2^32-1. |
  | idleTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The idle timeout for upstream connection pool connections. The idle timeout  is defined as the period in which there are no active requests. If not set, the default is 1 hour. When the idle timeout is reached,  the connection will be closed. If the connection is an HTTP/2  connection a drain sequence will occur prior to closing the connection.  Note that request based timeouts mean that HTTP/2 PINGs will not  keep the connection alive. Applies to both HTTP1.1 and HTTP2 connections. |
  | h2UpgradePolicy | [istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings.H2UpgradePolicy]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings.H2UpgradePolicy" >}}) |  | Specify if http1.1 connection should be upgraded to http2 for the associated destination. |
  | useClientProtocol | bool |  | If set to true, client protocol will be preserved while initiating connection to backend. Note that when this is set to true, h2_upgrade_policy will be ineffective i.e. the client connections will not be upgraded to http2. |
  





<a name="istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings"></a>

### ConnectionPoolSettings.TCPSettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxConnections | int32 |  | Maximum number of HTTP1 /TCP connections to a destination host. Default 2^32-1. |
  | connectTimeout | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | TCP connection timeout. format: 1h/1m/1s/1ms. MUST BE >=1ms. Default is 10s. |
  | tcpKeepalive | [istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings.TcpKeepalive]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings.TcpKeepalive" >}}) |  | If set then set SO_KEEPALIVE on the socket to enable TCP Keepalives. |
  





<a name="istio.networking.v1alpha3.ConnectionPoolSettings.TCPSettings.TcpKeepalive"></a>

### ConnectionPoolSettings.TCPSettings.TcpKeepalive



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| probes | uint32 |  | Maximum number of keepalive probes to send without response before deciding the connection is dead. Default is to use the OS level configuration (unless overridden, Linux defaults to 9.) |
  | time | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The time duration a connection needs to be idle before keep-alive probes start being sent. Default is to use the OS level configuration (unless overridden, Linux defaults to 7200s (ie 2 hours.) |
  | interval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | The time duration between keep-alive probes. Default is to use the OS level configuration (unless overridden, Linux defaults to 75s.) |
  





<a name="istio.networking.v1alpha3.DestinationRule"></a>

### DestinationRule



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| host | string |  | The name of a service from the service registry. Service names are looked up from the platform's service registry (e.g., Kubernetes services, Consul services, etc.) and from the hosts declared by [ServiceEntries](https://istio.io/docs/reference/config/networking/service-entry/#ServiceEntry). Rules defined for services that do not exist in the service registry will be ignored.<br>*Note for Kubernetes users*: When short names are used (e.g. "reviews" instead of "reviews.default.svc.cluster.local"), Istio will interpret the short name based on the namespace of the rule, not the service. A rule in the "default" namespace containing a host "reviews" will be interpreted as "reviews.default.svc.cluster.local", irrespective of the actual namespace associated with the reviews service. _To avoid potential misconfigurations, it is recommended to always use fully qualified domain names over short names._<br>Note that the host field applies to both HTTP and TCP services. |
  | trafficPolicy | [istio.networking.v1alpha3.TrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.TrafficPolicy" >}}) |  | Traffic policies to apply (load balancing policy, connection pool sizes, outlier detection). |
  | subsets | [][istio.networking.v1alpha3.Subset]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.Subset" >}}) | repeated | One or more named sets that represent individual versions of a service. Traffic policies can be overridden at subset level. |
  | exportTo | []string | repeated | A list of namespaces to which this destination rule is exported. The resolution of a destination rule to apply to a service occurs in the context of a hierarchy of namespaces. Exporting a destination rule allows it to be included in the resolution hierarchy for services in other namespaces. This feature provides a mechanism for service owners and mesh administrators to control the visibility of destination rules across namespace boundaries.<br>If no namespaces are specified then the destination rule is exported to all namespaces by default.<br>The value "." is reserved and defines an export to the same namespace that the destination rule is declared in. Similarly, the value "*" is reserved and defines an export to all namespaces. |
  





<a name="istio.networking.v1alpha3.LoadBalancerSettings"></a>

### LoadBalancerSettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| simple | [istio.networking.v1alpha3.LoadBalancerSettings.SimpleLB]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.LoadBalancerSettings.SimpleLB" >}}) |  |  |
  | consistentHash | [istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB" >}}) |  |  |
  | localityLbSetting | [istio.networking.v1alpha3.LocalityLoadBalancerSetting]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.LocalityLoadBalancerSetting" >}}) |  | Locality load balancer settings, this will override mesh wide settings in entirety, meaning no merging would be performed between this object and the object one in MeshConfig |
  





<a name="istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB"></a>

### LoadBalancerSettings.ConsistentHashLB



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| httpHeaderName | string |  | Hash based on a specific HTTP header. |
  | httpCookie | [istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB.HTTPCookie]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB.HTTPCookie" >}}) |  | Hash based on HTTP cookie. |
  | useSourceIp | bool |  | Hash based on the source IP address. |
  | httpQueryParameterName | string |  | Hash based on a specific HTTP query parameter. |
  | minimumRingSize | uint64 |  | The minimum number of virtual nodes to use for the hash ring. Defaults to 1024. Larger ring sizes result in more granular load distributions. If the number of hosts in the load balancing pool is larger than the ring size, each host will be assigned a single virtual node. |
  





<a name="istio.networking.v1alpha3.LoadBalancerSettings.ConsistentHashLB.HTTPCookie"></a>

### LoadBalancerSettings.ConsistentHashLB.HTTPCookie



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name of the cookie. |
  | path | string |  | Path to set for the cookie. |
  | ttl | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Lifetime of the cookie. |
  





<a name="istio.networking.v1alpha3.LocalityLoadBalancerSetting"></a>

### LocalityLoadBalancerSetting



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| distribute | [][istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute" >}}) | repeated | Optional: only one of distribute or failover can be set. Explicitly specify loadbalancing weight across different zones and geographical locations. Refer to [Locality weighted load balancing](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/locality_weight) If empty, the locality weight is set according to the endpoints number within it. |
  | failover | [][istio.networking.v1alpha3.LocalityLoadBalancerSetting.Failover]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.LocalityLoadBalancerSetting.Failover" >}}) | repeated | Optional: only failover or distribute can be set. Explicitly specify the region traffic will land on when endpoints in local region becomes unhealthy. Should be used together with OutlierDetection to detect unhealthy endpoints. Note: if no OutlierDetection specified, this will not take effect. |
  | enabled | [google.protobuf.BoolValue]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.BoolValue" >}}) |  | enable locality load balancing, this is DestinationRule-level and will override mesh wide settings in entirety. e.g. true means that turn on locality load balancing for this DestinationRule no matter what mesh wide settings is. |
  





<a name="istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute"></a>

### LocalityLoadBalancerSetting.Distribute



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | string |  | Originating locality, '/' separated, e.g. 'region/zone/sub_zone'. |
  | to | [][istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute.ToEntry]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute.ToEntry" >}}) | repeated | Map of upstream localities to traffic distribution weights. The sum of all weights should be 100. Any locality not present will receive no traffic. |
  





<a name="istio.networking.v1alpha3.LocalityLoadBalancerSetting.Distribute.ToEntry"></a>

### LocalityLoadBalancerSetting.Distribute.ToEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | uint32 |  |  |
  





<a name="istio.networking.v1alpha3.LocalityLoadBalancerSetting.Failover"></a>

### LocalityLoadBalancerSetting.Failover



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| from | string |  | Originating region. |
  | to | string |  | Destination region the traffic will fail over to when endpoints in the 'from' region becomes unhealthy. |
  





<a name="istio.networking.v1alpha3.OutlierDetection"></a>

### OutlierDetection



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| consecutiveErrors | int32 |  | Number of errors before a host is ejected from the connection pool. Defaults to 5. When the upstream host is accessed over HTTP, a 502, 503, or 504 return code qualifies as an error. When the upstream host is accessed over an opaque TCP connection, connect timeouts and connection error/failure events qualify as an error. $hide_from_docs |
  | consecutiveGatewayErrors | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Number of gateway errors before a host is ejected from the connection pool. When the upstream host is accessed over HTTP, a 502, 503, or 504 return code qualifies as a gateway error. When the upstream host is accessed over an opaque TCP connection, connect timeouts and connection error/failure events qualify as a gateway error. This feature is disabled by default or when set to the value 0.<br>Note that consecutive_gateway_errors and consecutive_5xx_errors can be used separately or together. Because the errors counted by consecutive_gateway_errors are also included in consecutive_5xx_errors, if the value of consecutive_gateway_errors is greater than or equal to the value of consecutive_5xx_errors, consecutive_gateway_errors will have no effect. |
  | consecutive5XxErrors | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | Number of 5xx errors before a host is ejected from the connection pool. When the upstream host is accessed over an opaque TCP connection, connect timeouts, connection error/failure and request failure events qualify as a 5xx error. This feature defaults to 5 but can be disabled by setting the value to 0.<br>Note that consecutive_gateway_errors and consecutive_5xx_errors can be used separately or together. Because the errors counted by consecutive_gateway_errors are also included in consecutive_5xx_errors, if the value of consecutive_gateway_errors is greater than or equal to the value of consecutive_5xx_errors, consecutive_gateway_errors will have no effect. |
  | interval | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Time interval between ejection sweep analysis. format: 1h/1m/1s/1ms. MUST BE >=1ms. Default is 10s. |
  | baseEjectionTime | [google.protobuf.Duration]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.duration#google.protobuf.Duration" >}}) |  | Minimum ejection duration. A host will remain ejected for a period equal to the product of minimum ejection duration and the number of times the host has been ejected. This technique allows the system to automatically increase the ejection period for unhealthy upstream servers. format: 1h/1m/1s/1ms. MUST BE >=1ms. Default is 30s. |
  | maxEjectionPercent | int32 |  | Maximum % of hosts in the load balancing pool for the upstream service that can be ejected. Defaults to 10%. |
  | minHealthPercent | int32 |  | Outlier detection will be enabled as long as the associated load balancing pool has at least min_health_percent hosts in healthy mode. When the percentage of healthy hosts in the load balancing pool drops below this threshold, outlier detection will be disabled and the proxy will load balance across all hosts in the pool (healthy and unhealthy). The threshold can be disabled by setting it to 0%. The default is 0% as it's not typically applicable in k8s environments with few pods per service. |
  





<a name="istio.networking.v1alpha3.Subset"></a>

### Subset



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | Name of the subset. The service name and the subset name can be used for traffic splitting in a route rule. |
  | labels | [][istio.networking.v1alpha3.Subset.LabelsEntry]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.Subset.LabelsEntry" >}}) | repeated | Labels apply a filter over the endpoints of a service in the service registry. See route rules for examples of usage. |
  | trafficPolicy | [istio.networking.v1alpha3.TrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.TrafficPolicy" >}}) |  | Traffic policies that apply to this subset. Subsets inherit the traffic policies specified at the DestinationRule level. Settings specified at the subset level will override the corresponding settings specified at the DestinationRule level. |
  





<a name="istio.networking.v1alpha3.Subset.LabelsEntry"></a>

### Subset.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.TrafficPolicy"></a>

### TrafficPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| loadBalancer | [istio.networking.v1alpha3.LoadBalancerSettings]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.LoadBalancerSettings" >}}) |  | Settings controlling the load balancer algorithms. |
  | connectionPool | [istio.networking.v1alpha3.ConnectionPoolSettings]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.ConnectionPoolSettings" >}}) |  | Settings controlling the volume of connections to an upstream service |
  | outlierDetection | [istio.networking.v1alpha3.OutlierDetection]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.OutlierDetection" >}}) |  | Settings controlling eviction of unhealthy hosts from the load balancing pool |
  | tls | [istio.networking.v1alpha3.ClientTLSSettings]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.ClientTLSSettings" >}}) |  | TLS related settings for connections to the upstream service. |
  | portLevelSettings | [][istio.networking.v1alpha3.TrafficPolicy.PortTrafficPolicy]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.TrafficPolicy.PortTrafficPolicy" >}}) | repeated | Traffic policies specific to individual ports. Note that port level settings will override the destination-level settings. Traffic settings specified at the destination-level will not be inherited when overridden by port-level settings, i.e. default values will be applied to fields omitted in port-level traffic policies. |
  





<a name="istio.networking.v1alpha3.TrafficPolicy.PortTrafficPolicy"></a>

### TrafficPolicy.PortTrafficPolicy



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [istio.networking.v1alpha3.PortSelector]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.virtual_service#istio.networking.v1alpha3.PortSelector" >}}) |  | Specifies the number of a port on the destination service on which this policy is being applied. |
  | loadBalancer | [istio.networking.v1alpha3.LoadBalancerSettings]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.LoadBalancerSettings" >}}) |  | Settings controlling the load balancer algorithms. |
  | connectionPool | [istio.networking.v1alpha3.ConnectionPoolSettings]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.ConnectionPoolSettings" >}}) |  | Settings controlling the volume of connections to an upstream service |
  | outlierDetection | [istio.networking.v1alpha3.OutlierDetection]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.OutlierDetection" >}}) |  | Settings controlling eviction of unhealthy hosts from the load balancing pool |
  | tls | [istio.networking.v1alpha3.ClientTLSSettings]({{< versioned_link_path fromRoot="/reference/api/istio.io.api.networking.v1alpha3.destination_rule#istio.networking.v1alpha3.ClientTLSSettings" >}}) |  | TLS related settings for connections to the upstream service. |
  




 <!-- end messages -->


<a name="istio.networking.v1alpha3.ClientTLSSettings.TLSmode"></a>

### ClientTLSSettings.TLSmode


| Name | Number | Description |
| ---- | ------ | ----------- |
| DISABLE | 0 | Do not setup a TLS connection to the upstream endpoint. |
| SIMPLE | 1 | Originate a TLS connection to the upstream endpoint. |
| MUTUAL | 2 | Secure connections to the upstream using mutual TLS by presenting client certificates for authentication. |
| ISTIO_MUTUAL | 3 | Secure connections to the upstream using mutual TLS by presenting client certificates for authentication. Compared to Mutual mode, this mode uses certificates generated automatically by Istio for mTLS authentication. When this mode is used, all other fields in `ClientTLSSettings` should be empty. |



<a name="istio.networking.v1alpha3.ConnectionPoolSettings.HTTPSettings.H2UpgradePolicy"></a>

### ConnectionPoolSettings.HTTPSettings.H2UpgradePolicy


| Name | Number | Description |
| ---- | ------ | ----------- |
| DEFAULT | 0 | Use the global default. |
| DO_NOT_UPGRADE | 1 | Do not upgrade the connection to http2. This opt-out option overrides the default. |
| UPGRADE | 2 | Upgrade the connection to http2. This opt-in option overrides the default. |



<a name="istio.networking.v1alpha3.LoadBalancerSettings.SimpleLB"></a>

### LoadBalancerSettings.SimpleLB


| Name | Number | Description |
| ---- | ------ | ----------- |
| ROUND_ROBIN | 0 | Round Robin policy. Default |
| LEAST_CONN | 1 | The least request load balancer uses an O(1) algorithm which selects two random healthy hosts and picks the host which has fewer active requests. |
| RANDOM | 2 | The random load balancer selects a random healthy host. The random load balancer generally performs better than round robin if no health checking policy is configured. |
| PASSTHROUGH | 3 | This option will forward the connection to the original IP address requested by the caller without doing any form of load balancing. This option must be used with care. It is meant for advanced use cases. Refer to Original Destination load balancer in Envoy for further details. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

