package upstream

import (
	"fmt"
	"path/filepath"

	"github.com/pkg/errors"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	"github.com/spf13/cobra"
)

const (
	certChain = "cert-chain.pem"
	key       = "key.pem"
	rootCa    = "root-cert.pem"
)

func setUpstreamTlsCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mtls",
		Short: "edit mtls settings for a Gloo upstream",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				us, err := surveyutils.SurveyUpstream(opts.Ctx)
				if err != nil {
					return err
				}
				opts.Metadata.Namespace = us.Namespace
				opts.Metadata.Name = us.Name
				mesh, err := surveyutils.SurveyMesh("select the mesh which you would like to connect to", opts.Ctx)
				if err != nil {
					return err
				}
				opts.EditUpstream.MtlsMeshMetadata = mesh
			}
			if err := validateSetUpstreamCmd(opts); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := editUpstream(opts); err != nil {
				return err
			}
			fmt.Printf("successfully updated upstream %s.%s", opts.Metadata.Namespace, opts.Metadata.Name)
			return nil
		},
	}

	flagutils.AddUpstreamTlsFlags(cmd.PersistentFlags(), &opts.EditUpstream)

	return cmd
}

func editUpstream(opts *options.Options) error {
	usClient := clients.MustUpstreamClient()
	usRef := opts.Metadata
	us, err := usClient.Read(usRef.Namespace, usRef.Name, skclients.ReadOpts{})
	if err != nil {
		return errors.Wrapf(err, "unable to find upstream %s.%s", usRef.Namespace, usRef.Name)
	}
	meshRef := opts.EditUpstream.MtlsMeshMetadata
	folderRoot := "/etc/certs"
	meshFolderPath := fmt.Sprintf("/%s/%s", meshRef.Namespace, meshRef.Name)
	sslConfig := &v1.UpstreamSslConfig{
		SslSecrets: &v1.UpstreamSslConfig_SslFiles{
			SslFiles: &v1.SSLFiles{
				TlsCert: filepath.Join(folderRoot, meshFolderPath, certChain),
				TlsKey:  filepath.Join(folderRoot, meshFolderPath, key),
				RootCa:  filepath.Join(folderRoot, meshFolderPath, rootCa),
			},
		},
	}
	us.UpstreamSpec.SslConfig = sslConfig
	_, err = usClient.Write(us, skclients.WriteOpts{
		OverwriteExisting: true,
	})
	if err != nil {
		return errors.Wrapf(err, "unable to save upstream %s.%s after editing ssl config", usRef.Namespace, usRef.Name)
	}
	return nil
}

func validateSetUpstreamCmd(opts *options.Options) error {
	if opts.EditUpstream.MtlsMeshMetadata.Namespace == "" || opts.EditUpstream.MtlsMeshMetadata.Name == "" {
		return fmt.Errorf("mesh resource name and namespace must be specified")
	}
	if opts.Metadata.Namespace == "" || opts.Metadata.Name == "" {
		return fmt.Errorf("upstream name and namespace must be specified")
	}

	// Check validity of mesh resource
	if !opts.Interactive {
		meshClient := clients.MustMeshClient()
		meshRef := opts.EditUpstream.MtlsMeshMetadata
		_, err := meshClient.Read(meshRef.Namespace, meshRef.Name, skclients.ReadOpts{Ctx: opts.Ctx})
		if err != nil {
			return errors.Wrapf(err, "unable to find mesh %s.%s", meshRef.Namespace, meshRef.Name)
		}
	}
	return nil
}
