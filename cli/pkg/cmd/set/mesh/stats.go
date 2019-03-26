package mesh

import (
	"fmt"

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

func setStatsCmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stats",
		Aliases: []string{"st"},
		Short:   `configure one or more prometheus instances to scrape a mesh for metrics.`,
		Long:    `Updates the target mesh to propagate metrics to (have them scraped by) one or more instances of Prometheus.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if opts.Interactive {
				if err := surveyutils.SurveySetStats(opts.Ctx, &opts.SetStats); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(c *cobra.Command, args []string) error {
			if err := setStats(opts); err != nil {
				return err
			}
			return nil
		},
	}

	flagutils.AddSetStatsFlags(cmd.PersistentFlags(), &opts.SetStats)

	return cmd
}

func setStats(opts *options.Options) error {
	meshRef := opts.SetStats.TargetMesh
	if meshRef.Name == "" || meshRef.Namespace == "" {
		return errors.Errorf("must provide --target-mesh: %v", meshRef)
	}
	promConfigs := []core.ResourceRef(opts.SetStats.PrometheusConfigMaps)

	mesh, err := clients.MustMeshClient().Read(meshRef.Namespace, meshRef.Name, skclients.ReadOpts{Ctx: opts.Ctx})
	if err != nil {
		return err
	}

	// validate prom configs
	for _, cfg := range promConfigs {
		if _, err := clients.MustPrometheusConfigClient().Read(cfg.Namespace, cfg.Name, skclients.ReadOpts{Ctx: opts.Ctx}); err != nil {
			return errors.Wrapf(err, "failed to find prometheus configmap")
		}
	}

	if mesh.MonitoringConfig == nil {
		mesh.MonitoringConfig = &v1.MonitoringConfig{}
	}
	mesh.MonitoringConfig.PrometheusConfigmaps = promConfigs

	mesh, err = clients.MustMeshClient().Write(mesh, skclients.WriteOpts{Ctx: opts.Ctx, OverwriteExisting: true})
	if err != nil {
		return err
	}
	fmt.Printf("configured mesh %v to propagate metrics to %v prometheus instances\n", mesh.Metadata.Ref(), len(promConfigs))

	helpers.PrintMeshes(v1.MeshList{mesh}, opts.OutputType)

	return nil
}
