package meshservice

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/olekukonko/tablewriter"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/common/schemes"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/describe/printing"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:     "meshservice",
		Short:   "Description of managed mesh services",
		Aliases: []string{"meshservices"},
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := buildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			description, err := describeMeshServices(ctx, c)
			if err != nil {
				return err
			}
			fmt.Println(description)
			return nil
		},
	}
	opts.addToFlags(cmd.Flags())

	return cmd
}

func describeMeshServices(ctx context.Context, c client.Client) (string, error) {
	meshServiceClient := discoveryv1alpha2.NewMeshServiceClient(c)
	meshServiceList, err := meshServiceClient.ListMeshService(ctx)
	if err != nil {
		return "", err
	}
	var meshServiceDescriptions []meshServiceDescription
	for _, meshService := range meshServiceList.Items {
		meshService := meshService // pike
		meshServiceDescriptions = append(meshServiceDescriptions, describeMeshService(&meshService))
	}

	buf := new(bytes.Buffer)
	table := tablewriter.NewWriter(buf)
	table.SetHeader([]string{"Metadata", "Traffic_Policies", "Access_Policies"})
	table.SetRowLine(true)
	table.SetAutoWrapText(false)

	for _, description := range meshServiceDescriptions {
		table.Append([]string{
			description.Metadata.string(),
			printing.FormattedObjectRefs(description.TrafficPolicies),
			printing.FormattedObjectRefs(description.AccessPolicies),
		})
	}
	table.Render()

	return buf.String(), nil
}

func (m meshServiceMetadata) string() string {
	var s strings.Builder
	s.WriteString(printing.FormattedField("Name", m.Name))
	s.WriteString(printing.FormattedField("Namespace", m.Namespace))
	s.WriteString(printing.FormattedField("Cluster", m.Cluster))
	s.WriteString(printing.FormattedField("Type", m.Type))
	return s.String()
}

type meshServiceDescription struct {
	Metadata        *meshServiceMetadata
	TrafficPolicies []*v1.ObjectRef
	AccessPolicies  []*v1.ObjectRef
}

type meshServiceMetadata struct {
	Type      string
	Name      string
	Namespace string
	Cluster   string
}

func describeMeshService(meshService *discoveryv1alpha2.MeshService) meshServiceDescription {
	meshMeta := getMeshServiceMetadata(meshService)

	var trafficPolicies []*v1.ObjectRef
	for _, fs := range meshService.Status.AppliedTrafficPolicies {
		trafficPolicies = append(trafficPolicies, fs.Ref)
	}

	var accessPolicies []*v1.ObjectRef
	for _, vm := range meshService.Status.AppliedAccessPolicies {
		accessPolicies = append(accessPolicies, vm.Ref)
	}

	return meshServiceDescription{
		Metadata:        &meshMeta,
		TrafficPolicies: trafficPolicies,
		AccessPolicies:  accessPolicies,
	}
}

func getMeshServiceMetadata(meshService *discoveryv1alpha2.MeshService) meshServiceMetadata {
	switch meshService.Spec.GetType().(type) {
	case *discoveryv1alpha2.MeshServiceSpec_KubeService_:
		kubeServiceRef := meshService.Spec.GetKubeService().Ref
		return meshServiceMetadata{
			Type:      "kubernetes service",
			Name:      kubeServiceRef.Name,
			Namespace: kubeServiceRef.Namespace,
			Cluster:   kubeServiceRef.ClusterName,
		}
	}
	return meshServiceMetadata{}
}

type options struct {
	kubeconfig  string
	kubecontext string
	namespace   string
}

func (o *options) addToFlags(set *pflag.FlagSet) {
	set.StringVar(&o.kubeconfig, "kubeconfig", "", "path to the kubeconfig from which the registered cluster will be accessed")
	set.StringVar(&o.kubecontext, "kubecontext", "", "name of the kubeconfig context to use for the management cluster")
	set.StringVar(&o.namespace, "namespace", defaults.DefaultPodNamespace, "namespace that Service MeshService Hub is installed in")
}

// TODO(harveyxia) move this into a shared CLI util
func buildClient(kubeconfig, kubecontext string) (client.Client, error) {
	if kubeconfig == "" {
		kubeconfig = clientcmd.RecommendedHomeFile
	}
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.ExplicitPath = kubeconfig
	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: kubecontext}

	cfg, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}

	scheme := scheme.Scheme
	if err := schemes.SchemeBuilder.AddToScheme(scheme); err != nil {
		return nil, err
	}

	client, err := client.New(cfg, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}
