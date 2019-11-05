|Option|Type|Default Value|Description|
|------|----|-----------|-------------|
|meshDiscovery.disabled|bool|false||
|meshDiscovery.deployment.image.tag|string|v0.0.2-16-gc7a76c1|tag for the container|
|meshDiscovery.deployment.image.repository|string|mc-mesh-discovery|image name (repository) for the container.|
|meshDiscovery.deployment.image.registry|string||image prefix/registry e.g. (quay.io/solo-io)|
|meshDiscovery.deployment.image.pullPolicy|string||image pull policy for the container|
|meshDiscovery.deployment.image.pullSecret|string||image pull policy for the container |
|meshDiscovery.deployment.stats|bool|true|enable prometheus stats|
|meshDiscovery.deployment.replicas|int|1|number of instances to deploy|
|meshDiscovery.deployment.resources.limits.memory|string||amount of memory|
|meshDiscovery.deployment.resources.limits.cpu|string||amount of CPUs|
|meshDiscovery.deployment.resources.requests.memory|string||amount of memory|
|meshDiscovery.deployment.resources.requests.cpu|string||amount of CPUs|
|meshBridge.disabled|bool|false||
|meshBridge.deployment.image.tag|string|v0.0.2-16-gc7a76c1|tag for the container|
|meshBridge.deployment.image.repository|string|mc-mesh-bridge|image name (repository) for the container.|
|meshBridge.deployment.image.registry|string||image prefix/registry e.g. (quay.io/solo-io)|
|meshBridge.deployment.image.pullPolicy|string||image pull policy for the container|
|meshBridge.deployment.image.pullSecret|string||image pull policy for the container |
|meshBridge.deployment.stats|bool|true|enable prometheus stats|
|meshBridge.deployment.replicas|int|1|number of instances to deploy|
|meshBridge.deployment.resources.limits.memory|string||amount of memory|
|meshBridge.deployment.resources.limits.cpu|string||amount of CPUs|
|meshBridge.deployment.resources.requests.memory|string||amount of memory|
|meshBridge.deployment.resources.requests.cpu|string||amount of CPUs|
|global.image.tag|string||tag for the container|
|global.image.repository|string||image name (repository) for the container.|
|global.image.registry|string|quay.io/solo-io|image prefix/registry e.g. (quay.io/solo-io)|
|global.image.pullPolicy|string|IfNotPresent|image pull policy for the container|
|global.image.pullSecret|string||image pull policy for the container |
|global.rbac.create|bool|true|create rbac rules for the gloo-system service account|
|global.rbac.Namespaced|bool|false|use Roles instead of ClusterRoles|
|global.crds.create|bool|true|create CRDs for MeshDiscovery (turn off if installing with Helm to a cluster that already has MeshDiscovery CRDs)|
