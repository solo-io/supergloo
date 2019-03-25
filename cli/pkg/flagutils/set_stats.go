package flagutils

import (
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/spf13/pflag"
)

func AddSetStatsFlags(set *pflag.FlagSet, in *options.SetStats) {
	set.Var(&in.TargetMesh, "target-mesh", "resource reference the mesh for which you wish to configure metrics. format must be <NAMESPACE>.<NAME>")
	set.Var(&in.PrometheusConfigMaps, "prometheus-configmap",
		"resource reference to a prometheus configmap (used to configure prometheus in kubernetes) "+
			"to which supergloo will ensure metrics are propagated. if empty, the any existing metric propagation will be disconnected. format must be <NAMESPACE>.<NAME>")
}
