package checks

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	corev1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	apps_v1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type deploymentsCheck struct{}

func NewDeploymentsCheck() Check {
	return &deploymentsCheck{}
}

func (d *deploymentsCheck) GetDescription() string {
	return "Gloo Mesh pods are running"
}

func (d *deploymentsCheck) Run(ctx context.Context, checkCtx CheckContext) *Failure {
	installNamespace := checkCtx.Environment().Namespace
	namespaceClient := corev1.NewNamespaceClient(checkCtx.Client())
	_, err := namespaceClient.GetNamespace(ctx, installNamespace)
	if err != nil {
		return &Failure{
			Errors: []error{eris.Wrapf(err, "specified namespace %s doesn't exist", installNamespace)},
		}
	}
	deploymentClient := appsv1.NewDeploymentClient(checkCtx.Client())
	deployments, err := deploymentClient.ListDeployment(ctx, client.InNamespace(installNamespace))
	if err != nil {
		return &Failure{
			Errors: []error{err},
		}
	}

	return d.checkDeployments(deployments, checkCtx.Environment())
}

func (d *deploymentsCheck) checkDeployments(deployments *apps_v1.DeploymentList, env Environment) *Failure {
	installNamespace := env.Namespace
	failure := new(Failure)
	if len(deployments.Items) < 1 {
		failure.AddError(eris.Errorf("no deployments found in namespace %s", installNamespace))
		if !env.InCluster {
			failure.AddHint(fmt.Sprintf(
				`Gloo Mesh installation namespace can be supplied to this cmd with the "--namespace" flag, which defaults to %s`,
				defaults.DefaultPodNamespace), "")
		}
		return failure
	}

	for _, deployment := range deployments.Items {
		if deployment.Status.AvailableReplicas < 1 {
			failure.AddError(eris.Errorf(`deployment "%s" has no available pods`, deployment.Name))
		}
	}
	if len(failure.Errors) > 0 {
		failure.AddHint(d.buildHint(installNamespace), "")
	}
	return failure
}

func (d *deploymentsCheck) buildHint(installNamespace string) string {
	return fmt.Sprintf(`check the status of Gloo Mesh deployments with "kubectl -n %s get deployments -oyaml"`, installNamespace)
}
