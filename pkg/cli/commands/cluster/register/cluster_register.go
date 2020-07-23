package register

import (
	"context"

	"github.com/solo-io/skv2/pkg/multicluster/register"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var smhRbacRequirements = []rbacv1.PolicyRule{
	{
		Verbs: []string{"*"},
		APIGroups: []string{
			"networking.istio.io",
		},
		Resources: []string{
			"destinationrules",
			"envoyfilters",
			"virtualservices",
		},
	},
	{
		Verbs: []string{"*"},
		APIGroups: []string{
			"security.istio.io",
		},
		Resources: []string{
			"authorizationpolicies",
		},
	},
	{
		Verbs:     []string{"get", "list", "watch"},
		APIGroups: []string{"apps"},
		Resources: []string{
			"deployments",
			"daemonsets",
			"replicasets",
			"statefulsets",
		},
	},
	{
		Verbs:     []string{"get", "list", "watch"},
		APIGroups: []string{""},
		Resources: []string{
			"pods",
			"services",
			"configmaps",
		},
	},
}

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "register",
		Short: "Register a Kubernetes cluster with Service Mesh Hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			return registerCluster(ctx, opts)
		},
	}
	opts.AddRegisterFlags(cmd.Flags())
	return cmd
}

type options register.RegistrationOptions

func (register *options) AddRegisterFlags(set *pflag.FlagSet) {
	set.StringVar(&register.ClusterName, "cluster-name", "", "name of the cluster to register")
	set.StringVar(&register.KubeCfgPath, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	set.StringVar(&register.KubeContext, "context", "", "name of the kubeconfig context to use for registration")
	set.StringVar(&register.Namespace, "federation-namespace", defaults.DefaultPodNamespace, "namespace of the Service-Mesh-Hub control plane in which the secret for the registered cluster will be created")
	set.StringVar(&register.RemoteNamespace, "remote-namespace", defaults.DefaultPodNamespace, "namespace in the target cluster where a service account enabling remote access will be created. If the namespace does not exist it will be created.")
	set.StringVar(&register.APIServerAddress, "api-server-address", "", "Swap out the address of the remote cluster's k8s API server for the value of this flag. Set this flag when the address of the cluster domain used by the Service Mesh Hub is different than that specified in the local kubeconfig.")
	set.StringVar(&register.ClusterDomain, "cluster-domain", "", "The Cluster Domain used by the Kubernetes DNS Service in the registered cluster. Defaults to 'cluster.local'. Read more: https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/")
}

func registerCluster(ctx context.Context, opts *options) error {

	opts.ClusterRoles = []*rbacv1.ClusterRole{
		{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: opts.RemoteNamespace,
				Name:      "smh-remote-access",
			},
			Rules: smhRbacRequirements,
		},
	}

	return register.RegistrationOptions(*opts).RegisterCluster(ctx)
}
