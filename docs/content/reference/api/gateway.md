
---
title: "gateway.proto"
---

## Package : `istio.networking.v1alpha3`



<a name="top"></a>

<a name="API Reference for gateway.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## gateway.proto


## Table of Contents
  - [Gateway](#istio.networking.v1alpha3.Gateway)
  - [Gateway.SelectorEntry](#istio.networking.v1alpha3.Gateway.SelectorEntry)
  - [Port](#istio.networking.v1alpha3.Port)
  - [Server](#istio.networking.v1alpha3.Server)
  - [ServerTLSSettings](#istio.networking.v1alpha3.ServerTLSSettings)

  - [ServerTLSSettings.TLSProtocol](#istio.networking.v1alpha3.ServerTLSSettings.TLSProtocol)
  - [ServerTLSSettings.TLSmode](#istio.networking.v1alpha3.ServerTLSSettings.TLSmode)






<a name="istio.networking.v1alpha3.Gateway"></a>

### Gateway
Gateway describes a load balancer operating at the edge of the mesh receiving incoming or outgoing HTTP/TCP connections.<br><!-- crd generation tags +cue-gen:Gateway:groupName:networking.istio.io +cue-gen:Gateway:version:v1alpha3 +cue-gen:Gateway:storageVersion +cue-gen:Gateway:annotations:helm.sh/resource-policy=keep +cue-gen:Gateway:labels:app=istio-pilot,chart=istio,heritage=Tiller,release=istio +cue-gen:Gateway:subresource:status +cue-gen:Gateway:scope:Namespaced +cue-gen:Gateway:resource:categories=istio-io,networking-istio-io,shortNames=gw --><br><!-- go code generation tags +kubetype-gen +kubetype-gen:groupVersion=networking.istio.io/v1alpha3 +genclient +k8s:deepcopy-gen=true -->


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| servers | [][istio.networking.v1alpha3.Server]({{< ref "gateway.md#istio.networking.v1alpha3.Server" >}}) | repeated | A list of server specifications. |
  | selector | [][istio.networking.v1alpha3.Gateway.SelectorEntry]({{< ref "gateway.md#istio.networking.v1alpha3.Gateway.SelectorEntry" >}}) | repeated | One or more labels that indicate a specific set of pods/VMs on which this gateway configuration should be applied. By default workloads are searched across all namespaces based on label selectors. This implies that a gateway resource in the namespace "foo" can select pods in the namespace "bar" based on labels. This behavior can be controlled via the PILOT_SCOPE_GATEWAY_TO_NAMESPACE environment variable in istiod. If this variable is set to true, the scope of label search is restricted to the configuration namespace in which the the resource is present. In other words, the Gateway resource must reside in the same namespace as the gateway workload instance. If selector is nil, the Gateway will be applied to all workloads. |
  





<a name="istio.networking.v1alpha3.Gateway.SelectorEntry"></a>

### Gateway.SelectorEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  





<a name="istio.networking.v1alpha3.Port"></a>

### Port
Port describes the properties of a specific port of a service.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| number | uint32 |  | A valid non-negative integer port number. |
  | protocol | string |  | The protocol exposed on the port. MUST BE one of HTTP|HTTPS|GRPC|HTTP2|MONGO|TCP|TLS. TLS implies the connection will be routed based on the SNI header to the destination without terminating the TLS connection. |
  | name | string |  | Label assigned to the port. |
  | targetPort | uint32 |  | The port number on the endpoint where the traffic will be received. Applicable only when used with ServiceEntries. |
  





<a name="istio.networking.v1alpha3.Server"></a>

### Server
`Server` describes the properties of the proxy on a given load balancer port. For example,<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: Gateway metadata:   name: my-ingress spec:   selector:     app: my-ingressgateway   servers:   - port:       number: 80       name: http2       protocol: HTTP2     hosts:     - "*" ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: Gateway metadata:   name: my-ingress spec:   selector:     app: my-ingressgateway   servers:   - port:       number: 80       name: http2       protocol: HTTP2     hosts:     - "*" ``` {{</tab>}} {{</tabs>}}<br>Another example<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: Gateway metadata:   name: my-tcp-ingress spec:   selector:     app: my-tcp-ingressgateway   servers:   - port:       number: 27018       name: mongo       protocol: MONGO     hosts:     - "*" ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: Gateway metadata:   name: my-tcp-ingress spec:   selector:     app: my-tcp-ingressgateway   servers:   - port:       number: 27018       name: mongo       protocol: MONGO     hosts:     - "*" ``` {{</tab>}} {{</tabs>}}<br>The following is an example of TLS configuration for port 443<br>{{<tabs category-name="example">}} {{<tab name="v1alpha3" category-value="v1alpha3">}} ```yaml apiVersion: networking.istio.io/v1alpha3 kind: Gateway metadata:   name: my-tls-ingress spec:   selector:     app: my-tls-ingressgateway   servers:   - port:       number: 443       name: https       protocol: HTTPS     hosts:     - "*"     tls:       mode: SIMPLE       serverCertificate: /etc/certs/server.pem       privateKey: /etc/certs/privatekey.pem ``` {{</tab>}}<br>{{<tab name="v1beta1" category-value="v1beta1">}} ```yaml apiVersion: networking.istio.io/v1beta1 kind: Gateway metadata:   name: my-tls-ingress spec:   selector:     app: my-tls-ingressgateway   servers:   - port:       number: 443       name: https       protocol: HTTPS     hosts:     - "*"     tls:       mode: SIMPLE       serverCertificate: /etc/certs/server.pem       privateKey: /etc/certs/privatekey.pem ``` {{</tab>}} {{</tabs>}}


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [istio.networking.v1alpha3.Port]({{< ref "gateway.md#istio.networking.v1alpha3.Port" >}}) |  | The Port on which the proxy should listen for incoming connections. |
  | bind | string |  | $hide_from_docs The ip or the Unix domain socket to which the listener should be bound to. Format: `x.x.x.x` or `unix:///path/to/uds` or `unix://@foobar` (Linux abstract namespace). When using Unix domain sockets, the port number should be 0. |
  | hosts | []string | repeated | One or more hosts exposed by this gateway. While typically applicable to HTTP services, it can also be used for TCP services using TLS with SNI. A host is specified as a `dnsName` with an optional `namespace/` prefix. The `dnsName` should be specified using FQDN format, optionally including a wildcard character in the left-most component (e.g., `prod/*.example.com`). Set the `dnsName` to `*` to select all `VirtualService` hosts from the specified namespace (e.g.,`prod/*`).<br>The `namespace` can be set to `*` or `.`, representing any or the current namespace, respectively. For example, `*/foo.example.com` selects the service from any available namespace while `./foo.example.com` only selects the service from the namespace of the sidecar. The default, if no `namespace/` is specified, is `*/`, that is, select services from any namespace. Any associated `DestinationRule` in the selected namespace will also be used.<br>A `VirtualService` must be bound to the gateway and must have one or more hosts that match the hosts specified in a server. The match could be an exact match or a suffix match with the server's hosts. For example, if the server's hosts specifies `*.example.com`, a `VirtualService` with hosts `dev.example.com` or `prod.example.com` will match. However, a `VirtualService` with host `example.com` or `newexample.com` will not match.<br>NOTE: Only virtual services exported to the gateway's namespace (e.g., `exportTo` value of `*`) can be referenced. Private configurations (e.g., `exportTo` set to `.`) will not be available. Refer to the `exportTo` setting in `VirtualService`, `DestinationRule`, and `ServiceEntry` configurations for details. |
  | tls | [istio.networking.v1alpha3.ServerTLSSettings]({{< ref "gateway.md#istio.networking.v1alpha3.ServerTLSSettings" >}}) |  | Set of TLS related options that govern the server's behavior. Use these options to control if all http requests should be redirected to https, and the TLS modes to use. |
  | defaultEndpoint | string |  | The loopback IP endpoint or Unix domain socket to which traffic should be forwarded to by default. Format should be `127.0.0.1:PORT` or `unix:///path/to/socket` or `unix://@foobar` (Linux abstract namespace). NOT IMPLEMENTED. $hide_from_docs |
  | name | string |  | An optional name of the server, when set must be unique across all servers. This will be used for variety of purposes like prefixing stats generated with this name etc. |
  





<a name="istio.networking.v1alpha3.ServerTLSSettings"></a>

### ServerTLSSettings



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| httpsRedirect | bool |  | If set to true, the load balancer will send a 301 redirect for all http connections, asking the clients to use HTTPS. |
  | mode | [istio.networking.v1alpha3.ServerTLSSettings.TLSmode]({{< ref "gateway.md#istio.networking.v1alpha3.ServerTLSSettings.TLSmode" >}}) |  | Optional: Indicates whether connections to this port should be secured using TLS. The value of this field determines how TLS is enforced. |
  | serverCertificate | string |  | REQUIRED if mode is `SIMPLE` or `MUTUAL`. The path to the file holding the server-side TLS certificate to use. |
  | privateKey | string |  | REQUIRED if mode is `SIMPLE` or `MUTUAL`. The path to the file holding the server's private key. |
  | caCertificates | string |  | REQUIRED if mode is `MUTUAL`. The path to a file containing certificate authority certificates to use in verifying a presented client side certificate. |
  | credentialName | string |  | For gateways running on Kubernetes, the name of the secret that holds the TLS certs including the CA certificates. Applicable only on Kubernetes. The secret (of type `generic`) should contain the following keys and values: `key: <privateKey>` and `cert: <serverCert>`. For mutual TLS,  `cacert: <CACertificate>` can be provided in the same secret or  a separate secret named `<secret>-cacert`. Secret of type tls for server certificates along with ca.crt key for CA certificates is also supported. Only one of server certificates and CA certificate or credentialName can be specified. |
  | subjectAltNames | []string | repeated | A list of alternate names to verify the subject identity in the certificate presented by the client. |
  | verifyCertificateSpki | []string | repeated | An optional list of base64-encoded SHA-256 hashes of the SKPIs of authorized client certificates. Note: When both verify_certificate_hash and verify_certificate_spki are specified, a hash matching either value will result in the certificate being accepted. |
  | verifyCertificateHash | []string | repeated | An optional list of hex-encoded SHA-256 hashes of the authorized client certificates. Both simple and colon separated formats are acceptable. Note: When both verify_certificate_hash and verify_certificate_spki are specified, a hash matching either value will result in the certificate being accepted. |
  | minProtocolVersion | [istio.networking.v1alpha3.ServerTLSSettings.TLSProtocol]({{< ref "gateway.md#istio.networking.v1alpha3.ServerTLSSettings.TLSProtocol" >}}) |  | Optional: Minimum TLS protocol version. |
  | maxProtocolVersion | [istio.networking.v1alpha3.ServerTLSSettings.TLSProtocol]({{< ref "gateway.md#istio.networking.v1alpha3.ServerTLSSettings.TLSProtocol" >}}) |  | Optional: Maximum TLS protocol version. |
  | cipherSuites | []string | repeated | Optional: If specified, only support the specified cipher list. Otherwise default to the default cipher list supported by Envoy. |
  




 <!-- end messages -->


<a name="istio.networking.v1alpha3.ServerTLSSettings.TLSProtocol"></a>

### ServerTLSSettings.TLSProtocol
TLS protocol versions.

| Name | Number | Description |
| ---- | ------ | ----------- |
| TLS_AUTO | 0 | Automatically choose the optimal TLS version. |
| TLSV1_0 | 1 | TLS version 1.0 |
| TLSV1_1 | 2 | TLS version 1.1 |
| TLSV1_2 | 3 | TLS version 1.2 |
| TLSV1_3 | 4 | TLS version 1.3 |



<a name="istio.networking.v1alpha3.ServerTLSSettings.TLSmode"></a>

### ServerTLSSettings.TLSmode
TLS modes enforced by the proxy

| Name | Number | Description |
| ---- | ------ | ----------- |
| PASSTHROUGH | 0 | The SNI string presented by the client will be used as the match criterion in a VirtualService TLS route to determine the destination service from the service registry. |
| SIMPLE | 1 | Secure connections with standard TLS semantics. |
| MUTUAL | 2 | Secure connections to the downstream using mutual TLS by presenting server certificates for authentication. |
| AUTO_PASSTHROUGH | 3 | Similar to the passthrough mode, except servers with this TLS mode do not require an associated VirtualService to map from the SNI value to service in the registry. The destination details such as the service/subset/port are encoded in the SNI value. The proxy will forward to the upstream (Envoy) cluster (a group of endpoints) specified by the SNI value. This server is typically used to provide connectivity between services in disparate L3 networks that otherwise do not have direct connectivity between their respective endpoints. Use of this mode assumes that both the source and the destination are using Istio mTLS to secure traffic. |
| ISTIO_MUTUAL | 4 | Secure connections from the downstream using mutual TLS by presenting server certificates for authentication.  Compared to Mutual mode, this mode uses certificates, representing gateway workload identity, generated automatically by Istio for mTLS authentication. When this mode is used, all other fields in `TLSOptions` should be empty. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

