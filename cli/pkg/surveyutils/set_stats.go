package surveyutils

import (
	"context"

	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"

	"github.com/solo-io/supergloo/cli/pkg/options"
)

func SurveySetStats(ctx context.Context, in *options.SetStats) error {
	mesh, err := SurveyMesh("select the mesh for which you wish to propagate metrics", ctx)
	if err != nil {
		return err
	}

	promCfgs, err := SurveyPrometheusConfigs(ctx)
	if err != nil {
		return err
	}

	in.TargetMesh = options.ResourceRefValue(mesh)
	in.PrometheusConfigMaps = options.ResourceRefsValue(promCfgs)
	return nil
}

func SurveyPrometheusConfigs(ctx context.Context) ([]core.ResourceRef, error) {
	// collect prom configmaps list
	promClient := clients.MustPrometheusConfigClient()
	promCfgs, err := promClient.List("", skclients.ListOpts{Ctx: ctx})
	if err != nil {
		return nil, err
	}

	var selected []core.ResourceRef
	for {
		cfg, err := surveyResources("prometheus configmaps", "add a prometheus configmap (choose <done> to finish): ", "<done>", promCfgs.AsResources())
		if err != nil {
			return nil, err
		}
		// the user chose <done>
		if cfg.Namespace == "" {
			return selected, nil
		}
		selected = append(selected, cfg)
	}
}
