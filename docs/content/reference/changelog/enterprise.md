---
title: Gloo Mesh Enterprise 
description: |
  Changelogs for Gloo Mesh Enterprise
weight: 8
---
### v1.1.0-beta17

**New Features**

- Access logs are now viewable from the UI, within the workload details modal. (https://github.com/solo-io/gloo-mesh-enterprise/issues/562)

**Fixes**

- Several cluster columns in tables were displaying incorrect names. That is now fixed. (https://github.com/solo-io/gloo-mesh-enterprise/issues/394)
- Switches SWR keys from slash delimeted strings to arrays. (https://github.com/solo-io/gloo-mesh-enterprise/issues/451)
- Set ServiceEntry resolution to `STATIC` if any endpoint addresses are ipv6 to workaround [this Istio issue](https://github.com/envoyproxy/envoy/issues/10489#issuecomment-606290733). (https://github.com/solo-io/gloo-mesh-enterprise/issues/697)


### v1.1.0-beta16

**Dependency Bumps**

- solo-io/gloo-mesh has been upgraded to v1.1.0-beta13.

**New Features**

- Back up docker images to Quay every release. (https://github.com/solo-io/gloo-mesh/issues/1553)
- Add ability to optionally disable relay CA functionality. (https://github.com/solo-io/gloo-mesh/issues/1622)


### v1.1.0-beta15

**New Features**

- Add OIDC authentication support to the dashboard. (https://github.com/solo-io/gloo-mesh-enterprise/issues/544)
- Add external intermediate Vault CA. (https://github.com/solo-io/gloo-mesh/issues/588)

**Fixes**

- The observability graph now properly refreshes on pulse and manual request. (https://github.com/solo-io/gloo-mesh-enterprise/issues/629)
- In the observability graph, edges with security are now highlighted even more strongly. (https://github.com/solo-io/gloo-mesh-enterprise/issues/669)
- Fix a nil error caused by requesting mtls data about an inactive edge that has no other data. This can be done by first selecting an edge, then shrinking the time window to a point where it has no activity anymore. (https://github.com/solo-io/gloo-mesh-enterprise/issues/667)
- Fix metrics api to be able to return more than 1 edge's worth of mtls data. (https://github.com/solo-io/gloo-mesh-enterprise/issues/711)


### v1.0.14

**Helm Changes**

- Add a `defaultMetricsPort` to the enterprise-agent helm value to view stats on the enterprise-agent pod's 9091 port. (https://github.com/solo-io/gloo-mesh-enterprise/issues/696)


### v1.1.0-beta14

**Dependency Bumps**

- solo-io/skv2-enterprise has been upgraded to v0.1.9.

**Helm Changes**

- Add a `defaultMetricsPort` to the enterprise-agent helm value to view stats on the enterprise-agent pod's 9091 port. (https://github.com/solo-io/gloo-mesh-enterprise/issues/696)


### v1.0.13

**Fixes**

- Fix traffic shift to subsets on VirtualDestinations. (https://github.com/solo-io/gloo-mesh-enterprise/issues/623)


### v1.1.0-beta13

**Fixes**

- Removed dead graph dropdown options (https://github.com/solo-io/gloo-mesh-enterprise/issues/656)
- Fix traffic shift to subsets on VirtualDestinations. (https://github.com/solo-io/gloo-mesh-enterprise/issues/623)


### v1.1.0-beta12

**New Features**

- Allow pagination on the debug apiserver backend. (https://github.com/solo-io/gloo-mesh-enterprise/issues/500)
- The tables within debug meshes on the ui are now pagable, for your large dataset needs. (https://github.com/solo-io/gloo-mesh-enterprise/issues/553)
- The observability graph now has buttons to control zoom and centering for those who prefer or have to avoid trackpad movements. (https://github.com/solo-io/gloo-mesh-enterprise/issues/517)
- Observability graph edges now have more information. Data is kept more up-to-date for details of edges within the sidebar. (https://github.com/solo-io/gloo-mesh-enterprise/issues/561)
- The tables within policies on the ui are now pagable, for your large dataset needs. (https://github.com/solo-io/gloo-mesh-enterprise/issues/553)
- The wasm page of the ui now has pagination, for your large dataset needs. (https://github.com/solo-io/gloo-mesh-enterprise/issues/553)

**Fixes**

- Within the observability graph, we now display vertically 'stacking' namespaces without issue. (https://github.com/solo-io/gloo-mesh-enterprise/issues/648)
- Errors with policies are now shown properly on their details. (https://github.com/solo-io/gloo-mesh-enterprise/issues/571)


### v1.0.12

**Fixes**

- Fix VirtualDestination implementation to account for Istio's [new implementation of Gateway `AUTO_PASSTHROUGH`](https://istio.io/latest/news/security/istio-security-2021-006/). (https://github.com/solo-io/gloo-mesh/issues/1560)


### v1.1.0-beta11

**New Features**

- It is now possible to 'jump back' in the UI using the breadcrumb with more context kept across areas. (https://github.com/solo-io/gloo-mesh-enterprise/issues/438)
- Add a ListAllMeshKeys endpoint to the MeshApi that allows the UI to get all the mesh refs and virtual mesh refs to display as filterable options. (https://github.com/solo-io/gloo-mesh-enterprise/issues/504)
- Pagination hooked up the role-based access tables within the admin section. (https://github.com/solo-io/gloo-mesh-enterprise/issues/553)
- Implement the ServiceDependency API against Istio. (https://github.com/solo-io/gloo-mesh/issues/750)

**Fixes**

- The Observability graph now prevents clusters and namespace boxes from being drawn overlapping. (https://github.com/solo-io/gloo-mesh-enterprise/issues/543)
- Fix VirtualDestination implementation to account for Istio's [new implementation of Gateway `AUTO_PASSTHROUGH`](https://istio.io/latest/news/security/istio-security-2021-006/). (https://github.com/solo-io/gloo-mesh/issues/1560)
- Workload policies now properly name themselves as such, rather than destinations, within the UI. (https://github.com/solo-io/gloo-mesh-enterprise/issues/440)


### v1.0.11

**Dependency Bumps**

- solo-io/gloo-mesh has been upgraded to v1.0.6.
- solo-io/skv2 has been upgraded to v0.17.16.

**Fixes**

- Opt-out of sidecar injection in the Gloo Mesh Enterprise pod annotations. (https://github.com/solo-io/gloo-mesh/issues/1431)
- Switch ServiceEntries with DNS hostnames to type DNS to handle 
VirtualDestinations which failover to LoadBalancers with hostnames. (https://github.com/solo-io/gloo-mesh/issues/1415)


### v1.1.0-beta10

**Dependency Bumps**

- solo-io/skv2 has been upgraded to v0.17.18.
- solo-io/gloo-mesh has been upgraded to v1.1.0-beta6.

**New Features**

- Edge details are now available to examine in the metrics graph of the ui. (https://github.com/solo-io/gloo-mesh-enterprise/issues/429)
- Add the ability to paginate clusters, meshes, rbac roles, and rbac subjects on the apiserver. (https://github.com/solo-io/gloo-mesh-enterprise/issues/494)
- Can now filter the metrics graph by virtual meshes. (https://github.com/solo-io/gloo-mesh-enterprise/issues/471)

**Fixes**

- Respect the list options in the inMemorySnapshotBuilder. (https://github.com/solo-io/gloo-mesh-enterprise/issues/502)
- The metrics main graph and workload details graph now both show proper times. (https://github.com/solo-io/gloo-mesh-enterprise/issues/507)
- The detailed views for workloads and destinations now open properly from their mesh tables. (https://github.com/solo-io/gloo-mesh-enterprise/issues/520)
- Multiselect's typeahead functionality is far more user friendly. (https://github.com/solo-io/gloo-mesh-enterprise/issues/478)
- The metrics graph within workload details now allows you to choose inbound, outbound, or both data types. (https://github.com/solo-io/gloo-mesh-enterprise/issues/475)
- The metrics main graph now handles even smaller heighted screens more gracefully. (https://github.com/solo-io/gloo-mesh-enterprise/issues/469)
- Pagination hooked up. (https://github.com/solo-io/gloo-mesh-enterprise/issues/499)


### v1.1.0-beta9

**Fixes**

- Namespaces are now shown in the metrics graph in each cluster, and one-offs of cluster and namespace nodes are all highlighted. (https://github.com/solo-io/gloo-mesh-enterprise/issues/493)
- The node details tabm shown from clicking a node in the graph, now keeps its data fresh. (https://github.com/solo-io/gloo-mesh-enterprise/issues/474)
- Metrics data is now polled better. (https://github.com/solo-io/gloo-mesh-enterprise/issues/451)
- The graph shown in Workload Details now responds better to time choice changes. (https://github.com/solo-io/gloo-mesh-enterprise/issues/492)
- Information in the node details tab is now more cleanly displayed. (https://github.com/solo-io/gloo-mesh-enterprise/issues/491)
- Metrics data is now refreshed when new time constraints are chosen. (https://github.com/solo-io/gloo-mesh-enterprise/issues/457)


### v1.0.10

**Fixes**

- Refer to the Gloo Mesh docs in the Virtual Mesh Settings text rather than the Gloo Edge docs. (https://github.com/solo-io/gloo-mesh-enterprise/issues/420)


### v1.1.0-beta8

**New Features**

- Add mtls data to the metrics object returned by the metrics api, and populate it with data from prometheus. (https://github.com/solo-io/gloo-mesh-enterprise/issues/430)
- Add the ability to filter and paginate on the apiserver. (https://github.com/solo-io/gloo-mesh/issues/1379)

**Fixes**

- The Policies page filters no longer fully reset on every data refetch. (https://github.com/solo-io/gloo-mesh/issues/1534)
- The Mesh Details (virtual and non) tables, now load only when tabbed on to. (https://github.com/solo-io/gloo-mesh/issues/1529)


### v1.1.0-beta7

**Dependency Bumps**

- solo-io/skv2 has been upgraded to v0.17.16.
- solo-io/gloo-mesh has been upgraded to v1.1.0-beta5.

**New Features**

- Install Prometheus by default. (https://github.com/solo-io/gloo-mesh-enterprise/issues/446)
- Link to graphs from main menu and mesh / virtual mesh details. (https://github.com/solo-io/gloo-mesh-enterprise/issues/340)
- Implement server-side filtering & pagination for Gloo Mesh Policies. (https://github.com/solo-io/gloo-mesh/issues/1400)
- The list of nodes viewable via the graph is now trimmed by a mesh selector. (https://github.com/solo-io/gloo-mesh-enterprise/issues/401)
- The approach to determining node and edge health in the observability graph now takes both directions into account at once. (https://github.com/solo-io/gloo-mesh-enterprise/issues/410)
- Users are now shown some preview of what is coming to the observability graph and other features. (https://github.com/solo-io/gloo-mesh-enterprise/issues/402)
- Implement workload metrics API in apiserver. (https://github.com/solo-io/gloo-mesh-enterprise/issues/291)
- Return both workload ref and controller ref in graph response. (https://github.com/solo-io/gloo-mesh-enterprise/issues/452)
- Add apiserver API for workload metrics. (https://github.com/solo-io/gloo-mesh-enterprise/issues/291)

**Fixes**

- Fix for retrieving node metrics when the same workload exists on multiple clusters. (https://github.com/solo-io/gloo-mesh-enterprise/issues/472)
- Surface Prometheus metric registration errors to logs instead of panicking. Panic only if in debug mode. (https://github.com/solo-io/gloo-mesh-enterprise/issues/449)
- Graph now shows up when at least one namespace and filter are selected. (https://github.com/solo-io/gloo-mesh-enterprise/issues/426)
- Fix graph icon scaling. (https://github.com/solo-io/gloo-mesh-enterprise/issues/418)
- Opt-out of sidecar injection in the Gloo Mesh Enterprise pod annotations. (https://github.com/solo-io/gloo-mesh/issues/1431)
- Turn off RBAC installation by default. (https://github.com/solo-io/gloo-mesh/issues/1464)
- Refer to the Gloo Mesh docs in the Virtual Mesh Settings text rather than the Gloo Edge docs. (https://github.com/solo-io/gloo-mesh-enterprise/issues/420)
- The UI graph no longer gets skewy when toggling data options. (https://github.com/solo-io/gloo-mesh-enterprise/issues/444)
- The metrics node details sidebar now gives more information. (https://github.com/solo-io/gloo-mesh-enterprise/issues/404)
- The metrics tab of the workload details modal always loads, though sometimes with a pause. (https://github.com/solo-io/gloo-mesh-enterprise/issues/476)
- The controls for the metrics and traffic tabs of a Workload Details modal now work as intended. (https://github.com/solo-io/gloo-mesh-enterprise/issues/218)
- The workload details Metric graph is now connected to real-time data. (https://github.com/solo-io/gloo-mesh-enterprise/issues/400)
- The workload details Traffic tab is now connected to real-time data. (https://github.com/solo-io/gloo-mesh-enterprise/issues/400)


### v1.1.0-beta6

**Upgrade Notes**

- Add a new input to the GetGraphs endpoint to allow filtering by meshes. If both this new option and the existing namepsace/cluster filtering option are both present, then the results will consist of the intersection of filters. (https://github.com/solo-io/gloo-mesh-enterprise/issues/383)

**New Features**

- Separate clusters are now highlighted within the observability graph. (https://github.com/solo-io/gloo-mesh-enterprise/issues/219)
- Add DescribeWorkload API and hook up UI to it. (https://github.com/solo-io/gloo-mesh-enterprise/issues/292)


### v1.0.9

**Fixes**

- Surface a metric `relay_pull_clients_connected` and `relay_push_clients_connected` to reflect the number of connected Gloo Mesh Agents at any point in time. (https://github.com/solo-io/gloo-mesh/issues/1461)


### v1.1.0-beta5

**New Features**

- System settings are now viewable in the admin section. (https://github.com/solo-io/gloo-mesh-enterprise/issues/335)
- A legend is now available to help clarify metrics graph nodes when confused. (https://github.com/solo-io/gloo-mesh-enterprise/issues/398)

**Fixes**

- Fix enterprise-networking's self-signed cert generation for helm upgrades. (https://github.com/solo-io/gloo-mesh/issues/1485)
- Nodes and edges may now be idle in the observability graph. (https://github.com/solo-io/gloo-mesh-enterprise/issues/407)
- Graph refresh rates now are tracked properly and the manual Refresh button is working properly. (https://github.com/solo-io/gloo-mesh-enterprise/issues/403)
- Add the floating user ID and runAsUser options to the envoy container on the dashboard pod. (https://github.com/solo-io/gloo-mesh/issues/1456)
- The graph data is now only retrieved when appropriate filters are selected. (https://github.com/solo-io/gloo-mesh-enterprise/issues/425)


### v1.0.8

**Fixes**

- Fix enterprise-networking's self-signed cert generation for helm upgrades. (https://github.com/solo-io/gloo-mesh/issues/1485)
- The top menu has been cleaned, docs now lives in a help menu alongside a link to the Slack community (https://github.com/solo-io/gloo-mesh-enterprise/issues/298)


### v1.0.7

**Fixes**

- Add the floating user ID and runAsUser options to the envoy container on the dashboard pod. (https://github.com/solo-io/gloo-mesh/issues/1456)


### v1.0.6

**Fixes**

- Fix permissions for the Gloo Mesh UI Dockerfiles. This fixes the dashboard pod startup on OpenShift. (https://github.com/solo-io/gloo-mesh/issues/1456)


### v1.1.0-beta4

**Fixes**

- Fix VirtualDestination port targeting to use target_port rather than the kubernetes service exposed port. (https://github.com/solo-io/gloo-mesh/issues/1414)


### v1.1.0-beta3

This release contained no user-facing changes.


### v1.1.0-beta2

**New Features**

- Surface a metric `relay_pull_clients_connected` and `relay_push_clients_connected` to reflect the number of connected Gloo Mesh Agents at any point in time. (https://github.com/solo-io/gloo-mesh/issues/1461)

**Fixes**

- Fix permissions for the Gloo Mesh UI Dockerfiles. (https://github.com/solo-io/gloo-mesh/issues/1456)


### v1.0.5

**Fixes**

- Fix an issue with deregistration where Discovery CRDs are not garbage collected in the management cluster, specifically due to an agent disconnecting before the corresponding `KubernetesCluster` CR is removed by the user. (https://github.com/solo-io/gloo-mesh/issues/1453)


### v1.1.0-beta1

**Helm Changes**

- Set enterprise-networking (relay) service type to LoadBalancer by default in order to streamline the install process for users running on managed Kubernetes. (https://github.com/solo-io/gloo-mesh/issues/1411)

**New Features**

- Add apiserver Graph API to get filters (https://github.com/solo-io/gloo-mesh-enterprise/issues/289)
- Add apiserver Graph API to get nodes and edges (https://github.com/solo-io/gloo-mesh-enterprise/issues/288)
- Add generation of swagger spec for REST API. (https://github.com/solo-io/gloo-mesh-enterprise/issues/314)

**Fixes**

- Rename AccessLogs to AccessLogRecords to match with the actual CRD name. (https://github.com/solo-io/gloo-mesh-enterprise/issues/260)
- On apiserver startup, don't depend on connections to the forwarding relay server or the metrics relay server. (https://github.com/solo-io/gloo-mesh-enterprise/issues/304)
- Fix an issue with deregistration where Discovery CRDs are not garbage collected in the management cluster, specifically due to an agent disconnecting before the corresponding `KubernetesCluster` CR is removed by the user. (https://github.com/solo-io/gloo-mesh/issues/1453)
- Fix issue where Settings CRD is overwritten by agent when running agent in management cluster. (https://github.com/solo-io/gloo-mesh/issues/1394)
- Do not segfault if the garbage collector cannot construct a label selector for objects from an inactive cluster. (https://github.com/solo-io/gloo-mesh/issues/1408)
- Include operator values in Helm values documentation. (https://github.com/solo-io/gloo-mesh/issues/1361)
- Cycles will no longer break the graph ui, though they may not display in a pretty manner. (https://github.com/solo-io/gloo-mesh-enterprise/issues/359)
- The top menu has been cleaned, docs now lives in a help menu alongside a link to the Slack community (https://github.com/solo-io/gloo-mesh-enterprise/issues/298)
- Switch ServiceEntries with DNS hostnames to type DNS to handle 
VirtualDestinations which failover to LoadBalancers with hostnames. (https://github.com/solo-io/gloo-mesh/issues/1415)


### v1.0.4

**Helm Changes**

- Set enterprise-networking (relay) service type to LoadBalancer by default in order to streamline the install process for users running on managed Kubernetes. (https://github.com/solo-io/gloo-mesh/issues/1411)


### v1.0.3

**Fixes**

- Fix issue where Settings CRD is overwritten by agent when running agent in management cluster. (https://github.com/solo-io/gloo-mesh/issues/1394)


### v1.0.2

**Fixes**

- Do not segfault if the garbage collector cannot construct a label selector for objects from an inactive cluster. (https://github.com/solo-io/gloo-mesh/issues/1408)
- Optimize the ListAllMeshes call by only returning counts and error counts of its mesh resources (workloads, access log records, wasm filters) instead of listing out each resource. (https://github.com/solo-io/gloo-mesh-enterprise/issues/260)


### v1.0.1

**Fixes**

- Include operator values in Helm values documentation. (https://github.com/solo-io/gloo-mesh/issues/1361)
- Include operator values in Helm values documentation for RBAC webhook. (https://github.com/solo-io/gloo-mesh/issues/1381)
- Add generation of swagger spec for REST API. (https://github.com/solo-io/gloo-mesh-enterprise/issues/314)


### v1.0.0

Gloo Mesh Enterprise 1.0 is now available! See our docs for information on getting started with Gloo Mesh: https://docs.solo.io/gloo-mesh/latest/.

**API changes**
- Enterprise Networking (`networking.enterprise.mesh.gloo.solo.io`) version bump to v1beta1
- Observability (`observability.enterprise.mesh.gloo.solo.io`) version bump to v1
- RBAC (`rbac.enterprise.mesh.gloo.solo.io`) version bump to v1
- Networking Extensions (`extensions.networking.mesh.gloo.solo.io`) version bump to v1beta1

**Relay architecture**
- New pull-based architecture enhances Gloo Mesh’s security posture by removing the requirement of granting Gloo Mesh credentials to managed clusters.

**VirtualDestination API**
- Virtual Destination is an abstraction that allows global service routing and failover based on geographical locality and service health.

**Meshctl plugins**
- Extend meshctl with plugins for added client-side functionality.
- Plugins are managed via a Krew-inspired package manager for easy install, upgrade, and management.

**Flat networking support**
- Gloo Mesh Enterprise now supports flat network topologies, or workloads across compute environments which can communicate with one another without intermediary ingress gateways.

**Access logging**
- AccessLogRecord CRD for configuring collection.
- Enterprise-networking REST endpoint for retrieving logs (`/v0/observability/logs`).
- `meshctl accesslog` plugin for retrieving logs.

**Metrics**
- Batteries-included support for golden metrics (request/success/failure counts, request latencies) for Istio service meshes.
- A Prometheus component is now available in the Gloo Mesh Helm chart and can be optionally installed with GME.
- Enterprise-networking REST endpoint for retrieving metrics (`/v0/observability/metrics/node` and `/v0/observability/metrics/edge`).

### v1.0.0-beta16

This release contained no user-facing changes.


### v1.0.0-beta15

**Dependency Bumps**

- solo-io/gloo-mesh has been upgraded to v1.0.0-beta14.
- solo-io/gloo-mesh has been upgraded to v1.0.0-beta13.

**Helm Changes**

- Add `verbose` Helm value to the enterprise-networking and enterprise-agent helm charts, set to 'false' by default. Setting this value to 'true' will enable debug logging. (https://github.com/solo-io/gloo-mesh-enterprise/issues/262)

**New Features**

- Allow access logs to capture additional data set by envoy filters, using filterStateObjects. For example, this is particularly useful to log filter states set by custom WASM filters. (https://github.com/solo-io/gloo-mesh-enterprise/issues/165)

**Fixes**

- Retry failed helm releases to alleviate race condition where gcloud storage has not been updated with latest versions from previous release steps. (https://github.com/solo-io/gloo-mesh-enterprise/issues/284)
- Expose the admin grpc service separately from relay. (https://github.com/solo-io/gloo-mesh-enterprise/issues/268)


### v1.0.0-beta14

**New Features**

- Add generation for Helm values docs. (https://github.com/solo-io/gloo-mesh-enterprise/issues/270)

**Fixes**

- Fix initialization of AccessLogRecord status. (https://github.com/solo-io/gloo-mesh-enterprise/issues/277)
- Graphing has been hidden away until it is ready to reemerge fully grown. (https://github.com/solo-io/gloo-mesh-enterprise/issues/263)
- Policies in the policy table should no longer shift and churn like so sand in an mobius hourglass. (https://github.com/solo-io/gloo-mesh-enterprise/issues/264)
- Counts for mesh details are now accurate in all locations. (https://github.com/solo-io/gloo-mesh-ui/issues/1565)
- Access log tables are now stacked to avoid overflow issues with longer names. (https://github.com/solo-io/gloo-mesh-enterprise/issues/278)


### v1.0.0-beta13

**Fixes**

- Long lists of meshes in virtual mesh listings now scroll rather than flow off page. (https://github.com/solo-io/gloo-mesh-enterprise/issues/252)


### v1.0.0-beta12

**Fixes**

- Implement selected destinations on virtual destination status. (https://github.com/solo-io/gloo-mesh-enterprise/issues/257)


### v1.0.0-beta11

**Dependency Bumps**

- solo-io/skv2-enterprise has been upgraded to v0.1.4.
- solo-io/gloo-mesh has been upgraded to v1.0.0-beta11.


### v1.0.0-beta10

**Helm Changes**

- Add self-signed helm value for easy cert initialization. (https://github.com/solo-io/gloo-mesh-enterprise/issues/201)

**New Features**

- Translate new ALR fields for including HTTP request/response header/trailers. (https://github.com/solo-io/gloo-mesh-enterprise/issues/212)
- Add golden metrics API for querying request count, success count, failure count, and request latencies for the network modeled as nodes and edges. (https://github.com/solo-io/gloo-mesh-enterprise/issues/181)
- Add Prometheus as subchart to Helm chart. (https://github.com/solo-io/gloo-mesh-enterprise/issues/207)
- Implement Envoy metrics sink to aggregate metrics from Istio workloads configured with enterprise-networking server as the metrics sink. (https://github.com/solo-io/gloo-mesh-enterprise/issues/128)

**Fixes**

- Update ALR status with workloads it applies to. (https://github.com/solo-io/gloo-mesh-enterprise/issues/229)
- If the apiserver is unable to dial the relay server, try again instead of crashing. (https://github.com/solo-io/gloo-mesh-ui/issues/1525)
- ≥ The `NewKubernetesClusterReconciler` call was non-blocking. Due to the addition of a while loop that repeatedly made calls to the relay server, the number of goroutines would increase constantly. This has been fixed by 1) Moving the KubernetesClusterReconciler out of the while loop and separating it from the grpc.Dial retry logic, 2) making the `WatchForRelaySnapshots` call with the grpc.Dial a blocking function. (https://github.com/solo-io/gloo-mesh-enterprise/issues/241)


### v1.0.0-beta9

**Dependency Bumps**

- solo-io/gloo-mesh has been upgraded to v1.0.0-beta7.
- solo-io/cli-kit has been upgraded to v0.1.2.

**Helm Changes**

- Rename GlooMeshApiserver to GlooMeshDashboard, and move all apiserver helm values one level down under GlooMeshDashboard.apiserver. (https://github.com/solo-io/gloo-mesh/issues/1276)

**New Features**

- Allow users to override the enterprise-networking service type via helm. (https://github.com/solo-io/gloo-mesh-enterprise/issues/185)
- Add a new Snapshot Forwarding Relay Server that will forward all the snapshots sent by the remote relay agents to the apiserver. (https://github.com/solo-io/gloo-mesh-enterprise/issues/180)

**Fixes**

- Add a license check to enterprise-networking such that expired trial licenses no longer work, but expired enterprise licenses just log a warning. (https://github.com/solo-io/gloo-mesh-enterprise/issues/195)
- Rename the gloo-mesh-apiserver pod to dashboard, and update the helm chart. (https://github.com/solo-io/gloo-mesh/issues/1276)


### v1.0.0-beta8

**Fixes**

- Fix wasm plugin version command. (https://github.com/solo-io/gloo-mesh-enterprise/issues/189)


### v1.0.0-beta7

**New Features**

- Automate creation of Service Entries for non-mesh external services across all clusters in the virtual mesh using TrafficTargets. (https://github.com/solo-io/gloo-mesh-enterprise/issues/121)



