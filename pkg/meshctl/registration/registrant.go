package registration

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/service-mesh-hub/codegen/io"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/smh"
	"github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1"
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

//go:generate mockgen -source ./registrant.go -destination ./mocks/mock_registrant.go

type ClusterRegistrant interface {
	RegisterCluster(ctx context.Context) error
	RegisterProviderCluster(ctx context.Context, providerInfo *v1alpha1.KubernetesClusterSpec_ProviderInfo) error
	DeregisterCluster(ctx context.Context) error
}

type clusterRegistrant struct {
	*RegistrantOptions
}

type RegistrantOptions struct {
	register.RegistrationOptions
	CertAgentInstallOptions
	Verbose bool
}

func NewRegistrant(opts *RegistrantOptions) ClusterRegistrant {
	registrant := &clusterRegistrant{opts}
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

func (r *clusterRegistrant) RegisterCluster(ctx context.Context) error {
	// TODO(ilackarms): move verbose option to global flag at root level of meshctl
	if r.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if err := r.installCertAgent(ctx); err != nil {
		return err
	}
	return r.registerCluster(ctx, nil)
}

func (r *clusterRegistrant) RegisterProviderCluster(
	ctx context.Context,
	providerInfo *v1alpha1.KubernetesClusterSpec_ProviderInfo,
) error {
	// TODO(ilackarms): move verbose option to global flag at root level of meshctl
	if r.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if err := r.installCertAgent(ctx); err != nil {
		return err
	}
	return r.registerCluster(ctx, providerInfo)
}

func (r *clusterRegistrant) DeregisterCluster(ctx context.Context) error {
	if r.Verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}
	if err := r.uninstallCertAgent(ctx); err != nil {
		return err
	}
	return r.RegistrationOptions.DeregisterCluster(ctx)
}

func (r *clusterRegistrant) registerCluster(
	ctx context.Context,
	providerInfo *v1alpha1.KubernetesClusterSpec_ProviderInfo,
) error {
	logrus.Debugf("registering cluster with opts %+v\n", r.RegistrationOptions)

	if err := r.RegistrationOptions.RegisterProviderCluster(ctx, providerInfo); err != nil {
		return err
	}

	logrus.Infof("successfully registered cluster %v", r.ClusterName)
	return nil
}

func (r *clusterRegistrant) installCertAgent(ctx context.Context) error {
	return smh.Installer{
		HelmChartPath:  r.CertAgentInstallOptions.ChartPath,
		HelmValuesPath: r.CertAgentInstallOptions.ChartValues,
		KubeConfig:     r.RemoteKubeCfg,
		Namespace:      r.RemoteNamespace,
		Verbose:        r.Verbose,
	}.InstallCertAgent(
		ctx,
	)
}

func (r *clusterRegistrant) uninstallCertAgent(ctx context.Context) error {
	return smh.Uninstaller{
		KubeConfig: r.RemoteKubeCfg,
		Namespace:  r.RemoteNamespace,
		Verbose:    r.Verbose,
	}.UninstallCertAgent(
		ctx,
	)
}
