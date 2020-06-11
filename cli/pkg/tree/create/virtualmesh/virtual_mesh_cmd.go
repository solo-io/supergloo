package virtualmesh

import (
	"context"
	"io"
	"strconv"

	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/interactive"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/validate"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/common/csr/certgen"
	"github.com/solo-io/service-mesh-hub/pkg/common/kube/kubeconfig"
	"github.com/spf13/cobra"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateVirtualMeshCmd *cobra.Command

func CreateVirtualMeshCommand(
	ctx context.Context,
	out io.Writer,
	opts *options.Options,
	kubeLoader kubeconfig.KubeLoader,
	kubeClientsFactory common.KubeClientsFactory,
	interactivePrompt interactive.InteractivePrompt,
	printers common.Printers,
) CreateVirtualMeshCmd {
	cmd := cliconstants.CreateVirtualMeshCommand
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return createVirtualMesh(ctx, out, kubeLoader, kubeClientsFactory, opts, interactivePrompt, printers.ResourcePrinter)
	}
	return &cmd
}

func createVirtualMesh(
	ctx context.Context,
	out io.Writer,
	kubeLoader kubeconfig.KubeLoader,
	kubeClientsFactory common.KubeClientsFactory,
	opts *options.Options,
	interactivePrompt interactive.InteractivePrompt,
	resourcePrinter resource_printing.ResourcePrinter,
) error {
	masterCfg, err := kubeLoader.GetRestConfigForContext(opts.Root.KubeConfig, opts.Root.KubeContext)
	if err != nil {
		return err
	}
	masterKubeClients, err := kubeClientsFactory(masterCfg, opts.Root.WriteNamespace)
	if err != nil {
		return err
	}
	allMeshNames, err := getAllMeshNames(ctx, masterKubeClients.MeshClient)
	if err != nil {
		return err
	}
	vm, err := populateVirtualMeshInteractive(allMeshNames, interactivePrompt)
	if err != nil {
		return err
	}
	if !opts.Create.DryRun {
		return masterKubeClients.VirtualMeshClient.CreateVirtualMesh(ctx, vm)
	} else {
		return resourcePrinter.Print(out, vm, resource_printing.OutputFormat(opts.Create.OutputFormat))
	}
}

func populateVirtualMeshInteractive(
	allMeshNames []string,
	interactivePrompt interactive.InteractivePrompt,
) (*smh_networking.VirtualMesh, error) {
	var err error
	var displayName string
	var selectedMeshes []*smh_core_types.ResourceRef
	var certificateAuthority *smh_networking_types.VirtualMeshSpec_CertificateAuthority
	if displayName, err = interactivePrompt.PromptValueWithValidator("Resource Name", "", validate.K8sName); err != nil {
		return nil, err
	}
	if selectedMeshes, err = selectVirtualMeshesInteractive(allMeshNames, interactivePrompt); err != nil {
		return nil, err
	}
	if certificateAuthority, err = selectCertificateAuthority(interactivePrompt); err != nil {
		return nil, err
	}
	vm := &smh_networking.VirtualMesh{
		TypeMeta: k8s_meta_types.TypeMeta{Kind: "VirtualMesh"}, // k8s resource printers will complain unless this is set
		ObjectMeta: k8s_meta_types.ObjectMeta{
			Name:      displayName,
			Namespace: container_runtime.GetWriteNamespace(),
		},
		Spec: smh_networking_types.VirtualMeshSpec{
			DisplayName:          displayName,
			Meshes:               selectedMeshes,
			CertificateAuthority: certificateAuthority,
			Federation: &smh_networking_types.VirtualMeshSpec_Federation{
				Mode: smh_networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
			},
			TrustModel: &smh_networking_types.VirtualMeshSpec_Shared{
				Shared: &smh_networking_types.VirtualMeshSpec_SharedTrust{},
			},
		},
	}
	return vm, err
}

func selectVirtualMeshesInteractive(
	meshNames []string,
	interactivePrompt interactive.InteractivePrompt,
) ([]*smh_core_types.ResourceRef, error) {
	selections, err := interactivePrompt.SelectMultipleValues("Select the Meshes to include in the VirtualMesh", meshNames)
	if err != nil {
		return nil, err
	}
	var selectedMeshNames []*smh_core_types.ResourceRef
	for _, selection := range selections {
		selectedMeshNames = append(selectedMeshNames, &smh_core_types.ResourceRef{
			Name:      selection,
			Namespace: container_runtime.GetWriteNamespace(),
		})
	}
	return selectedMeshNames, nil
}

func selectCertificateAuthority(interactivePrompt interactive.InteractivePrompt) (*smh_networking_types.VirtualMeshSpec_CertificateAuthority, error) {
	builtin := "builtin"
	provided := "provided (user-supplied)"
	value, err := interactivePrompt.SelectValue("Certificate authority", []string{builtin, provided})
	if err != nil {
		return nil, err
	}
	if value == builtin {
		var err error
		var orgName, ttlString, rsaKeySizeString string
		var ttl, rsaKeySize int
		if ttlString, err = interactivePrompt.PromptValueWithValidator(
			"Root certificate TTL in days",
			strconv.Itoa(certgen.DefaultRootCertTTLDays),
			validate.PositiveInteger,
		); err != nil {
			return nil, err
		}
		if ttl, err = strconv.Atoi(ttlString); err != nil {
			return nil, err
		}
		if rsaKeySizeString, err = interactivePrompt.PromptValueWithValidator(
			"Root certificate RSA key size in bytes",
			strconv.Itoa(certgen.DefaultRootCertRsaKeySize),
			validate.PositiveInteger); err != nil {
			return nil, err
		}
		if rsaKeySize, err = strconv.Atoi(rsaKeySizeString); err != nil {
			return nil, err
		}
		if orgName, err = interactivePrompt.PromptRequiredValue("Root certificate organization name"); err != nil {
			return nil, err
		}
		return &smh_networking_types.VirtualMeshSpec_CertificateAuthority{
			Type: &smh_networking_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
				Builtin: &smh_networking_types.VirtualMeshSpec_CertificateAuthority_Builtin{
					TtlDays:         uint32(ttl),
					RsaKeySizeBytes: uint32(rsaKeySize),
					OrgName:         orgName,
				},
			},
		}, nil
	} else {
		var err error
		var name, namespace string
		if name, err = interactivePrompt.PromptRequiredValue("Root certificate k8s Secret name"); err != nil {
			return nil, err
		}
		if name, err = interactivePrompt.PromptRequiredValue("Root certificate k8s Secret namespace"); err != nil {
			return nil, err
		}
		return &smh_networking_types.VirtualMeshSpec_CertificateAuthority{
			Type: &smh_networking_types.VirtualMeshSpec_CertificateAuthority_Provided_{
				Provided: &smh_networking_types.VirtualMeshSpec_CertificateAuthority_Provided{
					Certificate: &smh_core_types.ResourceRef{
						Name:      name,
						Namespace: namespace,
					},
				},
			},
		}, nil
	}
}

func getAllMeshNames(ctx context.Context, meshClient smh_discovery.MeshClient) ([]string, error) {
	meshList, err := meshClient.ListMesh(ctx)
	if err != nil {
		return nil, err
	}
	var meshes []string
	for _, mesh := range meshList.Items {
		mesh := mesh
		meshes = append(meshes, mesh.GetName())
	}
	return meshes, nil
}
