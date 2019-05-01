package register

import (
	"strings"

	"github.com/solo-io/supergloo/cli/pkg/helpers"

	"github.com/pkg/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/cmd/apply"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/registration/appmesh"
	"github.com/spf13/cobra"
)

func registerAwsAppMeshCommand(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "appmesh",
		Short: `Register an AWS App Mesh with SuperGloo`,
		Long: `Creates a SuperGloo Mesh object representing an AWS App Mesh. The object will contain the information required to 
connect to the App Mesh control plane in AWS and, optionally, to automatically inject new pods with the AWS App Mesh sidecar proxy`,

		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveyMetadata("mesh", &opts.Metadata); err != nil {
					return err
				}
				if err := surveyutils.SurveyAppmeshRegistration(opts.Ctx, opts); err != nil {
					return err
				}
			}
			return nil
		},

		RunE: func(c *cobra.Command, args []string) error {

			if err := validateFlags(opts); err != nil {
				return err
			}

			autoInjectionEnabled := opts.RegisterAppMesh.EnableAutoInjection == "true"

			mesh := &v1.Mesh{
				Metadata: core.Metadata{
					Name:      opts.Metadata.Name,
					Namespace: opts.Metadata.Namespace,
				},
				MeshType: &v1.Mesh_AwsAppMesh{
					AwsAppMesh: &v1.AwsAppMesh{
						AwsSecret: &core.ResourceRef{
							Namespace: opts.RegisterAppMesh.Secret.Namespace,
							Name:      opts.RegisterAppMesh.Secret.Name,
						},
						Region:           opts.RegisterAppMesh.Region,
						EnableAutoInject: autoInjectionEnabled,
					},
				},
			}

			if autoInjectionEnabled {

				selector, err := apply.ConvertSelector(opts.RegisterAppMesh.PodSelector)
				if err != nil {
					return errors.Wrapf(err, "failed to convert pod selector")
				}
				mesh.GetAwsAppMesh().InjectionSelector = selector

				mesh.GetAwsAppMesh().VirtualNodeLabel = opts.RegisterAppMesh.VirtualNodeLabel

				if opts.RegisterAppMesh.ConfigMap.Namespace != "" && opts.RegisterAppMesh.ConfigMap.Name != "" {
					mesh.GetAwsAppMesh().SidecarPatchConfigMap = &core.ResourceRef{
						Namespace: opts.RegisterAppMesh.ConfigMap.Namespace,
						Name:      opts.RegisterAppMesh.ConfigMap.Name,
					}
				}
			}

			_, err := clients.MustMeshClient().Write(mesh, skclients.WriteOpts{})
			if err != nil {
				return errors.Wrapf(err, "failed to create mesh %v", mesh.Metadata.String())
			}

			helpers.PrintMeshes(v1.MeshList{mesh}, opts.OutputType)

			return nil
		},
	}

	flagutils.RegisterAwsAppMeshFlags(cmd.PersistentFlags(), &opts.RegisterAppMesh)

	return cmd
}

func validateFlags(opts *options.Options) error {

	if opts.Metadata.Name == "" {
		return errors.Errorf("name cannot be empty, provide with --name flag")
	}

	if opts.RegisterAppMesh.Region == "" {
		return errors.Errorf("AWS region cannot be empty, provide with --region flag")
	}

	if !strings.Contains(appmesh.AppMeshAvailableRegions, opts.RegisterAppMesh.Region) {
		return errors.Errorf("invalid region. AWS App Mesh is currently available in the following regions: \n%s\n",
			strings.ReplaceAll(appmesh.AppMeshAvailableRegions, ",", "\n"))
	}

	if opts.RegisterAppMesh.Secret.Namespace == "" || opts.RegisterAppMesh.Secret.Name == "" {
		return errors.Errorf("you must provide a fully qualified secret name in the format NAMESPACE.NAME via the --secret flag")
	}

	if opts.RegisterAppMesh.EnableAutoInjection != "true" && opts.RegisterAppMesh.EnableAutoInjection != "false" {
		return errors.Errorf("invalid value for --auto-inject flag. Must be either true or false")
	}

	if opts.RegisterAppMesh.EnableAutoInjection == "false" {
		return nil
	}

	if len(opts.RegisterAppMesh.PodSelector.SelectedNamespaces) == 0 &&
		len(opts.RegisterAppMesh.PodSelector.SelectedLabels) == 0 {
		return errors.Errorf("you must provide a pod selector if auto-injection is enabled")
	}

	if opts.RegisterAppMesh.VirtualNodeLabel == "" {
		return errors.Errorf("you must provide a virtual node label if auto-injection is enabled")
	}

	return nil
}
