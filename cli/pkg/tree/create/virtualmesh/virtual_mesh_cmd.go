package virtualmesh

import (
	"context"
	"io"
	"strconv"

	"github.com/solo-io/service-mesh-hub/cli/pkg/cliconstants"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	common_config "github.com/solo-io/service-mesh-hub/cli/pkg/common/config"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/interactive"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/resource_printing"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	"github.com/solo-io/service-mesh-hub/cli/pkg/tree/create/validate"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/pkg/security/certgen"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CreateVirtualMeshCmd *cobra.Command

func CreateVirtualMeshCommand(
	ctx context.Context,
	out io.Writer,
	opts *options.Options,
	kubeLoader common_config.KubeLoader,
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
	kubeLoader common_config.KubeLoader,
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
) (*zephyr_networking.VirtualMesh, error) {
	var err error
	var displayName string
	var selectedMeshes []*core_types.ResourceRef
	var certificateAuthority *networking_types.VirtualMeshSpec_CertificateAuthority
	if displayName, err = interactivePrompt.PromptValueWithValidator("Resource Name", "", validate.K8sName); err != nil {
		return nil, err
	}
	if selectedMeshes, err = selectVirtualMeshesInteractive(allMeshNames, interactivePrompt); err != nil {
		return nil, err
	}
	if certificateAuthority, err = selectCertificateAuthority(interactivePrompt); err != nil {
		return nil, err
	}
	vm := &zephyr_networking.VirtualMesh{
		TypeMeta: v1.TypeMeta{Kind: "VirtualMesh"}, // k8s resource printers will complain unless this is set
		ObjectMeta: v1.ObjectMeta{
			Name:      displayName,
			Namespace: env.GetWriteNamespace(),
		},
		Spec: networking_types.VirtualMeshSpec{
			DisplayName:          displayName,
			Meshes:               selectedMeshes,
			CertificateAuthority: certificateAuthority,
			Federation: &networking_types.VirtualMeshSpec_Federation{
				Mode: networking_types.VirtualMeshSpec_Federation_PERMISSIVE,
			},
			TrustModel: &networking_types.VirtualMeshSpec_Shared{
				Shared: &networking_types.VirtualMeshSpec_SharedTrust{},
			},
		},
	}
	return vm, err
}

func selectVirtualMeshesInteractive(
	meshNames []string,
	interactivePrompt interactive.InteractivePrompt,
) ([]*core_types.ResourceRef, error) {
	selections, err := interactivePrompt.SelectMultipleValues("Select the Meshes to include in the VirtualMesh", meshNames)
	if err != nil {
		return nil, err
	}
	var selectedMeshNames []*core_types.ResourceRef
	for _, selection := range selections {
		selectedMeshNames = append(selectedMeshNames, &core_types.ResourceRef{
			Name:      selection,
			Namespace: env.GetWriteNamespace(),
		})
	}
	return selectedMeshNames, nil
}

func selectCertificateAuthority(interactivePrompt interactive.InteractivePrompt) (*networking_types.VirtualMeshSpec_CertificateAuthority, error) {
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
		return &networking_types.VirtualMeshSpec_CertificateAuthority{
			Type: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin_{
				Builtin: &networking_types.VirtualMeshSpec_CertificateAuthority_Builtin{
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
		return &networking_types.VirtualMeshSpec_CertificateAuthority{
			Type: &networking_types.VirtualMeshSpec_CertificateAuthority_Provided_{
				Provided: &networking_types.VirtualMeshSpec_CertificateAuthority_Provided{
					Certificate: &core_types.ResourceRef{
						Name:      name,
						Namespace: namespace,
					},
				},
			},
		}, nil
	}
}

func getAllMeshNames(ctx context.Context, meshClient zephyr_discovery.MeshClient) ([]string, error) {
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
