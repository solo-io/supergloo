---
title: Gloo Mesh Community 
description: |
  Changelogs for Gloo Mesh Community
weight: 7
---
### v1.1.0-beta15

**Fixes**

- Report error if ServiceEntry endpoints have both ipv6 and hostname addresses. (https://github.com/solo-io/gloo-mesh-enterprise/issues/697)


### v1.1.0-beta14

**Fixes**

- Reduce verbosity of discovery region log. (https://github.com/solo-io/gloo-mesh/issues/1656)
- Set ServiceEntry resolution to `STATIC` if any endpoint addresses are ipv6 to workaround [this Istio issue](https://github.com/envoyproxy/envoy/issues/10489#issuecomment-606290733). (https://github.com/solo-io/gloo-mesh-enterprise/issues/697)
- Improve wording of meshctl check cluster output column names and hint message. (https://github.com/solo-io/gloo-mesh/issues/1562)


### v1.1.0-beta13

**New Features**

- Capture all external address information on the Destination CRD so that any externally-addressable Kubernetes Service can be used as an ingress gateway. (https://github.com/solo-io/gloo-mesh/issues/1611)
- Mark as deprecated the `IngressGatewayDetector` in Settings, which is no longer needed now that external address information is colocated in the relevant Destination. (https://github.com/solo-io/gloo-mesh/issues/1611)
- Add a `meshctl cluster configure` command that interactively allows the user to generate a file that is referenced in subsequent meshctl commands which automate cluster interaction. (https://github.com/solo-io/gloo-mesh/issues/1584)
- Add a `meshctl cluster list` command that allows the user to list all the registered clusters. (https://github.com/solo-io/gloo-mesh/issues/1584)
- Add a `meshctl debug report` command that selectively captures cluster information and logs into an archive to help diagnose problems. (https://github.com/solo-io/gloo-mesh-enterprise/issues/581)
- Update `meshctl uninstall` to delete all Gloo Mesh management plane CRDs. (https://github.com/solo-io/gloo-mesh/issues/1644)

**Fixes**

- Discover externally addressable Kubernetes Services from any namespace, any of which can act as an ingress gateway. (https://github.com/solo-io/gloo-mesh-enterprise/issues/664)
- In the destination federation translator, gracefully handle case when Istio mesh has no ingress gateway. (https://github.com/solo-io/gloo-mesh/issues/1630)
- Fix VirtualMesh link in traffic policy doc. (https://github.com/solo-io/gloo-mesh/issues/1385)
- Fix TrafficPolicySpec link in traffic policy doc. (https://github.com/solo-io/gloo-mesh/issues/1540)
- Modify the `meshctl check` command such that we are reporting the number of connected pull and push agents instead of checking that it is a number. (https://github.com/solo-io/gloo-mesh/issues/1562)


### v1.0.9

**Dependency Bumps**

- solo-io/skv2 has been upgraded to v0.17.20.

**Helm Changes**

- Expose helm options for setting custom labels and annotations on pods, deployments, and services. (https://github.com/solo-io/gloo-mesh/issues/1450)

**Fixes**

- Back up docker images to Quay every release. (https://github.com/solo-io/gloo-mesh/issues/1553)
- Update federation to use port protocol for ServiceEntry port names when no name is specified on the Kube Service. (https://github.com/solo-io/gloo-mesh/issues/1554)
- Fix multicluster subset routing. (https://github.com/solo-io/gloo-mesh-enterprise/issues/645)


### v1.1.0-beta12

**Dependency Bumps**

- solo-io/skv2 has been upgraded to v0.17.20.

**Helm Changes**

- Expose helm options for setting custom labels and annotations on pods, deployments, and services. (https://github.com/solo-io/gloo-mesh/issues/1450)

**Fixes**

- Fixed kubeconfig for cli command meshctl cluster register (https://github.com/solo-io/gloo-mesh/issues/1627)


### v1.1.0-beta11

**Upgrade Notes**

- Rename most instances of 'mgmt' and 'remote' contexts and clusters in docs to 1 and 2. The few exceptions that remain are for contexts where the distinction matters, and for values produced from meshctl, which still uses 'mgmt' and 'remote' and its default naming. (https://github.com/solo-io/gloo-mesh/issues/1452)

**New Features**

- Updates the settings CRD with settings to configure authentication on the dashboard. (https://github.com/solo-io/gloo-mesh-enterprise/issues/544)

**Fixes**

- Update federation to use port protocol for ServiceEntry port names when no name is specified on the Kube Service. (https://github.com/solo-io/gloo-mesh/issues/1554)
- Fix multicluster subset routing. (https://github.com/solo-io/gloo-mesh-enterprise/issues/645)
- Exposes --mgmt-kubeconfig as an option when registering a cluster using meshctl cluster register (https://github.com/solo-io/gloo-mesh/issues/1547)
- Honor the kubeconfig flag in the meshctl dashboard command. (https://github.com/solo-io/gloo-mesh/issues/1616)
- Adds redis options to the dashboard CRD. (https://github.com/solo-io/gloo-mesh-enterprise/issues/544)


### v1.1.0-beta10

**New Features**

- Back up docker images to Quay every release. (https://github.com/solo-io/gloo-mesh/issues/1553)

**Fixes**

- Fix the meshctl install script to use either python3 or python2 for installation, instead of accepting only  python2. (https://github.com/solo-io/gloo-mesh/issues/1386)


### v1.0.8

**Fixes**

- Backport the handling of VirtualDestination as a traffic shift subset target. (https://github.com/solo-io/gloo-mesh/issues/1589)


### v1.0.7

**Fixes**

- Fix cross cluster federation implementation to account for Istio's [new implementation of Gateway `AUTO_PASSTHROUGH`](https://istio.io/latest/news/security/istio-security-2021-006/). (https://github.com/solo-io/gloo-mesh/issues/1560)


### v1.1.0-beta9

This release contained no user-facing changes.


### v1.1.0-beta8

**Dependency Bumps**

- solo-io/k8s.io/client-go has been upgraded to v0.20.4.
- solo-io/k8s.io/cli-runtime has been upgraded to v0.20.4.
- solo-io/k8s.io/apimachinery has been upgraded to v0.20.4.
- solo-io/k8s.io/apiextensions-apiserver has been upgraded to v0.20.4.
- solo-io/k8s.io/api has been upgraded to v0.20.4.
- solo-io/istio.io/istio has been upgraded to v1.9.4.
- solo-io/istio.io/pkg has been upgraded to v1.9.4.
- solo-io/istio.io/api has been upgraded to v1.9.4.

**New Features**

- Gloo Mesh will now discover an ExternalIP on an Ingress Gateway service if the status is not set. This value should be populated by the user when using a non-kubernetes LoadBalancer. (https://github.com/solo-io/gloo-mesh/issues/1564)
- Deploy istio/gloo-mesh to multiple clusters and run e2e tests based on different scenarios. Currently only tested with Enterprise version. (https://github.com/solo-io/gloo-mesh/issues/1507)

**Fixes**

- Fix cross cluster federation implementation to account for Istio's [new implementation of Gateway `AUTO_PASSTHROUGH`](https://istio.io/latest/news/security/istio-security-2021-006/). (https://github.com/solo-io/gloo-mesh/issues/1560)


### v1.0.6

**Fixes**

- Update the VirtualMesh examples to explicitly specify permissive federation. (https://github.com/solo-io/gloo-mesh/issues/1536)
- Update the helm urls to refer to the major and minor version in the links, and write the full version on the helm reference page. (https://github.com/solo-io/gloo-mesh/issues/1395)
- Tag Istio 1.8.5 in the docs. (https://github.com/solo-io/gloo-mesh/issues/1472)
- Gloo Mesh will now discover an ExternalIP on an Ingress Gateway service if the status is not set. This value should be populated by the user when using a non-kubernetes LoadBalancer. (https://github.com/solo-io/gloo-mesh/issues/1564)
- Opt-out of sidecar injection in the Gloo Mesh cert-agent pod annotations. (https://github.com/solo-io/gloo-mesh/issues/1431)


### v1.1.0-beta7

**New Features**

- Add API for external CA feature. (https://github.com/solo-io/gloo-mesh/issues/588)
- Introduce the new ServiceDependency API. (https://github.com/solo-io/gloo-mesh/issues/750)

**Fixes**

- Conflict detection should respect cluster. (https://github.com/solo-io/gloo-mesh/issues/1544)


### v1.1.0-beta6

**Fixes**

- Update the helm urls to refer to the major and minor version in the links, and write the full version on the helm reference page. (https://github.com/solo-io/gloo-mesh/issues/1395)
- Fix the API docs for reference based selectors to say that all fields are required. Remove the incorrect statement that omission means select all. (https://github.com/solo-io/gloo-mesh/issues/1517)
- Tag Istio 1.8.5 in the docs. (https://github.com/solo-io/gloo-mesh/issues/1472)
- Update the Gloo Mesh gloomesh-remote-access clusterrole in the docs. (https://github.com/solo-io/gloo-mesh/issues/1523)
- Respect the list options in the inMemorySnapshotBuilder. (https://github.com/solo-io/gloo-mesh-enterprise/issues/502)
- Opt-out of sidecar injection in the Gloo Mesh cert-agent pod annotations. (https://github.com/solo-io/gloo-mesh/issues/1431)


### v1.1.0-beta5

**New Features**

- Add the option to pass --set to add extra helm values to the Gloo Mesh installation using meshctl. (https://github.com/solo-io/gloo-mesh/issues/1478)
- Optimize subset translation by reducing iteration across Destinations. Surface required subsets on Destination status with new `required_subsets` field. (https://github.com/solo-io/gloo-mesh/issues/1490)


### v1.1.0-beta4

**Dependency Bumps**

- solo-io/skv2 has been upgraded to v0.17.16.

**New Features**

- Add cluster registration check in `meshctl check`. Shows the number of connected Gloo Mesh Agents at any point in time and compares it to the number of registered clusters (`KubernetesCluster` CRs in the Gloo Mesh namespace). Requires Gloo Mesh Enterprise v1.1.0-beta3 or greater. (https://github.com/solo-io/gloo-mesh/issues/1461)
- Implement restrictive federation mode for VirtualMeshes, which allows for selectively exposing Destinations to external Meshes. (https://github.com/solo-io/gloo-mesh/issues/804)

**Fixes**

- Clarify the deregister command flows, specifically with regards to context. (https://github.com/solo-io/gloo-mesh/issues/1412)
- Add Enterprise disclaimer for Locality Routing. (https://github.com/solo-io/gloo-mesh/issues/1428)
- Improve the enterprise install and register documentation. (https://github.com/solo-io/gloo-mesh/issues/1451)
- Deprecate the `skip-rbac` flag and add the `include-rbac` flag. (https://github.com/solo-io/gloo-mesh/issues/1464)
- Opt-out of sidecar injection in the Gloo Mesh pod annotations. (https://github.com/solo-io/gloo-mesh/issues/1431)
- Opt-out of sidecar injection in the Gloo Mesh install/registration namespace. (https://github.com/solo-io/gloo-mesh/issues/1431)


### v1.0.5

**Fixes**

- Add Enterprise disclaimer for Locality Routing. (https://github.com/solo-io/gloo-mesh/issues/1428)
- Add cluster registration check in `meshctl check`. Shows the number of connected Gloo Mesh Agents at any point in time and compares it to the number of registered clusters (`KubernetesCluster` CRs in the Gloo Mesh namespace). Requires Gloo Mesh Enterprise v1.0.5 or greater. (https://github.com/solo-io/gloo-mesh/issues/1461)
- Lower OSM destination detector log level to debug. (https://github.com/solo-io/gloo-mesh/issues/1439)


### v1.0.4

**Fixes**

- Improve the enterprise install and register documentation. (https://github.com/solo-io/gloo-mesh/issues/1451)


### v1.0.3

**Fixes**

- Fix ordering of VirtualService HttpRoutes such that routes *with matchers* precede routes without matchers. (https://github.com/solo-io/gloo-mesh/issues/1426)


### v1.1.0-beta3

**Dependency Bumps**

- solo-io/skv2 has been upgraded to v0.17.15.

**Fixes**

- Lower OSM destination detector log level to debug. (https://github.com/solo-io/gloo-mesh/issues/1439)


### v1.1.0-beta2

**Dependency Bumps**

- solo-io/skv2 has been upgraded to v0.17.14.

**Fixes**

- Fix VirtualDestination port targeting to use target_port rather than the kubernetes service exposed port. (https://github.com/solo-io/gloo-mesh/issues/1414)


### v1.1.0-beta1

**New Features**

- Add note about prefixing Helm chart values with subchart names when using the bundled Gloo Mesh Enterprise helm chart. (https://github.com/solo-io/gloo-mesh/issues/1390)
- Add guide on getting started with Gloo Mesh metrics features. (https://github.com/solo-io/gloo-mesh-enterprise/issues/308)
- Publish Helm values documentation for GME, OSS, and RBAC Webhook. (https://github.com/solo-io/gloo-mesh/issues/1382)

**Fixes**

- Re-enable docs search. (https://github.com/solo-io/gloo-mesh/issues/1195)
- Improve performance of TrafficShift decorator by using set's unsorted list method. (https://github.com/solo-io/gloo-mesh/issues/1425)
- Add API to distinguish between hostnames and ips in our ingress gateway info. (https://github.com/solo-io/gloo-mesh/issues/1415)
- Fix ordering of VirtualService HttpRoutes such that routes *with matchers* precede routes without matchers. (https://github.com/solo-io/gloo-mesh/issues/1426)


### v1.0.2

**Fixes**

- Add guide on getting started with Gloo Mesh metrics features. (https://github.com/solo-io/gloo-mesh-enterprise/issues/308)


### v1.0.1

**Fixes**

- Publish Helm values documentation for GME, OSS, and RBAC Webhook. (https://github.com/solo-io/gloo-mesh/issues/1382)


### v1.0.0

Gloo Mesh 1.0 is now available! See our docs for information on getting started with Gloo Mesh: https://docs.solo.io/gloo-mesh/latest/.
 https://docs.solo.io/gloo-mesh/latest/.

**API changes**
- Discovery (`discovery.mesh.gloo.solo.io`) version bump to v1
- Networking (`networking.mesh.gloo.solo.io`) version bump to v1
- Settings (`settings.mesh.gloo.solo.io`) version bump to v1

### v1.0.0-beta16

**New Features**

- Pre-release versions of meshctl will install the latest version of enterprise including pre-releases. (https://github.com/solo-io/gloo-mesh/issues/1362)


### v1.0.0-beta15

**New Features**

- Add step to docs generation for copying over generated Helm values doc from Gloo Mesh Enterprise. (https://github.com/solo-io/gloo-mesh-enterprise/issues/270)


### v1.0.0-beta14

**New Features**

- Allow access log records to specify additional filter state objects to include, so that filters can add arbitrary data to access logs. This is particularly useful in custom filters, such as WASM filters. (https://github.com/solo-io/gloo-mesh-enterprise/issues/165)


### v1.0.0-beta13

**Fixes**

- Set the cluster domain to `cluster.local` during enterprise registration. (https://github.com/solo-io/gloo-mesh/issues/1356)
- Remove the metrics service from networking controller. (https://github.com/solo-io/gloo-mesh/issues/1353)


### v1.0.0-beta12

**Helm Changes**

- Add `verbose` Helm value to the gloo-mesh and cert-agent helm charts, set to 'false' by default. Setting this value to 'true' will enable debug logging. (https://github.com/solo-io/gloo-mesh/issues/1347)

**New Features**

- Community now installs the matching CLI version and enterprise installs the latest version with the same major and minor number as the CLI. (https://github.com/solo-io/gloo-mesh/issues/1330)

**Fixes**

- Update the "Using Role Based API" doc to mention install time admin role. Fix RBAC docs with updated workload selector object. (https://github.com/solo-io/gloo-mesh/issues/1344)
- Fix the implementation of the `--register` flag for `meshctl install enterprise`. (https://github.com/solo-io/gloo-mesh/issues/1339)


### v1.0.0-beta11

This release contained no user-facing changes.


### v1.0.0-beta10

**New Features**

- Improve UX of registering enterprise clusters. (https://github.com/solo-io/gloo-mesh-enterprise/issues/201)


### v1.0.0-beta9

**Fixes**

- Create doc for cli plugins (https://github.com/solo-io/gloo-mesh-enterprise/issues/166)


### v0.12.8

**New Features**

- Adds better version selection logic to get the latest stable version of enterprise that matches the major/minor version of community. (https://github.com/solo-io/gloo-mesh/issues/1330)


### v1.0.0-beta8

**New Features**

- Extend the AccessLogRecord CRD with options for including HTTP request/response header/trailers. (https://github.com/solo-io/gloo-mesh-enterprise/issues/212)
- Use better formatted CLI output. (https://github.com/solo-io/gloo-mesh/issues/1302)
- Adds logic to support the forced re-initialize of the plugin manager. (https://github.com/solo-io/gloo-mesh/issues/1312)

**Fixes**

- Implement install script logic directly. (https://github.com/solo-io/gloo-mesh-enterprise/issues/188)
- Remove references to `$HOME/.meshctl` (https://github.com/solo-io/gloo-mesh-enterprise/issues/188)
- Rename the gloo-mesh-apiserver pod to dashboard, and update the `meshctl dashboard` command. (https://github.com/solo-io/gloo-mesh/issues/1276)


### v1.0.0-beta7

**New Features**

- Setup community and enterprise subcommands for install, register, and deregister. (https://github.com/solo-io/gloo-mesh-enterprise/issues/149)


### v1.0.0-beta6

**New Features**

- Add API for externalService Destination type. (https://github.com/solo-io/gloo-mesh/issues/1281)


### v1.0.0-beta5

**Breaking Changes**

- Group Kubernetes matches into its own message inside WorkloadSelector. (https://github.com/solo-io/gloo-mesh-enterprise/issues/142)


### v1.0.0-beta4

**Breaking Changes**

- Update Gloo Mesh API to 1.0 (https://github.com/solo-io/gloo-mesh-enterprise/issues/142)

**New Features**

- Add VirtualDestination as TrafficPolicy target (https://github.com/solo-io/gloo-mesh/issues/1293)

**Fixes**

- Make AccessLogRecord's filter a oneof to indicate mutual exclusiveness. (https://github.com/solo-io/gloo-mesh-enterprise/issues/167)
- Fix and simplify cluster registration docs, particularly for Kind users on MacOS. (https://github.com/solo-io/gloo-mesh/issues/1296)


### v1.0.0-beta3

**Fixes**

- Fix and simplify cluster registration docs, particularly for Kind users on MacOS. (https://github.com/solo-io/gloo-mesh/issues/1296)


### v1.0.0-beta2

**New Features**

- Add VirtualDestination API. (https://github.com/solo-io/gloo-mesh/issues/1288)

**Fixes**

- Fix translation of empty cluster name in KubeServiceAccountRefs for AccessPolicy. (https://github.com/solo-io/gloo-mesh/issues/1285)


### v1.0.0-beta1

**Breaking Changes**

- Gloo Mesh is going GA! (https://github.com/solo-io/gloo-mesh/issues/1278)


### v0.12.7

**New Features**

- Add guide for new enterprise access logging feature. (https://github.com/solo-io/gloo-mesh-enterprise/issues/103)
- Extend Role API with AccessLogRecordScope for controlling RBAC on AccessLogRecords. (https://github.com/solo-io/gloo-mesh-enterprise/issues/123)
- Add a `meshctl init-plugin-manager` which will download and install the meshctl plugin manager. (https://github.com/solo-io/gloo-mesh-enterprise/issues/129)

**Fixes**

- Remove unnecessary setup steps. Update CI setup to use enterprise-agent as local access log sync. (https://github.com/solo-io/gloo-mesh/issues/1270)
- Reset traffic policy status correctly so deleted traffic targets are removed. (https://github.com/solo-io/gloo-mesh/issues/1232)


### v0.12.6

**New Features**

- Update certificate rotation workflow to support Istio 1.9. (https://github.com/solo-io/gloo-mesh/issues/1257)

**Fixes**

- Update cluster registration doc to reference the management cluster, Istio install docs to merge duplicate keys. (https://github.com/solo-io/gloo-mesh/issues/1251)
- Transfer endpoint discovery from the TrafficTarget to the Workload (https://github.com/solo-io/gloo-mesh/issues/1255)
- Update the `meshctl version` command to include recently added components. (https://github.com/solo-io/gloo-mesh/issues/1123)


### v0.12.5

This release contained no user-facing changes.


### v0.12.4

**Dependency Bumps**

- solo-io/protoc-gen-ext has been upgraded to 0.0.15.
- solo-io/skv2 has been upgraded to 0.17.2.
- solo-io/go-utils has been upgraded to v0.20.2.

**Fixes**

- Fix the cluster registration panic which would occur when registration artifacts were already present in the registered cluster. (https://github.com/solo-io/gloo-mesh/issues/1225)
- Fix request matchers proto imports to use full path. (https://github.com/solo-io/gloo-mesh/issues/1242)
- Respect `KUBECONFIG` env var if kubeconfig is not explicitly specified through `--kubeconfig`. (https://github.com/solo-io/gloo-mesh/issues/1238)
- Do not create namespace during `meshctl install` if `--dry-run` is specified. (https://github.com/solo-io/gloo-mesh/issues/1236)


### v0.12.3

**New Features**

- Add endpoint discovery for flat networking. (https://github.com/solo-io/gloo-mesh/issues/1176)


### v0.12.2

**Dependency Bumps**

- solo-io/skv2 has been upgraded to v0.17.0.


### v0.12.1

**New Features**

- Extract the manifests for enterprise CRDs to their own chart. (https://github.com/solo-io/gloo-mesh/issues/1210)

**Fixes**

- Make `install --dry-run` output a pipe-able k8s manifest. (https://github.com/solo-io/gloo-mesh/issues/1106)


### v0.12.0

**Breaking Changes**

- Support customizable DNS suffixes for Istio meshes with smart DNS proxying enabled. (https://github.com/solo-io/gloo-mesh/issues/1179)
- Move discovery settings to its own message. (https://github.com/solo-io/gloo-mesh/issues/1186)

**New Features**

- Adds flags to the demo commands to install enterprise features. (https://github.com/solo-io/gloo-mesh/issues/1158)
- API for flat networking enterprise feature (https://github.com/solo-io/gloo-mesh/issues/1176)
- Ensure that gloo-mesh works properly with new Istio Smart DNS, and update documentation  and various scripts to make use of the feature when deploying relevant versions of Istio. (https://github.com/solo-io/gloo-mesh/issues/1179)
- meshctl's `install` and `cluster register` will update CRD definitions. (https://github.com/solo-io/gloo-mesh/issues/1079)
- Add a pre-release upgrade guide. Note that this guide is **not guaranteed** to provide a seamless upgrade experience, and that at this time we recommend users install a fresh instance of Gloo Mesh to access new functionality. (https://github.com/solo-io/gloo-mesh/issues/1157)

**Fixes**

- Don't generate changelog on release to avoid blank site for latest. (https://github.com/solo-io/gloo-mesh/issues/1182)
- Add docs for all dependent API objects. Add links between API objects. Improve API index page using generated page. (https://github.com/solo-io/gloo-mesh/issues/1153)


### v0.11.3

This release contained no user-facing changes.


### v0.11.2

**Fixes**

- Fix nil dereference in Istio mTLS translator. (https://github.com/solo-io/gloo-mesh/issues/1167)


### v0.11.1

**New Features**

- Add a toggle to Settings to enable detection and error reporting when networking translates VirtualServices or Destination rules that intersect with externally provided VirtualServices/DestinationRules. (https://github.com/solo-io/gloo-mesh/issues/1086)

**Fixes**

- Cert-Agent - Ensure that the istiod pods are restarted and ready before bouncing any other pods when handling PodBounceDirective. This ensures that gateway pods will get the correct TLS certificates for communicating with Istiod after new root certs are issued. (https://github.com/solo-io/gloo-mesh/issues/1155)
- Fix DestinationRules not being output even if they contain outlier detection config. (https://github.com/solo-io/gloo-mesh/issues/1096)


### v0.11.0

**Breaking Changes**

- Initial release of the Gloo Mesh Enterprise beta. (https://github.com/solo-io/gloo-mesh/issues/1151)

**Fixes**

- Remove search bar from docs site. (https://github.com/solo-io/gloo-mesh/issues/627)
- Generate changelog for specific versions. (https://github.com/solo-io/gloo-mesh/issues/1105)


### v0.10.10

**New Features**

- Allow the workload labels and tls port name used for Istio ingress disovery to be configurable. (https://github.com/solo-io/gloo-mesh/issues/1094)

**Fixes**

- Fallback to port protocol when failing to determine App Protocol for federation translation. (https://github.com/solo-io/gloo-mesh/issues/1146)
- Fix merge and conflict detection logic in VirtualService translator. (https://github.com/solo-io/gloo-mesh/issues/1131)
- Add virtual meshes to meshctl check output. (https://github.com/solo-io/gloo-mesh/issues/1142)
- Update validation schema on the WasmDeployment to allow filter config. (https://github.com/solo-io/gloo-mesh/issues/1148)


### v0.10.9

This release contained no user-facing changes.


### v0.10.8

**Fixes**

- Improve the user experience of installing the enterprise Wasm Agent. (https://github.com/solo-io/gloo-mesh/issues/1122)
- Relegate Helm release info to debug output. (https://github.com/solo-io/gloo-mesh/issues/1117)


### v0.10.7

**Fixes**

- Update meshctl install enterprise to be compatible with v0.2.0 of the enterprise helm chart. (https://github.com/solo-io/gloo-mesh/issues/1115)


### v0.10.6

**Fixes**

- Respect global verbose flag in meshctl, default to false. (https://github.com/solo-io/gloo-mesh/issues/1117)


### v0.10.5

**New Features**

- Create new Helm chart to bundle all CRDs required by remote agents. (https://github.com/solo-io/gloo-mesh/issues/1110)
- Allow full configurability of Settings CR in Helm Chart. (https://github.com/solo-io/gloo-mesh/issues/872)
- Write translation errors to the parents of the output resource. (https://github.com/solo-io/gloo-mesh/issues/872)
- Allows running of plugins from anywhere in the system path. (https://github.com/solo-io/gloo-mesh/issues/1101)


### v0.10.4

**Fixes**

- Fix changelog generation by inserting environment variable `RELEASE=true` in Github workflow. (https://github.com/solo-io/gloo-mesh/issues/1041)
- Use DNS-1123 compliant name for generated subset names (replace underscore delimiter with hyphen). (https://github.com/solo-io/gloo-mesh/issues/1091)


### v0.10.3

**New Features**

- Add automatic changelog generation to the doc generation step. (https://github.com/solo-io/service-mesh-hub/issues/1041)
- Add a command `meshctl dashboard` that sets up a port forward to the Gloo Mesh Enterprise dashboard if available. (https://github.com/solo-io/service-mesh-hub/issues/1053)
- Add `meshctl install enterprise` that installs Gloo Mesh Enterprise. (https://github.com/solo-io/gloo-mesh/issues/1020)
- Adds support for failover services, virtual meshes, workloads, access policies, and traffic policies for meshctl describe. (https://github.com/solo-io/service-mesh-hub/issues/1058)


### v0.10.2

This release contained no user-facing changes.


### v0.10.1

**Fixes**

- Fix name of `post_install_gloomesh.sh`. (https://github.com/solo-io/gloo-mesh/issues/1068)


### v0.10.0

**Breaking Changes**

- Rebrand from Service Mesh Hub to Gloo Mesh. (https://github.com/solo-io/service-mesh-hub/issues/1062)

**New Features**

- Add the ability to filter traffic targets and meshes by a list of search terms provided at the end of the meshctl describe command. (https://github.com/solo-io/service-mesh-hub/issues/1012)
- Add the federated DNS name to the output of `meshctl describe traffictarget`. (https://github.com/solo-io/service-mesh-hub/issues/931)

**Fixes**

- Add guide for role-based API. (https://github.com/solo-io/service-mesh-hub/issues/1061)
- Log all errors that occur during meshctl version to stderr and always show output. (https://github.com/solo-io/service-mesh-hub/issues/897)


### v0.9.2

**New Features**

- Store the appProtocol of a k8s Service if it exists on the Service. (https://github.com/solo-io/service-mesh-hub/issues/1039)
- Add the local and (if applicable) remote fully qualified domain names to the traffic target status. (https://github.com/solo-io/service-mesh-hub/issues/1035)
- Add the abiltiy to configure Service Mesh Hub with (gRPC) Extensions Servers. Extension Servers can be used to update and append output configurations produced by SMH before they are applied to managed clusters. (https://github.com/solo-io/service-mesh-hub/issues/1018)
- Update meshctl cluster register with options for installing wasm-agent. (https://github.com/solo-io/service-mesh-hub/issues/1049)

**Fixes**

- TrafficPolicy and AccessPolicy status workloads now only include workloads in the same mesh or virtual mesh of the policy's selected traffic targets. (https://github.com/solo-io/service-mesh-hub/issues/1003)
- Update TrafficPolicy translation to output VirtualServices for federated services. (https://github.com/solo-io/service-mesh-hub/issues/1002)
- Update the Istio TrafficPolicy translator to respect the WorkloadSelector's specified clusters. (https://github.com/solo-io/service-mesh-hub/issues/1045)


### v0.9.1

**Dependency Bumps**

- solo-io/go-utils has been upgraded to v0.18.1.

**New Features**

- Add the command `meshctl debug snapshot`. (https://github.com/solo-io/service-mesh-hub/issues/992)

**Fixes**

- Fix invalid values being generated for Settings in the SMH Helm Chart. (https://github.com/solo-io/service-mesh-hub/issues/1031)


### v0.9.0

**Breaking Changes**

- Update CRDs with validation schemas. Restructure TrafficPolicy.FaultInjection. (https://github.com/solo-io/service-mesh-hub/issues/512)
- VirtualMeshes now require that you have a non-null (i.e. `{}`) mtlsConfig.shared.rootCertificateAuthority.generated field. (https://github.com/solo-io/service-mesh-hub/issues/1021)

**New Features**

- Implement mesh, workload, and traffic target discovery for AWS App Mesh. Traffic Policy translation will arrive in a later release. (https://github.com/solo-io/service-mesh-hub/issues/994)
- Prior to translation of networking configuration, validate that referenced config targets exist. (https://github.com/solo-io/service-mesh-hub/issues/962)
- As part of validation referenced config targets, report non-existent TrafficTargets to TrafficPolicy status (https://github.com/solo-io/service-mesh-hub/issues/963)
- Add end to end tests for OSM v0.4.0. (https://github.com/solo-io/service-mesh-hub/issues/987)
- Introduce new Settings CRD for global settings configuration. (https://github.com/solo-io/service-mesh-hub/issues/996)

**Fixes**

- Fix for panic in meshctl when calling "describe mesh" against a cluster where OSM was discovered. (https://github.com/solo-io/service-mesh-hub/issues/1011)
- Respect kubeconfig path if specified in meshctl. (https://github.com/solo-io/service-mesh-hub/issues/1000)
- Respect KUBECONFIG env var for certain meshctl commands (e.g. version) (https://github.com/solo-io/service-mesh-hub/issues/945)
- Fix to prevent writing "null" ingress gateways on Istio mesh resources. (https://github.com/solo-io/service-mesh-hub/issues/1008)


### v0.8.1

**Fixes**

- Fix for potential panic when translating traffic policies into virtual services. (https://github.com/solo-io/service-mesh-hub/issues/979)


### v0.8.0

**Breaking Changes**

- Add ability to specify subsets when referencing a FailoverService as a traffic shift destination in TrafficPolicies. (https://github.com/solo-io/service-mesh-hub/issues/953)

**Fixes**

- Fix unnecessary pod restart caused by IssuedCertificate. (https://github.com/solo-io/service-mesh-hub/issues/975)
- Handle both istioctl version < 1.7 and >= 1.7 in `meshctl demo istio-multicluster init`. (https://github.com/solo-io/service-mesh-hub/issues/958)
- Rename "KubernertesWorkload" to "KubernetesWorkload". (https://github.com/solo-io/service-mesh-hub/issues/971)
- Rename "ServiceSelector" to "TrafficTargetSelector" (https://github.com/solo-io/service-mesh-hub/issues/965)
- Select workloads associated with a policy based on its ServiceAccounts, while checking that the workload belongs to the correct namespace and cluster. (https://github.com/solo-io/service-mesh-hub/issues/967)
- Validate workloads by updating the observed generation. (https://github.com/solo-io/service-mesh-hub/issues/973)


### v0.7.5

**New Features**

- Add section to troubleshooting doc on how to read and interpret CRD statuses for diagnostic purposes. (https://github.com/solo-io/service-mesh-hub/issues/946)
- Update traffic policy status and access policy status to include selected workloads. (https://github.com/solo-io/service-mesh-hub/issues/942)

**Fixes**

- Update management-cluster to mgmt-cluster for consistency (https://github.com/solo-io/service-mesh-hub/issues/915)


### v0.7.4

**New Features**

- Expose max_ejection_percent on TrafficPolicy's OutlierDetection config. Defaults to 100. (https://github.com/solo-io/service-mesh-hub/issues/948)

**Fixes**

- Fix Subject Names in Issued Certificates. (https://github.com/solo-io/service-mesh-hub/issues/943)


### v0.7.3

**Fixes**

- Fix EnvoyFilter incorrect cluster string when translated for local k8s service. (https://github.com/solo-io/service-mesh-hub/issues/933)
- Fix for meshctl auth errors when running against GKE clusters. (https://github.com/solo-io/service-mesh-hub/issues/919)


### v0.7.2

**New Features**

- Support Open Service Mesh (https://github.com/solo-io/service-mesh-hub/issues/893)
- Allow selecting FailoverServices as traffic shift destinations in the TrafficPolicy CRD. (https://github.com/solo-io/service-mesh-hub/issues/899)

**Fixes**

- Update the images on the concepts page to reflect v0.7.x. (https://github.com/solo-io/service-mesh-hub/issues/922)
- Update the 2 cluster image on the setup and installing Istio pages. (https://github.com/solo-io/service-mesh-hub/issues/921)
- Fix meshctl cluster register remote kubeconfig location. (https://github.com/solo-io/service-mesh-hub/issues/924)


### v0.7.1

**Fixes**

- Use packr to build meshctl so that scripts are included. (https://github.com/solo-io/service-mesh-hub/issues/892)


### v0.7.0

**Breaking Changes**

- Rename the following CRDs—“MeshService” to “TrafficTarget”, “MeshWorkload” to “Workload”, “AccessControlPolicy” to “AccessPolicy” (https://github.com/solo-io/service-mesh-hub/issues/882)
- Drop support for installing service meshes using `meshctl mesh install ...` (https://github.com/solo-io/service-mesh-hub/issues/883)
- Temporarily drop support for Appmesh and EKS cluster discovery. Support will be restored in the subsequent release. (https://github.com/solo-io/service-mesh-hub/issues/883)

**New Features**

- Add support for multicluster subset routing in TrafficPolicy for Istio meshes. (https://github.com/solo-io/service-mesh-hub/issues/671)
- Add support for Istio 1.7 (https://github.com/solo-io/service-mesh-hub/issues/817)
- Enable discovery for k8s workloads controlled by StatefulSets and Daemonsets. (https://github.com/solo-io/service-mesh-hub/issues/791)

**Fixes**

- Remove imports to offending github.com/skv2/contrib package. (https://github.com/solo-io/service-mesh-hub/issues/887)


### v0.6.1

**New Features**

- Add a tutorial (guide) for FailoverService. (https://github.com/solo-io/service-mesh-hub/issues/826)

**Fixes**

- In meshctl, silence verbose help message on non-zero exit codes. (https://github.com/solo-io/service-mesh-hub/issues/824)
- Fix copy for istio multicluster demo cmd. (https://github.com/solo-io/service-mesh-hub/issues/823)
- Fix bug where version is missing from `meshctl` and service mesh hub pods (https://github.com/solo-io/service-mesh-hub/pull/821)


### v0.6.0

**Breaking Changes**

- Rename instances of `zephyr` to `smh` in the SMH API. (https://github.com/solo-io/service-mesh-hub/issues/765)

**New Features**

- Enable access control translation on Appmesh meshes (https://github.com/solo-io/service-mesh-hub/issues/742)
- Update API with new FailoverService CRD and extend TrafficPolicy CRD with outlier detection configuration. (https://github.com/solo-io/service-mesh-hub/issues/789)
- Add in-code support for Istio 1.6. Still need to do e2e testing. (https://github.com/solo-io/service-mesh-hub/issues/719)

**Fixes**

- Distinguish workloads with same name in different namespaces. (https://github.com/solo-io/service-mesh-hub/issues/758)
- Add observed generation to traffic policies, and make sure that only policies that have been observed are aggregated. (https://github.com/solo-io/service-mesh-hub/issues/753)
- Add documentation on how to provide a structured root cert for a VM. (https://github.com/solo-io/service-mesh-hub/issues/786)
- Update setup-kind.sh is used to solve the problem of default King image running failure (istiod error) (https://github.com/solo-io/service-mesh-hub/issues/718)
- Update VirtualMesh docs with new enum for global access control enforcement. (https://github.com/solo-io/service-mesh-hub/issues/737)
- Deploy a versioned docs website that tracks released versions as well as master. (https://github.com/solo-io/service-mesh-hub/issues/811)


### v0.5.0

**Breaking Changes**

- Support varying default access control behavior depending on the mesh. (https://github.com/solo-io/service-mesh-hub/issues/721)

**New Features**

- Create a cmd to showcase SMH functionalioty for Appmesh + EKS. (https://github.com/solo-io/service-mesh-hub/issues/708)
- Document basic AWS discovery functionality for Appmesh and EKS. (https://github.com/solo-io/service-mesh-hub/issues/701)
- Expose new Settings CRD that allows user configuration of SMH discovery for AWS resources. (https://github.com/solo-io/service-mesh-hub/issues/685)
- Upon receipt of AWS credentials, SMH should automatically discover EKS clusters associated with that AWS account. (https://github.com/solo-io/service-mesh-hub/issues/665)
- Add nodeSelector for mesh-discovery, mesh-networking, and csr-agent helm templates (https://github.com/solo-io/service-mesh-hub/pull/715)
- Add `cluster deregister` command for deregistering manually registered clusters. (https://github.com/solo-io/service-mesh-hub/issues/675)
- Expose Helm chart values files option to `meshctl cluster register` for csr-agent. (https://github.com/solo-io/service-mesh-hub/issues/716)

**Fixes**

- Fix infinite event loop in cluster tenancy finder by using pure reconcile logic (i.e. reconcile the entire state cluster tenancy). (https://github.com/solo-io/service-mesh-hub/issues/657)
- Clear an invalid configstatus if things were once invalid but are now valid (https://github.com/solo-io/service-mesh-hub/issues/660)
- Use "aws-auth.kube-system" ConfigMap to determine AWS account ID. (https://github.com/solo-io/service-mesh-hub/issues/696)
- Fix setup-kind.sh to work on linux. (https://github.com/solo-io/service-mesh-hub/issues/692)
- Fix casting ttlDays in virtual_mesh_printer.go with strconv.Itoa (https://github.com/solo-io/service-mesh-hub/issues/725)
- Fix MeshService discovery not recreating MeshServices by using a pure reconcile approach. (https://github.com/solo-io/service-mesh-hub/issues/694)
- Handle deletion of MeshWorkloads and migrate to reconcile paradigm. (https://github.com/solo-io/service-mesh-hub/issues/674)
- Migrate to Docker Hub. (https://github.com/solo-io/service-mesh-hub/issues/735)
- Add random fuzz to periodic reconcilers to avoid a thundering herd problem (https://github.com/solo-io/service-mesh-hub/issues/705)
- Clean up a number of typos in user-facing messages (https://github.com/solo-io/service-mesh-hub/issues/697)


### v0.4.11

**New Features**

- Clean up resources related to access control enforcement when the relevant VirtualMesh is deleted (https://github.com/solo-io/service-mesh-hub/issues/585)
- Add initial AppMesh discovery. (https://github.com/solo-io/service-mesh-hub/issues/599)
- Add cluster tenancy finder for AppMesh, which scans pods for AppMesh injection and updates the relevant Mesh CRD with the cluster. (https://github.com/solo-io/service-mesh-hub/issues/630)
- Clean up Mesh CRs when the deployment backing them is deleted (https://github.com/solo-io/service-mesh-hub/issues/584)
- Clean up Mesh Service CRs when they need to be garbage collected (https://github.com/solo-io/service-mesh-hub/issues/584)
- Clean up Mesh Workload CRs when they need to be garbage collected (https://github.com/solo-io/service-mesh-hub/issues/584)

**Fixes**

- Don't delete all kind clusters when demo cleanup (https://github.com/solo-io/service-mesh-hub/issues/641)
- Use Upsert method for federation MeshService update. (https://github.com/solo-io/service-mesh-hub/issues/638)


### v0.4.10

**Fixes**

- Make sure that mesh workload discovery only notices meshes on the same cluster that it's watching (https://github.com/solo-io/service-mesh-hub/issues/617)


### v0.4.9

**Fixes**

- Fix the AccessControl guide. (https://github.com/solo-io/service-mesh-hub/issues/605)
- Change the default MD rendered for the docs to fix the yaml rendering. (https://github.com/solo-io/service-mesh-hub/issues/601)
- Ensure that we clean up more resources (specifically, the kubeconfig secrets) on `meshctl uninstall` (https://github.com/solo-io/service-mesh-hub/issues/593)
- Separate out CSR agent CRDs from the parent CRD chart so that we can clean them up in a more targeted way during `meshctl uninstall` (https://github.com/solo-io/service-mesh-hub/issues/603)


### v0.4.8

**Fixes**

- Actually fix the config creator so it will read the certificate data from a file if specified. (https://github.com/solo-io/service-mesh-hub/issues/590)


### v0.4.7

**Fixes**

- Fix the config creator so it will read the certificate data from a file if specified. (https://github.com/solo-io/service-mesh-hub/issues/590)
- Fix several instances of the word "Istio" getting replaced to "Mesh" in string constants (https://github.com/solo-io/service-mesh-hub/issues/564)
- Fix race in mesh workload discovery, resulting in nil Mesh refs on MeshWorkloads (https://github.com/solo-io/service-mesh-hub/issues/576)
- Fix panic in mesh-networking when an ACP selector references a nonexistent cluster (https://github.com/solo-io/service-mesh-hub/issues/580)


### v0.4.6

**Fixes**

- Bump the timeout on meshctl upgrade, which was too restrictive before. (https://github.com/solo-io/service-mesh-hub/issues/506)
- Fix bug where `meshctl uninstall` can report that a CRD was not found when the management plane cluster is registered. (https://github.com/solo-io/service-mesh-hub/issues/565)
- Fix bug where `meshctl describe` was including policies more than once because of a loop iter var getting closed over (https://github.com/solo-io/service-mesh-hub/issues/567)
- Fix bug where `meshctl describe` has confusing output when no policies apply (https://github.com/solo-io/service-mesh-hub/issues/566)
- Fix error being dropped in mesh-discovery (https://github.com/solo-io/service-mesh-hub/issues/515)
- Stop sending Helm dry-run output to /dev/null during `meshctl install` (https://github.com/solo-io/service-mesh-hub/issues/510)


### v0.4.5

**Fixes**

- Fix race where event handler on remote cluster for VirtualMesh CSR was being initialized before the VirtualMesh CRD was registered. (https://github.com/solo-io/service-mesh-hub/issues/504)
- Report "PROCESSING_ERROR" instead "INVALID" which was confusing for the user. (https://github.com/solo-io/service-mesh-hub/issues/559)


### v0.4.4

**New Features**

- Implement interactive meshctl creation of TrafficPolicy and AccessControlPolicy (https://github.com/solo-io/service-mesh-hub/issues/519)


### v0.4.3

**New Features**

- Implement Network Configuration for Linkerd (https://github.com/solo-io/mesh-projects/issues/385)


### v0.4.2

**Fixes**

- Temporarily revert to the old demo language (https://github.com/solo-io/service-mesh-hub/issues/503)


### v0.4.1

**Fixes**

- Fix kind host-networking (https://github.com/solo-io/service-mesh-hub/issues/503)


### v0.4.0

**New Features**

- Allow the user to optionally register the management plane cluster when installing SMH (https://github.com/solo-io/service-mesh-hub/issues/284)
- Implement Network Configuration for Linkerd (https://github.com/solo-io/service-mesh-hub/issues/385)
- Implement simple meshctl get * commands (https://github.com/solo-io/service-mesh-hub/issues/294)

**Fixes**

- Refactor the istio install command to be in a more generic CLI structure (https://github.com/solo-io/service-mesh-hub/issues/372)


### v0.3.25

*This release build failed.*

**Fixes**

- fix discovery and config for permissive mode (https://github.com/solo-io/supergloo/issues/484)


### v0.3.24

*This release build failed.*

This release contained no user-facing changes.


### v0.3.23

This release contained no user-facing changes.


### v0.3.22

**New Features**

- Add a new `smi-install` option to the `supergloo install istio` command to deploy the SMI Istio adapter together with Istio. (https://github.com/solo-io/supergloo/issues/458)
- enable SMI discovery and translation (https://github.com/solo-io/supergloo/issues/456)

**Fixes**

- Fix meshdiscovery reconcile (https://github.com/solo-io/supergloo/issues/448)


### v0.3.21

**New Features**

- add upgrade command to cli (https://github.com/solo-io/supergloo/issues/444)


### v0.3.20

**Fixes**

- Fix namespace blacklist to filter by installation namespace. (https://github.com/solo-io/supergloo/issues/446)


### v0.3.19

**New Features**

- enable skipping bouncing of prometheus pods with a flag for supergloo (https://github.com/solo-io/supergloo/issues/438)

**Fixes**

- Discovery no longer considers resources that are in namespaces referenced in install CRDs. (https://github.com/solo-io/supergloo/pull/440)


### v0.3.18

**Dependency Bumps**

- solo-io/solo-kit has been upgraded to v0.9.0.


### v0.3.17

**Dependency Bumps**

- solo-io/gloo has been upgraded to v0.13.20.
- solo-io/solo-kit has been upgraded to v0.8.0.
- solo-io/go-utils has been upgraded to v0.8.4.

**New Features**

- Add automated linkerd ingress with gloo. (https://github.com/solo-io/supergloo/issues/422)



