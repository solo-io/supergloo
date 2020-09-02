package registration

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/service-mesh-hub/codegen/io"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/smh"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	register.RegistrationOptions
	CertAgentInstallOptions
	Verbose bool
}

func NewRegistrant(opts *RegistrantOptions) *Registrant {
	registrant := &Registrant{opts}
	registrant.ClusterRoles = []*rbacv1.ClusterRole{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: registrant.RemoteNamespace,
				Name:      "smh-remote-access",
			},
			Rules: smhRbacRequirements,
		},
	}
	// Use management kubeconfig for remote cluster if unset.
	if registrant.RemoteKubeCfg == nil {
		registrant.RemoteKubeCfg = registrant.KubeCfg
	}
	return registrant
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
	return r.RegistrationOptions.DeregisterCluster(ctx)
}

func (r *Registrant) registerCluster(ctx context.Context) error {
	//logrus.Debugf("registering cluster with kubeConfig %+v\n", r.RegistrationOptions.KubeCfg)
	//logrus.Debugf("registering cluster with remote kubeConfig %+v\n", r.RegistrationOptions.RemoteKubeCfg)
	//logrus.Debugf("registering cluster with opts %+v\n", r.RegistrationOptions)

	if err := r.RegistrationOptions.RegisterCluster(ctx); err != nil {
		return err
	}

	logrus.Infof("successfully registered cluster %v", r.ClusterName)
	return nil
}

func (r *Registrant) installCertAgent(ctx context.Context) error {
	return smh.Installer{
		HelmChartPath:  r.CertAgentInstallOptions.ChartPath,
		HelmValuesPath: r.CertAgentInstallOptions.ChartValues,
		KubeConfig:     r.KubeCfg,
		Namespace:      r.RemoteNamespace,
		Verbose:        r.Verbose,
	}.InstallCertAgent(
		ctx,
	)
}

func (r *Registrant) uninstallCertAgent(ctx context.Context) error {
	return smh.Uninstaller{
		KubeConfig: r.KubeCfg,
		Namespace:  r.RemoteNamespace,
		Verbose:    r.Verbose,
	}.UninstallCertAgent(
		ctx,
	)
}
