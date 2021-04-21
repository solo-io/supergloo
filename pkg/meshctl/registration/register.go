package registration

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/gloo-mesh/codegen/io"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/helm"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
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

type Options struct {
	KubeConfigPath         string
	MgmtContext            string
	MgmtNamespace          string
	RemoteContext          string
	RemoteNamespace        string
	Version                string
	AgentCrdsChartPath     string
	AgentChartPathOverride string
	AgentChartValuesPath   string
	ApiServerAddress       string
	ClusterName            string
	ClusterDomain          string
	Verbose                bool
}

func (o Options) GetChartPath(ctx context.Context, override, template string) (string, error) {
	// Return the override path if set
	if override != "" {
		return override, nil
	}
	// If the user provides a version, use it when building the chart path
	if o.Version != "" {
		return fmt.Sprintf(template, o.Version), nil
	}
	// Otherwise find the version of Gloo Mesh that's running
	version, err := utils.GetGlooMeshVersion(ctx, o.KubeConfigPath, o.MgmtContext, o.MgmtNamespace)
	if err != nil {
		return "", err
	}
	o.Version = version // Cache the version
	return fmt.Sprintf(template, version), nil
}

// Initialize a ClientConfig for the management and remote clusters from the options.
func (o Options) ConstructClientConfigs() (mgmtKubeCfg, remoteKubeCfg clientcmd.ClientConfig, err error) {
	mgmtKubeCfg, err = kubeconfig.GetClientConfigWithContext(o.KubeConfigPath, o.MgmtContext, "")
	if err != nil {
		return nil, nil, err
	}
	remoteKubeCfg, err = kubeconfig.GetClientConfigWithContext(o.KubeConfigPath, o.RemoteContext, "")
	if err != nil {
		return nil, nil, err
	}

	return mgmtKubeCfg, remoteKubeCfg, nil
}

type Registrant struct {
	Options
	Registration register.RegistrationOptions
}

func NewRegistrant(opts Options) (*Registrant, error) {
	registrant := &Registrant{Options: opts}
	registrant.Registration.ClusterRoles = []*rbacv1.ClusterRole{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: registrant.Registration.RemoteNamespace,
				Name:      "gloomesh-remote-access",
			},
			Rules: gloomeshRbacRequirements,
		},
	}
	// Convert kubeconfig path and context into ClientConfig for Registration
	mgmtClientConfig, remoteClientConfig, err := registrant.ConstructClientConfigs()
	if err != nil {
		return nil, err
	}
	registrant.Registration.KubeCfg = mgmtClientConfig
	registrant.Registration.RemoteKubeCfg = remoteClientConfig
	// We need to explicitly pass the remote context because of this open issue: https://github.com/kubernetes/client-go/issues/735
	registrant.Registration.RemoteCtx = opts.RemoteContext
	registrant.Registration.Namespace = opts.MgmtNamespace
	registrant.Registration.RemoteNamespace = opts.RemoteNamespace
	registrant.Registration.ClusterName = opts.ClusterName
	registrant.Registration.ClusterDomain = opts.ClusterDomain
	registrant.Registration.APIServerAddress = opts.ApiServerAddress
	return registrant, nil
}

func (r *Registrant) RegisterCluster(ctx context.Context) error {
	// agent CRDs should always be installed since they're required by any remote agents
	if err := r.installAgentCrds(ctx); err != nil {
		return err
	}
	if err := r.installCertAgent(ctx); err != nil {
		return err
	}
	return r.registerCluster(ctx)
}

func (r *Registrant) registerCluster(ctx context.Context) error {
	logrus.Debugf("%+v", r.Registration)
	if err := r.Registration.RegisterCluster(ctx); err != nil {
		return err
	}
	logrus.Infof("Successfully registered cluster: %s", r.Registration.ClusterName)
	return nil
}

func (r *Registrant) deregisterCluster(ctx context.Context) error {
	logrus.Debugf("%+v", r.Registration)
	if err := r.Registration.DeregisterCluster(ctx); err != nil {
		return err
	}
	logrus.Infof("Successfully deregistered cluster: %s", r.Registration.ClusterName)
	return nil
}

func (r *Registrant) installAgentCrds(ctx context.Context) error {
	chartPath, err := r.GetChartPath(ctx, r.AgentCrdsChartPath, gloomesh.AgentCrdsChartUriTemplate)
	if err != nil {
		return err
	}
	return helm.Installer{
		KubeConfig:  r.KubeConfigPath,
		KubeContext: r.RemoteContext,
		ChartUri:    chartPath,
		Namespace:   r.Registration.RemoteNamespace,
		ReleaseName: gloomesh.AgentCrdsReleaseName,
		Verbose:     r.Verbose,
	}.InstallChart(ctx)
}

func (r *Registrant) installCertAgent(ctx context.Context) error {
	chartPath, err := r.GetChartPath(ctx, r.AgentChartPathOverride, gloomesh.CertAgentChartUriTemplate)
	if err != nil {
		return err
	}

	return helm.Installer{
		KubeConfig:  r.KubeConfigPath,
		KubeContext: r.RemoteContext,
		ChartUri:    chartPath,
		ValuesFile:  r.AgentChartValuesPath,
		Namespace:   r.Registration.RemoteNamespace,
		ReleaseName: gloomesh.CertAgentReleaseName,
		Verbose:     r.Verbose,
	}.InstallChart(ctx)
}

func (r *Registrant) DeregisterCluster(ctx context.Context) error {
	if err := r.uninstallAgentCrds(ctx); err != nil {
		return err
	}
	if err := r.uninstallCertAgent(ctx); err != nil {
		return err
	}
	return r.deregisterCluster(ctx)
}

func (r *Registrant) uninstallAgentCrds(ctx context.Context) error {
	return helm.Uninstaller{
		KubeConfig:  r.KubeConfigPath,
		KubeContext: r.RemoteContext,
		ReleaseName: gloomesh.AgentCrdsReleaseName,
		Namespace:   r.Registration.RemoteNamespace,
		Verbose:     r.Verbose,
	}.UninstallChart(ctx)
}

func (r *Registrant) uninstallCertAgent(ctx context.Context) error {
	return helm.Uninstaller{
		KubeConfig:  r.KubeConfigPath,
		KubeContext: r.RemoteContext,
		ReleaseName: gloomesh.CertAgentReleaseName,
		Namespace:   r.Registration.RemoteNamespace,
		Verbose:     r.Verbose,
	}.UninstallChart(ctx)
}
