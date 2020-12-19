
---
title: "sidecar.proto"
---

## Package : `istio.networking.v1alpha3`



<a name="top"></a>

<a name="API Reference for sidecar.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## sidecar.proto


## Table of Contents
  - [IstioEgressListener](#istio.networking.v1alpha3.IstioEgressListener)
  - [IstioIngressListener](#istio.networking.v1alpha3.IstioIngressListener)
  - [OutboundTrafficPolicy](#istio.networking.v1alpha3.OutboundTrafficPolicy)
  - [Sidecar](#istio.networking.v1alpha3.Sidecar)
  - [WorkloadSelector](#istio.networking.v1alpha3.WorkloadSelector)
  - [WorkloadSelector.LabelsEntry](#istio.networking.v1alpha3.WorkloadSelector.LabelsEntry)

  - [CaptureMode](#istio.networking.v1alpha3.CaptureMode)
  - [OutboundTrafficPolicy.Mode](#istio.networking.v1alpha3.OutboundTrafficPolicy.Mode)






<a name="istio.networking.v1alpha3.IstioEgressListener"></a>

### IstioEgressListener
`IstioEgressListener` specifies the properties of an outbound traffic listener on the sidecar proxy attached to a workload instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [istio.networking.v1alpha3.Port]({{< ref "gateway.md#istio.networking.v1alpha3.Port" >}}) |  | The port associated with the listener. If using Unix domain socket, use 0 as the port number, with a valid protocol. The port if specified, will be used as the default destination port associated with the imported hosts. If the port is omitted, Istio will infer the listener ports based on the imported hosts. Note that when multiple egress listeners are specified, where one or more listeners have specific ports while others have no port, the hosts exposed on a listener port will be based on the listener with the most specific port. |
  | bind | string |  | The IP or the Unix domain socket to which the listener should be bound to. Port MUST be specified if bind is not empty. Format: `x.x.x.x` or `unix:///path/to/uds` or `unix://@foobar` (Linux abstract namespace). If omitted, Istio will automatically configure the defaults based on imported services, the workload instances to which this configuration is applied to and the captureMode. If captureMode is `NONE`, bind will default to 127.0.0.1. |
  | captureMode | [istio.networking.v1alpha3.CaptureMode]({{< ref "sidecar.md#istio.networking.v1alpha3.CaptureMode" >}}) |  | When the bind address is an IP, the captureMode option dictates how traffic to the listener is expected to be captured (or not). captureMode must be DEFAULT or `NONE` for Unix domain socket binds. |
  | hosts | []string | repeated | One or more service hosts exposed by the listener in `namespace/dnsName` format. Services in the specified namespace matching `dnsName` will be exposed. The corresponding service can be a service in the service registry (e.g., a Kubernetes or cloud foundry service) or a service specified using a `ServiceEntry` or `VirtualService` configuration. Any associated `DestinationRule` in the same namespace will also be used.<br>The `dnsName` should be specified using FQDN format, optionally including a wildcard character in the left-most component (e.g., `prod/*.example.com`). Set the `dnsName` to `*` to select all services from the specified namespace (e.g., `prod/*`).<br>The `namespace` can be set to `*`, `.`, or `~`, representing any, the current, or no namespace, respectively. For example, `*/foo.example.com` selects the service from any available namespace while `./foo.example.com` only selects the service from the namespace of the sidecar. If a host is set to `*/*`, Istio will configure the sidecar to be able to reach every service in the mesh that is exported to the sidecar's namespace. The value `~/*` can be used to completely trim the configuration for sidecars that simply receive traffic and respond, but make no outbound connections of their own.<br>NOTE: Only services and configuration artifacts exported to the sidecar's namespace (e.g., `exportTo` value of `*`) can be referenced. Private configurations (e.g., `exportTo` set to `.`) will not be available. Refer to the `exportTo` setting in `VirtualService`, `DestinationRule`, and `ServiceEntry` configurations for details.<br>**WARNING:** The list of egress hosts in a `Sidecar` must also include the Mixer control plane services if they are enabled. Envoy will not be able to reach them otherwise. For example, add host `istio-system/istio-telemetry.istio-system.svc.cluster.local` if telemetry is enabled, `istio-system/istio-policy.istio-system.svc.cluster.local` if policy is enabled, or add `istio-system/*` to allow all services in the `istio-system` namespace. This requirement is temporary and will be removed in a future Istio release. |
  





<a name="istio.networking.v1alpha3.IstioIngressListener"></a>

### IstioIngressListener
`IstioIngressListener` specifies the properties of an inbound traffic listener on the sidecar proxy attached to a workload instance.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| port | [istio.networking.v1alpha3.Port]({{< ref "gateway.md#istio.networking.v1alpha3.Port" >}}) |  | The port associated with the listener. |
  | bind | string |  | The IP to which the listener should be bound. Must be in the format `x.x.x.x`. Unix domain socket addresses are not allowed in the bind field for ingress listeners. If omitted, Istio will automatically configure the defaults based on imported services and the workload instances to which this configuration is applied to. |
  | captureMode | [istio.networking.v1alpha3.CaptureMode]({{< ref "sidecar.md#istio.networking.v1alpha3.CaptureMode" >}}) |  | The captureMode option dictates how traffic to the listener is expected to be captured (or not). |
  | defaultEndpoint | string |  | The loopback IP endpoint or Unix domain socket to which traffic should be forwarded to. This configuration can be used to redirect traffic arriving at the bind `IP:Port` on the sidecar to a `localhost:port` or Unix domain socket where the application workload instance is listening for connections. Format should be `127.0.0.1:PORT` or `unix:///path/to/socket` |
  





<a name="istio.networking.v1alpha3.OutboundTrafficPolicy"></a>

### OutboundTrafficPolicy
`OutboundTrafficPolicy` sets the default behavior of the sidecar for handling outbound traffic from the application. If your application uses one or more external services that are not known apriori, setting the policy to `ALLOW_ANY` will cause the sidecars to route any unknown traffic originating from the application to its requested destination.  Users are strongly encouraged to use `ServiceEntry` configurations to explicitly declare any external dependencies, instead of using `ALLOW_ANY`, so that traffic to these services can be monitored.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| mode | [istio.networking.v1alpha3.OutboundTrafficPolicy.Mode]({{< ref "sidecar.md#istio.networking.v1alpha3.OutboundTrafficPolicy.Mode" >}}) |  |  |
  | egressProxy | [istio.networking.v1alpha3.Destination]({{< ref "virtual_service.md#istio.networking.v1alpha3.Destination" >}}) |  | Specifies the details of the egress proxy to which unknown traffic should be forwarded to from the sidecar. Valid only if the mode is set to ALLOW_ANY. If not specified when the mode is ALLOW_ANY, the sidecar will send the unknown traffic directly to the IP requested by the application.<br>** NOTE 1**: The specified egress host must be imported in the egress section for the traffic forwarding to work.<br>** NOTE 2**: An Envoy based egress gateway is unlikely to be able to handle plain text TCP connections forwarded from the sidecar. Envoy's dynamic forward proxy can handle only HTTP and TLS connections. $hide_from_docs |
  





<a name="istio.networking.v1alpha3.Sidecar"></a>

### Sidecar
`Sidecar` describes the configuration of the sidecar proxy that mediates inbound and outbound communication of the workload instance to which it is attached.<br><!-- crd generation tags +cue-gen:Sidecar:groupName:networking.istio.io +cue-gen:Sidecar:version:v1alpha3 +cue-gen:Sidecar:storageVersion +cue-gen:Sidecar:annotations:helm.sh/resource-policy=keep +cue-gen:Sidecar:labels:app=istio-pilot,chart=istio,heritage=Tiller,release=istio +cue-gen:Sidecar:subresource:status +cue-gen:Sidecar:scope:Namespaced +cue-gen:Sidecar:resource:categories=istio-io,networking-istio-io --><br><!-- go code generation tags +kubetype-gen +kubetype-gen:groupVersion=networking.istio.io/v1alpha3 +genclient +k8s:deepcopy-gen=true -->


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workloadSelector | [istio.networking.v1alpha3.WorkloadSelector]({{< ref "sidecar.md#istio.networking.v1alpha3.WorkloadSelector" >}}) |  | Criteria used to select the specific set of pods/VMs on which this `Sidecar` configuration should be applied. If omitted, the `Sidecar` configuration will be applied to all workload instances in the same namespace. |
  | ingress | [][istio.networking.v1alpha3.IstioIngressListener]({{< ref "sidecar.md#istio.networking.v1alpha3.IstioIngressListener" >}}) | repeated | Ingress specifies the configuration of the sidecar for processing inbound traffic to the attached workload instance. If omitted, Istio will automatically configure the sidecar based on the information about the workload obtained from the orchestration platform (e.g., exposed ports, services, etc.). If specified, inbound ports are configured if and only if the workload instance is associated with a service. |
  | egress | [][istio.networking.v1alpha3.IstioEgressListener]({{< ref "sidecar.md#istio.networking.v1alpha3.IstioEgressListener" >}}) | repeated | Egress specifies the configuration of the sidecar for processing outbound traffic from the attached workload instance to other services in the mesh. If not specified, inherits the system detected defaults from the namespace-wide or the global default Sidecar. |
  | outboundTrafficPolicy | [istio.networking.v1alpha3.OutboundTrafficPolicy]({{< ref "sidecar.md#istio.networking.v1alpha3.OutboundTrafficPolicy" >}}) |  | Configuration for the outbound traffic policy.  If your application uses one or more external services that are not known apriori, setting the policy to `ALLOW_ANY` will cause the sidecars to route any unknown traffic originating from the application to its requested destination. If not specified, inherits the system detected defaults from the namespace-wide or the global default Sidecar. |
  





<a name="istio.networking.v1alpha3.WorkloadSelector"></a>

### WorkloadSelector
`WorkloadSelector` specifies the criteria used to determine if the `Gateway`, `Sidecar`, or `EnvoyFilter` or `ServiceEntry` configuration can be applied to a proxy. The matching criteria includes the metadata associated with a proxy, workload instance info such as labels attached to the pod/VM, or any other info that the proxy provides to Istio during the initial handshake. If multiple conditions are specified, all conditions need to match in order for the workload instance to be selected. Currently, only label based selection mechanism is supported.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| labels | [][istio.networking.v1alpha3.WorkloadSelector.LabelsEntry]({{< ref "sidecar.md#istio.networking.v1alpha3.WorkloadSelector.LabelsEntry" >}}) | repeated | One or more labels that indicate a specific set of pods/VMs on which the configuration should be applied. The scope of label search is restricted to the configuration namespace in which the the resource is present. |
  





<a name="istio.networking.v1alpha3.WorkloadSelector.LabelsEntry"></a>

### WorkloadSelector.LabelsEntry



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| key | string |  |  |
  | value | string |  |  |
  




 <!-- end messages -->


<a name="istio.networking.v1alpha3.CaptureMode"></a>

### CaptureMode
`CaptureMode` describes how traffic to a listener is expected to be captured. Applicable only when the listener is bound to an IP.

| Name | Number | Description |
| ---- | ------ | ----------- |
| DEFAULT | 0 | The default capture mode defined by the environment. |
| IPTABLES | 1 | Capture traffic using IPtables redirection. |
| NONE | 2 | No traffic capture. When used in an egress listener, the application is expected to explicitly communicate with the listener port or Unix domain socket. When used in an ingress listener, care needs to be taken to ensure that the listener port is not in use by other processes on the host. |



<a name="istio.networking.v1alpha3.OutboundTrafficPolicy.Mode"></a>

### OutboundTrafficPolicy.Mode


| Name | Number | Description |
| ---- | ------ | ----------- |
| REGISTRY_ONLY | 0 | Outbound traffic will be restricted to services defined in the service registry as well as those defined through `ServiceEntry` configurations. |
| ALLOW_ANY | 1 | Outbound traffic to unknown destinations will be allowed, in case there are no services or `ServiceEntry` configurations for the destination port. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

