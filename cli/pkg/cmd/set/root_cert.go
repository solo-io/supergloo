package set

import (
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/spf13/cobra"
)

func setRootCertCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rootcert",
		Aliases: []string{"rc"},
		Short:   `set the root certificate used to provision client and server certificates for a mesh`,
		Long: `Updates the target mesh to use the provided root certificate. Root certificate must be stored 
as a TLS secret created with ` + "`" + `supergloo create secret tls` + "`" + `. 
used to provision client and server certificates for a mesh`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveySetRootCert(opts.Ctx, &opts.SetRootCert); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			if err := setRootCert(opts); err != nil {
				return err
			}
			return nil
		},
	}

	flagutils.AddSetRootCertFlags(cmd.PersistentFlags(), &opts.SetRootCert)

	return cmd
}

func setRootCert(opts *options.Options) error {
	meshRef := opts.SetRootCert.TargetMesh
	if meshRef.Name == "" || meshRef.Namespace == "" {
		return errors.Errorf("must provide --target-mesh")
	}
	secretRef := core.ResourceRef(opts.SetRootCert.TlsSecret)

	mesh, err := clients.MustMeshClient().Read(meshRef.Namespace, meshRef.Name, skclients.ReadOpts{Ctx: opts.Ctx})
	if err != nil {
		return err
	}

	var rootCert *core.ResourceRef
	if secretRef.Namespace == "" && secretRef.Name == "" {
		contextutils.LoggerFrom(opts.Ctx).Warnf("no --tls-secret set, removing root cert if set")
	} else {
		_, err := clients.MustTlsSecretClient().Read(secretRef.Namespace, secretRef.Name, skclients.ReadOpts{Ctx: opts.Ctx})
		if err != nil {
			return err
		}

		rootCert = &secretRef
	}

	if mesh.MtlsConfig == nil {
		mesh.MtlsConfig = &v1.MtlsConfig{}
	}

	if rootCert != nil {
		mesh.MtlsConfig.MtlsEnabled = true
	}

	mesh.MtlsConfig.RootCertificate = rootCert

	mesh, err = clients.MustMeshClient().Write(mesh, skclients.WriteOpts{Ctx: opts.Ctx, OverwriteExisting: true})
	if err != nil {
		return err
	}

	helpers.PrintMeshes(v1.MeshList{mesh}, opts.OutputType)

	return nil
}
