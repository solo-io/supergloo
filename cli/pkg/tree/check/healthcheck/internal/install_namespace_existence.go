package internal

import (
	"context"
	"fmt"

	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
	"k8s.io/apimachinery/pkg/api/errors"
)

func NewInstallNamespaceExistenceCheck() healthcheck_types.HealthCheck {
	return &installNamespaceExistence{}
}

type installNamespaceExistence struct{}

func (*installNamespaceExistence) GetDescription() string {
	return "installation namespace exists"
}

func (i *installNamespaceExistence) Run(ctx context.Context, installNamespace string, clients healthcheck_types.Clients) (runFailure *healthcheck_types.RunFailure, checkApplies bool) {
	_, err := clients.NamespaceClient.Get(ctx, installNamespace)
	if errors.IsNotFound(err) {
		return &healthcheck_types.RunFailure{
			ErrorMessage: NamespaceDoesNotExist(installNamespace).Error(),
			Hint:         fmt.Sprintf("try running `kubectl create namespace %s`", installNamespace),
		}, true
	} else if err != nil {
		return &healthcheck_types.RunFailure{
			ErrorMessage: GenericCheckFailed(err).Error(),
			Hint:         "make sure the Kubernetes API server is reachable",
		}, true
	}
	return nil, true
}
