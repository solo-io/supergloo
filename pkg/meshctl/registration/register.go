package registration

import (
	"context"
	"fmt"

	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	appsv1 "github.com/solo-io/external-apis/pkg/api/k8s/apps/v1"
	"github.com/solo-io/gloo-mesh/codegen/io"
	"github.com/solo-io/gloo-mesh/pkg/mesh-discovery/utils/dockerutils"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var gloomeshRbacRequirements = func() []rbacv1.PolicyRule {
	var policyRules []rbacv1.PolicyRule
	policyRules = append(policyRules, io.DiscoveryInputTypes.RbacPoliciesWatch()...)
	policyRules = append(policyRules, io.LocalNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	policyRules = append(policyRules, io.IstioNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	policyRules = append(policyRules, io.SmiNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	policyRules = append(policyRules, io.CertificateIssuerInputTypes.RbacPoliciesWatch()...)
	policyRules = append(policyRules, io.CertificateIssuerInputTypes.RbacPoliciesUpdateStatus()...)

	return policyRules
}()

type RegistrantOptions struct {
	KubeConfigPath         string
	MgmtContext            string
	RemoteContext          string
	Registration           register.RegistrationOptions
	AgentCrdsChartPath     string
	AgentChartPathOverride string
	AgentChartValues       string
	Verbose                bool
}

// Initialize a ClientConfig for the management and remote clusters from the options.
func (m *RegistrantOptions) ConstructClientConfigs() (mgmtKubeCfg, remoteKubeCfg clientcmd.ClientConfig, err error) {
	mgmtKubeCfg, err = kubeconfig.GetClientConfigWithContext(m.KubeConfigPath, m.MgmtContext, "")
	if err != nil {
		return nil, nil, err
	}
	remoteKubeCfg, err = kubeconfig.GetClientConfigWithContext(m.KubeConfigPath, m.RemoteContext, "")
	if err != nil {
		return nil, nil, err
	}
	return mgmtKubeCfg, remoteKubeCfg, nil
}

type Registrant struct {
	// Optionally set a version manually
	VersionOverride string

	opts              RegistrantOptions
	agentReleaseName  string
	agentChartPathTpl string // Will replace single %s with version
}

func NewRegistrant(opts RegistrantOptions, agentReleaseName, agentChartPathTpl string) (*Registrant, error) {
	registrant := &Registrant{"", opts, agentReleaseName, agentChartPathTpl}
	registrant.opts.Registration.ClusterRoles = []*rbacv1.ClusterRole{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: registrant.opts.Registration.RemoteNamespace,
				Name:      "gloomesh-remote-access",
			},
			Rules: gloomeshRbacRequirements,
		},
	}
	// Convert kubeconfig path and context into ClientConfig for Registration
	mgmtClientConfig, remoteClientConfig, err := registrant.opts.ConstructClientConfigs()
	if err != nil {
		return nil, err
	}
	registrant.opts.Registration.KubeCfg = mgmtClientConfig
	registrant.opts.Registration.RemoteKubeCfg = remoteClientConfig
	// We need to explicitly pass the remote context because of this open issue: https://github.com/kubernetes/client-go/issues/735
	registrant.opts.Registration.RemoteCtx = opts.RemoteContext
	return registrant, nil
}

func (r *Registrant) RegisterCluster(ctx context.Context) error {
	// agent CRDs should always be installed since they're required by any remote agents
	if err := r.installAgentCrds(ctx); err != nil {
		return err
	}

	if err := r.installAgent(ctx); err != nil {
		return err
	}

	return r.registerCluster(ctx)
}

func (r *Registrant) registerCluster(ctx context.Context) error {
	logrus.Debugf("registering cluster with opts %+v\n", r.opts.Registration)

	if err := r.opts.Registration.RegisterCluster(ctx); err != nil {
		return err
	}

	logrus.Infof("successfully registered cluster %v", r.opts.Registration.ClusterName)
	return nil
}

func (r *Registrant) installAgentCrds(ctx context.Context) error {
	chartPath, err := r.getChartPath(ctx, r.opts.AgentCrdsChartPath, gloomesh.AgentCrdsChartUriTemplate)
	if err != nil {
		return err
	}

	return helm.Installer{
		KubeConfig:  r.opts.KubeConfigPath,
		KubeContext: r.opts.RemoteContext,
		ChartUri:    chartPath,
		Namespace:   r.opts.Registration.RemoteNamespace,
		ReleaseName: gloomesh.AgentCrdsReleaseName,
		Verbose:     r.opts.Verbose,
	}.InstallChart(ctx)
}

func (r *Registrant) installAgent(ctx context.Context) error {
	chartPath, err := r.getChartPath(ctx, r.opts.AgentChartPathOverride, r.agentChartPathTpl)
	if err != nil {
		return err
	}

	return helm.Installer{
		KubeConfig:  r.opts.KubeConfigPath,
		KubeContext: r.opts.RemoteContext,
		ChartUri:    chartPath,
		ValuesFile:  r.opts.AgentChartValues,
		Namespace:   r.opts.Registration.RemoteNamespace,
		ReleaseName: gloomesh.AgentCrdsReleaseName,
		Verbose:     r.opts.Verbose,
	}.InstallChart(ctx)
}

func (r *Registrant) getChartPath(ctx context.Context, pathOverride, pathTemplate string) (string, error) {
	// Use manual chart override path first
	if pathOverride != "" {
		return r.opts.AgentCrdsChartPath, nil
	}
	// Then use manually set version
	if r.VersionOverride != "" {
		return fmt.Sprintf(pathTemplate, r.VersionOverride), nil
	}

	// Lastly, determine version from Gloo Mesh deployment
	version, err := r.getGlooMeshVersion(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(pathTemplate, version), nil
}

func (r *Registrant) getGlooMeshVersion(ctx context.Context) (string, error) {
	kubeClient, err := utils.BuildClient(r.opts.KubeConfigPath, r.opts.MgmtContext)
	if err != nil {
		return "", err
	}

	depClient := appsv1.NewDeploymentClient(kubeClient)
	deployments, err := depClient.ListDeployment(ctx, &client.ListOptions{Namespace: r.opts.Registration.Namespace})
	if err != nil {
		return "", err
	}

	// Find the (enterprise-)networking deployment and return the tag of the
	// gloo-mesh image, in the event of multiple containers, the name of the
	// main container will be (enterprise-)networking as well.
	for _, deployment := range deployments.Items {
		if deployment.Name == "networking" || deployment.Name == "enterprise-networking" {
			for _, container := range deployment.Spec.Template.Spec.Containers {
				if container.Name == deployment.Name {
					image, err := dockerutils.ParseImageName(container.Image)
					if err != nil {
						return "", err
					}
					return image.Tag, err
				}
			}
		}
	}

	return "", eris.New("unable to find Gloo Mesh deployment in management cluster")
}

func (r *Registrant) DeregisterCluster(ctx context.Context) error {
	if err := r.uninstallAgentCrds(ctx); err != nil {
		return err
	}

	if err := r.uninstallCertAgent(ctx); err != nil {
		return err
	}

	return r.opts.Registration.DeregisterCluster(ctx)
}

func (r *Registrant) uninstallAgentCrds(ctx context.Context) error {
	return helm.Uninstaller{
		KubeConfig:  r.opts.KubeConfigPath,
		KubeContext: r.opts.RemoteContext,
		ReleaseName: gloomesh.AgentCrdsReleaseName,
		Namespace:   r.opts.Registration.RemoteNamespace,
		Verbose:     r.opts.Verbose,
	}.UninstallChart(ctx)
}

func (r *Registrant) uninstallCertAgent(ctx context.Context) error {
	return helm.Uninstaller{
		KubeConfig:  r.opts.KubeConfigPath,
		KubeContext: r.opts.RemoteContext,
		ReleaseName: r.agentReleaseName,
		Namespace:   r.opts.Registration.RemoteNamespace,
		Verbose:     r.opts.Verbose,
	}.UninstallChart(ctx)
}
