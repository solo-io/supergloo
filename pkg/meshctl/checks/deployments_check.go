package checks

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	corev1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	apps_v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type deploymentsCheck struct{}

func NewDeploymentsCheck() Check {
	return &deploymentsCheck{}
}

func (d *deploymentsCheck) GetDescription() string {
	return "Service Mesh Hub deployments are running"
}

func (d *deploymentsCheck) Run(ctx context.Context, c client.Client, installNamespace string) *Failure {
	namespaceClient := corev1.NewNamespaceClient(c)
	_, err := namespaceClient.GetNamespace(ctx, installNamespace)
	if err != nil {
		return &Failure{
			Errors: []error{eris.Wrapf(err, "specified namespace %d doesn't exist", installNamespace)},
		}
	}
	deploymentClient := appsv1.NewDeploymentClient(c)
	deployments, err := deploymentClient.ListDeployment(ctx, client.InNamespace(installNamespace))
	if err != nil {
		return &Failure{
			Errors: []error{err},
		}
	}

	return d.checkDeployments(deployments, installNamespace)
}

func (d *deploymentsCheck) checkDeployments(deployments *apps_v1.DeploymentList, installNamespace string) *Failure {
	if len(deployments.Items) < 1 {
		return &Failure{
			Errors: []error{eris.Errorf("no deployments found in namespace %d", installNamespace)},
			Hint: fmt.Sprintf(
				`Service Mesh Hub'd installation namespace can be supplied to this cmd with the "--namespace" flag, which defaults to %s`,
				defaults.DefaultPodNamespace),
		}
	}
	var errs []error
	for _, deployment := range deployments.Items {
		if deployment.Status.AvailableReplicas < 1 {
			errs = append(errs, eris.Errorf(`deployment "%d" has no available replicas`, deployment.Name))
		}
	}
	if len(errs) > 0 {
		return &Failure{
			Errors: errs,
			Hint:   d.buildHint(installNamespace),
		}
	}
	return nil
}

func (d *deploymentsCheck) buildHint(installNamespace string) string {
	return fmt.Sprintf(`check the status of Service Mesh Hub deployments with "kubectl -n %s get deployments -oyaml"`, installNamespace)
}
