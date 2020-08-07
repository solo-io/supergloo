package register

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/install/agent"

	"github.com/solo-io/service-mesh-hub/codegen/io"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/skv2/pkg/multicluster/register"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var smhRbacRequirements = func() []rbacv1.PolicyRule {
	var policyRules []rbacv1.PolicyRule
	policyRules = append(policyRules, io.DiscoveryInputTypes.RbacPoliciesWatch()...)
	policyRules = append(policyRules, io.NetworkingOutputTypes.RbacPoliciesWrite()...)
	policyRules = append(policyRules, io.CertificateIssuerInputTypes.RbacPoliciesWatch()...)
	policyRules = append(policyRules, io.CertificateIssuerInputTypes.RbacPoliciesUpdateStatus()...)
	return policyRules
}()

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a Kubernetes cluster with Service Mesh Hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO(ilackarms): move verbose option to global flag at root level of meshctl
			if opts.verbose {
				logrus.SetLevel(logrus.DebugLevel)
			}

			if err := installCertAgent(ctx, opts); err != nil {
				return err
			}
			return registerCluster(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())
	return cmd
}

type options struct {
	register.RegistrationOptions
	certAgentInstallOptions
	verbose bool
}

type certAgentInstallOptions struct {
	chartPath   string
	chartValues string
}

func (register *options) addToFlags(set *pflag.FlagSet) {
	set.StringVar(&register.ClusterName, "cluster-name", "", "name of the cluster to register")
	set.StringVar(&register.KubeCfgPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	set.StringVar(&register.KubeContext, "master-context", "", "name of the kubeconfig context to use for the master cluster")
	set.StringVar(&register.RemoteKubeContext, "remote-context", "", "name of the kubeconfig context to use for the remote cluster")
	set.StringVar(&register.Namespace, "federation-namespace", defaults.DefaultPodNamespace, "namespace of the Service-Mesh-Hub control plane in which the secret for the registered cluster will be created")
	set.StringVar(&register.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created.")
	set.StringVar(&register.APIServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Service Mesh Hub is different than that specified in the local kubeconfig.")
	set.StringVar(&register.ClusterDomain, "cluster-domain", "", "The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
	set.StringVar(&register.chartPath, "cert-agent-chart-file", "", "Path to a local Helm chart for installing the Certificate Agent. If unset, this command will install the Certificate Agent from the publicly released Helm chart.")
	set.StringVar(&register.chartValues, "cert-agent-chart-values", "", "Path to a Helm values.yaml file for customizing the installation of the Certificate Agent. If unset, this command will install the Certificate Agent with default Helm values.")
	set.BoolVar(&register.verbose, "verbose", true, "enable/disable verbose logging during installation of cert-agent")
}

func registerCluster(ctx context.Context, opts *options) error {
	logrus.Debugf("registering cluster with opts %+v", opts)

	opts.ClusterRoles = []*rbacv1.ClusterRole{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: opts.RemoteNamespace,
				Name:      "smh-remote-access",
			},
			Rules: smhRbacRequirements,
		},
	}

	if err := opts.RegistrationOptions.RegisterCluster(ctx); err != nil {
		return err
	}

	logrus.Infof("successfully registered cluster %v", opts.ClusterName)
	return nil
}

func installCertAgent(ctx context.Context, opts *options) error {
	return agent.Installer{
		HelmChartOverride: opts.chartPath,
		HelmValuesPath:    opts.chartValues,
		KubeConfig:        opts.RemoteKubeCfgPath,
		KubeContext:       opts.RemoteKubeContext,
		Namespace:         opts.RemoteNamespace,
		Verbose:           false,
	}.InstallCertAgent(
		ctx,
	)
}
