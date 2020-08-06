package mesh

import (
	"context"
	"fmt"

	"github.com/ghodss/yaml"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/solo-io/service-mesh-hub/pkg/common/schemes"
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
		Use:   "mesh",
		Short: "Description of managed meshes",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := buildClient(opts.kubeconfig, opts.kubecontext)
			if err != nil {
				return err
			}
			description, err := describeMeshes(ctx, c)
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

func describeMeshes(ctx context.Context, c client.Client) (string, error) {
	meshClient := discoveryv1alpha2.NewMeshClient(c)
	meshList, err := meshClient.ListMesh(ctx)
	if err != nil {
		return "", err
	}
	var meshDescriptions []meshDescription
	for _, mesh := range meshList.Items {
		mesh := mesh // pike
		meshDescriptions = append(meshDescriptions, describeMesh(&mesh))
	}

	d := description{Meshes: meshDescriptions}
	return d.toString()
}

type description struct {
	Meshes []meshDescription `json:"meshes,omitempty"`
}

func (m description) toString() (string, error) {
	bytes, err := yaml.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

type meshDescription struct {
	Mesh             *meshMetadata   `json:"metadata,omitempty"`
	VirtualMeshes    []*v1.ObjectRef `json:"virtualMeshes,omitempty"`
	FailoverServices []*v1.ObjectRef `json:"failoverServices,omitempty"`
}

type meshMetadata struct {
	Type         string   `json:"type,omitempty"`
	Name         string   `json:"name,omitempty"`
	Namespace    string   `json:"namespace,omitempty"`
	Region       string   `json:"region,omitempty"`
	AwsAccountId string   `json:"awsAccountId,omitempty"`
	Cluster      string   `json:"cluster,omitempty"`
	Clusters     []string `json:"clusters,omitempty"`
	Version      string   `json:"version,omitempty"`
}

func describeMesh(mesh *discoveryv1alpha2.Mesh) meshDescription {
	meshMeta := getMeshMetadata(mesh)

	var virtualMeshes []*v1.ObjectRef
	for _, vm := range mesh.Status.AppliedVirtualMeshes {
		virtualMeshes = append(virtualMeshes, vm.Ref)
	}

	var failoverServices []*v1.ObjectRef
	for _, fs := range mesh.Status.AppliedFailoverServices {
		failoverServices = append(failoverServices, fs.Ref)
	}

	return meshDescription{
		Mesh:             &meshMeta,
		VirtualMeshes:    virtualMeshes,
		FailoverServices: failoverServices,
	}
}

func getMeshMetadata(mesh *discoveryv1alpha2.Mesh) meshMetadata {
	var meshType string
	if mesh.Spec.GetAwsAppMesh() != nil {
		appmesh := mesh.Spec.GetAwsAppMesh()
		return meshMetadata{
			Type:         "appmesh",
			Name:         appmesh.Name,
			Region:       appmesh.Region,
			AwsAccountId: appmesh.AwsAccountId,
			Clusters:     appmesh.Clusters,
		}
	}
	var meshInstallation *discoveryv1alpha2.MeshSpec_MeshInstallation
	switch mesh.Spec.GetMeshType().(type) {
	case *discoveryv1alpha2.MeshSpec_Istio_:
		meshType = "istio"
		meshInstallation = mesh.Spec.GetIstio().Installation
	case *discoveryv1alpha2.MeshSpec_Linkerd:
		meshType = "linkerd"
		meshInstallation = mesh.Spec.GetLinkerd().Installation
	case *discoveryv1alpha2.MeshSpec_ConsulConnect:
		meshType = "consulconnect"
		meshInstallation = mesh.Spec.GetConsulConnect().Installation
	}
	return meshMetadata{
		Type:      meshType,
		Namespace: meshInstallation.Namespace,
		Cluster:   meshInstallation.Cluster,
		Version:   meshInstallation.Version,
	}
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
	if kubeconfig != "" {
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
