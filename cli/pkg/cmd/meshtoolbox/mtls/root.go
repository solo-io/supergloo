package mtls

import (
	"fmt"

	"github.com/solo-io/supergloo/cli/pkg/cliconstants"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	"github.com/solo-io/supergloo/cli/pkg/nsutil"
	superglooV1 "github.com/solo-io/supergloo/pkg2/api/v1"
	"github.com/spf13/cobra"
)

// strings that users will pass to trigger commands
const (
	ENABLE_MTLS  = "enable"
	DISABLE_MTLS = "disable"
	TOGGLE_MTLS  = "toggle"
)

var validRootArgs = []string{ENABLE_MTLS, DISABLE_MTLS, TOGGLE_MTLS} // for bash completion

func Root(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:       cliconstants.MTLS,
		Short:     `set mTLS status`,
		Long:      `set mTLS status`,
		ValidArgs: validRootArgs,
		Args:      rootArgValidation,
		RunE: func(c *cobra.Command, args []string) error {
			// this function does nothing but it triggers validation
			return nil
		},
	}
	cmd.AddCommand(
		Enable(opts),
		Disable(opts),
		Toggle(opts),
	)
	return cmd
}

func rootArgValidation(c *cobra.Command, args []string) error {
	exactArgs := cobra.ExactArgs(1)
	if err := exactArgs(c, args); err != nil {
		return err
	}
	subCommandName := args[0]
	if !common.Contains(validRootArgs, subCommandName) {
		return fmt.Errorf("%v is not a valid argument", subCommandName)
	}
	return nil
}

func Enable(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   ENABLE_MTLS,
		Short: `enable mTLS`,
		Long:  `enable mTLS`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := enableMtls(opts); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func Disable(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   DISABLE_MTLS,
		Short: `disable mTLS`,
		Long:  `disable mTLS`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := disableMtls(opts); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func Toggle(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   TOGGLE_MTLS,
		Short: `toggle mTLS`,
		Long:  `toggle mTLS`,
		RunE: func(c *cobra.Command, args []string) error {
			if err := toggleMtls(opts); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}

func enableMtls(opts *options.Options) error {

	if _, err := updateMtls(ENABLE_MTLS, opts); err != nil {
		return err
	}
	fmt.Printf("Enabled mTLS on mesh %v", opts.MeshTool.Mesh.Name)

	return nil
}

func disableMtls(opts *options.Options) error {
	if _, err := updateMtls(DISABLE_MTLS, opts); err != nil {
		return err
	}
	fmt.Printf("Disabled mTLS on mesh %v", opts.MeshTool.Mesh.Name)
	return nil
}

func toggleMtls(opts *options.Options) error {
	mesh, err := updateMtls(TOGGLE_MTLS, opts)
	if err != nil {
		return err
	}
	status := "disabled"
	if mesh.Encryption.TlsEnabled {
		status = "enabled"
	}
	fmt.Printf("Toggled (%v) mTLS on mesh %v", status, opts.MeshTool.Mesh.Name)
	return nil
}

// Ensure that all the needed user-specified values have been provided
func ensureFlags(operation string, opts *options.Options) error {

	// all operations require a target mesh spec
	meshRef := &(opts.MeshTool).Mesh
	if err := nsutil.EnsureMesh(meshRef, opts); err != nil {
		return err
	}

	return nil
}

func updateMtls(operation string, opts *options.Options) (*superglooV1.Install, error) {
	// 1. validate/aquire arguments
	if err := ensureFlags(operation, opts); err != nil {
		return nil, err
	}

	// 2. read the existing install config, since mtls toggling happens in the install syncer (via helm redeploying)
	installClient, err := common.GetInstallClient()
	if err != nil {
		return nil, err
	}
	meshRef := &(opts.MeshTool).Mesh
	installCrd, err := (*installClient).Read(meshRef.Namespace, meshRef.Name, clients.ReadOpts{})
	if err != nil {
		return nil, err
	}

	// 3. mutate the mesh structure
	switch operation {
	case ENABLE_MTLS:
		if installCrd.Encryption == nil {
			installCrd.Encryption = &superglooV1.Encryption{
				TlsEnabled: true,
			}
		} else {
			installCrd.Encryption.TlsEnabled = true

		}
	case DISABLE_MTLS:
		if installCrd.Encryption == nil {
			installCrd.Encryption = &superglooV1.Encryption{
				TlsEnabled: false,
			}
		} else {
			installCrd.Encryption.TlsEnabled = false

		}
	case TOGGLE_MTLS:
		// if encryption has not been specified, "toggle" will enable it
		if installCrd.Encryption == nil {
			installCrd.Encryption = &superglooV1.Encryption{
				TlsEnabled: true,
			}
		} else {
			installCrd.Encryption.TlsEnabled = !installCrd.Encryption.TlsEnabled

		}
	default:
		panic(fmt.Errorf("Operation %v not recognized", operation))
	}

	// 4. write the changes
	writtenInstallCrd, err := (*installClient).Write(installCrd, clients.WriteOpts{OverwriteExisting: true})
	if err != nil {
		return nil, err
	}
	return writtenInstallCrd, nil
}
