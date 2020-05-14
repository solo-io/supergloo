#!/bin/sh

mockgen -package mock_controller_runtime -destination ../../test/mocks/controller-runtime/mock_manager.go sigs.k8s.io/controller-runtime/pkg/manager Manager &
mockgen -package mock_controller_runtime -destination ../../test/mocks/controller-runtime/mock_predicate.go sigs.k8s.io/controller-runtime/pkg/predicate Predicate &
mockgen -package mock_controller_runtime -destination ../../test/mocks/controller-runtime/mock_cache.go sigs.k8s.io/controller-runtime/pkg/cache Cache &
mockgen -package mock_controller_runtime -destination ../../test/mocks/controller-runtime/mock_dynamic_client.go  sigs.k8s.io/controller-runtime/pkg/client Client,StatusWriter &
mockgen -package mock_cli_runtime -destination ../../test/mocks/cli_runtime/mock_rest_client_getter.go k8s.io/cli-runtime/pkg/resource RESTClientGetter &
mockgen -package mock_corev1 -destination ../../test/mocks/corev1/mock_service_controller.go github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1/controller ServiceEventWatcher &
mockgen -package mock_zephyr_discovery -destination ../../test/mocks/zephyr/discovery/mock_mesh_workload_controller.go github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/controller MeshWorkloadEventWatcher,MeshServiceEventWatcher,MeshEventWatcher &
mockgen -package mock_zephyr_networking -destination ../../test/mocks/zephyr/networking/mock_virtual_mesh_controller.go github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/controller VirtualMeshEventWatcher,TrafficPolicyEventWatcher,AccessControlPolicyEventWatcher &

# K8s clients
mockgen -package mock_k8s_core_clients -destination ../../test/mocks/clients/kubernetes/core/v1/clients.go github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1 ServiceClient,PodClient,NamespaceClient,NodeClient,ServiceAccountClient,SecretClient,ConfigMapClient &
mockgen -package mock_k8s_apps_clients -destination ../../test/mocks/clients/kubernetes/apps/v1/clients.go github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apps/v1 DeploymentClient,ReplicaSetClient &
mockgen -package mock_k8s_extension_clients -destination ../../test/mocks/clients/kubernetes/apiextensions.k8s.io/v1beta1/clients.go github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/apiextensions.k8s.io/v1beta1 CustomResourceDefinitionClient &
# K8s miscellaneous interfaces
mockgen -package mock_k8s_cliendcmd -destination ../../test/mocks/client-go/clientcmd/client_config.go k8s.io/client-go/tools/clientcmd ClientConfig &

# Zephyr clients
mockgen -package mock_zephyr_discovery_clients -destination ../../test/mocks/clients/discovery.zephyr.solo.io/v1alpha1/clients.go github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1 KubernetesClusterClient,MeshClient,MeshServiceClient,MeshWorkloadClient &
mockgen -package mock_zephyr_networking_clients -destination ../../test/mocks/clients/networking.zephyr.solo.io/v1alpha1/clients.go github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1 TrafficPolicyClient,AccessControlPolicyClient,VirtualMeshClient &
mockgen -package mock_zephyr_security_clients -destination ../../test/mocks/clients/security.zephyr.solo.io/v1alpha1/clients.go github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1 VirtualMeshCertificateSigningRequestClient &
mockgen -package mock_zephyr_settings_clients -destination ../../test/mocks/clients/settings.zephyr.solo.io/v1alpha1/clients.go github.com/solo-io/service-mesh-hub/pkg/api/settings.zephyr.solo.io/v1alpha1 SettingsClient &

# Istio clients
mockgen -package mock_istio_security_clients -destination ../../test/mocks/clients/istio/security/v1alpha3/clients.go github.com/solo-io/service-mesh-hub/pkg/api/istio/security/v1beta1 AuthorizationPolicyClient &
mockgen -package mock_istio_networking_clients -destination ../../test/mocks/clients/istio/networking/v1beta1/clients.go github.com/solo-io/service-mesh-hub/pkg/api/istio/networking/v1alpha3 DestinationRuleClient,EnvoyFilterClient,GatewayClient,ServiceEntryClient,VirtualServiceClient &

# Linkerd clients
mockgen -package mock_linkerd_clients -destination ../../test/mocks/clients/linkerd/v1alpha2/clients.go github.com/solo-io/service-mesh-hub/pkg/api/linkerd/v1alpha2 ServiceProfileClient &

# SMI clients
mockgen -package mock_smi_clients -destination ../../test/mocks/clients/smi/split/v1alpha1/clients.go github.com/solo-io/service-mesh-hub/pkg/api/smi/split/v1alpha1 TrafficSplitClient &

# AppMesh clients
mockgen -package mock_appmesh_clients -destination ../../test/mocks/clients/aws/appmesh/clients.go github.com/aws/aws-sdk-go/service/appmesh/appmeshiface AppMeshAPI &

wait
