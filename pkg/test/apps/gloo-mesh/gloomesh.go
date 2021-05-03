package gloo_mesh

import (
	"fmt"
	"github.com/solo-io/gloo-mesh/pkg/test/apps/context"
	"istio.io/istio/pkg/test/framework/resource"
	"istio.io/istio/pkg/test/util/retry"
	"time"
)

type Config struct {
	ClusterKubeConfigs                  map[string]string
	DeployControlPlaneToManagementPlane bool
}

const glooMeshVersion = "1.0.5"

func Deploy(deploymentCtx *context.DeploymentContext, cfg *Config, licenceKey string) resource.SetupFn {
	return func(ctx resource.Context) error {
		if deploymentCtx == nil {
			*deploymentCtx = context.DeploymentContext{}
		}
		deploymentCtx.Meshes = []context.GlooMeshInstance{}
		var i context.GlooMeshInstance
		var err error
		// install management plane
		// we need the MP to always be installed and added to the instance list first because
		// istio uninstalls in reverse order meaning control planes unregister first before uninstalling MP
		mpcfg := InstanceConfig{
			managementPlane:               true,
			managementPlaneKubeConfigPath: cfg.ClusterKubeConfigs[ctx.Clusters()[0].Name()],
			version:                       glooMeshVersion,
			clusterDomain:                 "",
			cluster:                       ctx.Clusters()[0],
		}
		mpInstance, err := newInstance(ctx, mpcfg, licenceKey)
		if err != nil {
			return err
		}
		deploymentCtx.Meshes = append(deploymentCtx.Meshes, mpInstance)

		var relayAddress string
		if licenceKey != "" {
			if err := retry.UntilSuccess(func() error {
				relayAddress, err = mpInstance.GetRelayServerAddress()
				if err != nil {
					return err
				}
				return nil
			}, retry.Timeout(30*time.Second), retry.Delay(2*time.Second)); err != nil {
				return fmt.Errorf("failed to find relay server address %s", err.Error())
			}
		}

		// install control planes
		var index = 0
		for n, p := range cfg.ClusterKubeConfigs {
			if n == ctx.Clusters()[index].Name() && !cfg.DeployControlPlaneToManagementPlane {
				// skip deploying CP to MP
				index++
				continue
			}

			cpcfg := InstanceConfig{
				managementPlane:                   false,
				controlPlaneKubeConfigPath:        p,
				managementPlaneKubeConfigPath:     cfg.ClusterKubeConfigs[ctx.Clusters()[0].Name()],
				version:                           glooMeshVersion,
				cluster:                           ctx.Clusters()[index],
				managementPlaneRelayServerAddress: relayAddress,
				clusterDomain:                     "",
			}
			i, err = newInstance(ctx, cpcfg, licenceKey)
			if err != nil {
				return err
			}
			deploymentCtx.Meshes = append(deploymentCtx.Meshes, i)
			index++
		}
		return nil
	}
}
