package main

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/common/container-runtime"
	mesh_discovery "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery"
)

func main() {
	ctx := container_runtime.CreateRootContext(context.Background(), "mesh-discovery")
	err := mesh_discovery.Start(ctx, mesh_discovery.Options{
		// TODO(ilackarms): populate these fields from Settings CR or CLI flags
		MetricsBindAddress: "",
		MasterNamespace:    "",
	})
	if err != nil {
		contextutils.LoggerFrom(ctx).Fatal(err)
	} else {
		contextutils.LoggerFrom(ctx).Info("exiting")
	}
}
