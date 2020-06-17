#!/bin/sh


mockgen -package mock_controller_runtime -destination ../../test/mocks/controller-runtime/mock_manager.go sigs.k8s.io/controller-runtime/pkg/manager Manager &
mockgen -package mock_controller_runtime -destination ../../test/mocks/controller-runtime/mock_predicate.go sigs.k8s.io/controller-runtime/pkg/predicate Predicate &
mockgen -package mock_controller_runtime -destination ../../test/mocks/controller-runtime/mock_cache.go sigs.k8s.io/controller-runtime/pkg/cache Cache &
mockgen -package mock_controller_runtime -destination ../../test/mocks/controller-runtime/mock_dynamic_client.go  sigs.k8s.io/controller-runtime/pkg/client Client,StatusWriter &
mockgen -package mock_cli_runtime -destination ../../test/mocks/cli_runtime/mock_rest_client_getter.go k8s.io/cli-runtime/pkg/resource RESTClientGetter &
mockgen -package mock_multicluster -destination ../../test/mocks/smh/clients/mock_multicluster_client.go github.com/solo-io/skv2/pkg/multicluster Client &

# K8s miscellaneous interfaces
mockgen -package mock_k8s_cliendcmd -destination ../../test/mocks/client-go/clientcmd/client_config.go k8s.io/client-go/tools/clientcmd ClientConfig &

# AppMesh clients
mockgen -package mock_appmesh_clients -destination ../../test/mocks/clients/aws/appmesh/clients.go github.com/aws/aws-sdk-go/service/appmesh/appmeshiface AppMeshAPI &

wait
