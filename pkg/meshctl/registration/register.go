package registration

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/service-mesh-hub/codegen/io"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/smh"
	"github.com/solo-io/skv2/pkg/multicluster/kubeconfig"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var smhRbacRequirements = func() []rbacv1.PolicyRule {
	var policyRules []rbacv1.PolicyRule
	policyRules = append(policyRules, io.DiscoveryInputTypes.RbacPoliciesWatch()...)
	policyRules = append(policyRules, io.LocalNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	policyRules = append(policyRules, io.IstioNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	policyRules = append(policyRules, io.SmiNetworkingOutputTypes.Snapshot.RbacPoliciesWrite()...)
	policyRules = append(policyRules, io.CertificateIssuerInputTypes.RbacPoliciesWatch()...)
	policyRules = append(policyRules, io.CertificateIssuerInputTypes.RbacPoliciesUpdateStatus()...)

	return policyRules
}()

type Registrant struct {
	*RegistrantOptions
}

type RegistrantOptions struct {
	KubeConfigPath string
	MgmtContext    string
	RemoteContext  string
	Registration   register.RegistrationOptions
	CertAgent      CertAgentInstallOptions
	Verbose        bool
}

// Initialize a ClientConfig for the management and remote clusters from the options.
func (m *RegistrantOptions) ConstructClientConfigs() (mgmtKubeCfg clientcmd.ClientConfig, remoteKubeCfg clientcmd.ClientConfig, err error) {
	mgmtKubeCfg, err = kubeconfig.GetClientConfigWithContext(m.KubeConfigPath, m.MgmtContext)
	if err != nil {
		return nil, nil, err
	}
	remoteKubeCfg, err = kubeconfig.GetClientConfigWithContext(m.KubeConfigPath, m.RemoteContext)
	if err != nil {
		return nil, nil, err
	}
	return mgmtKubeCfg, remoteKubeCfg, nil
}

func NewRegistrant(opts *RegistrantOptions) (*Registrant, error) {
	registrant := &Registrant{opts}
	registrant.Registration.ClusterRoles = []*rbacv1.ClusterRole{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: registrant.Registration.RemoteNamespace,
				Name:      "smh-remote-access",
			},
			Rules: smhRbacRequirements,
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
	return registrant, nil
}

type CertAgentInstallOptions struct {
	ChartPath   string
	ChartValues string
}

func (r *Registrant) RegisterCluster(ctx context.Context) error {
	// TODO(ilackarms): move verbose option to global flag at root level of meshctl
	if r.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if err := r.installCertAgent(ctx); err != nil {
		return err
	}
	return r.registerCluster(ctx)
}

func (r *Registrant) DeregisterCluster(ctx context.Context) error {
	if r.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if err := r.uninstallCertAgent(ctx); err != nil {
		return err
	}
	return r.Registration.DeregisterCluster(ctx)
}

func (r *Registrant) registerCluster(ctx context.Context) error {
	logrus.Debugf("registering cluster with opts %+v\n", r.Registration)

	if err := r.Registration.RegisterCluster(ctx); err != nil {
		return err
	}

	logrus.Infof("successfully registered cluster %v", r.Registration.ClusterName)
	return nil
}

func (r *Registrant) installCertAgent(ctx context.Context) error {
	return smh.Installer{
		HelmChartPath:  r.CertAgent.ChartPath,
		HelmValuesPath: r.CertAgent.ChartValues,
		KubeConfig:     r.KubeConfigPath,
		KubeContext:    r.RemoteContext,
		Namespace:      r.Registration.RemoteNamespace,
		Verbose:        r.Verbose,
	}.InstallCertAgent(
		ctx,
	)
}

func (r *Registrant) uninstallCertAgent(ctx context.Context) error {
	return smh.Uninstaller{
		KubeConfig:  r.KubeConfigPath,
		KubeContext: r.RemoteContext,
		Namespace:   r.Registration.RemoteNamespace,
		Verbose:     r.Verbose,
	}.UninstallCertAgent(
		ctx,
	)
}
