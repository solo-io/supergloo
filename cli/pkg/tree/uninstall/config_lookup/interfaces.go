package config_lookup

import (
	"context"

	"github.com/solo-io/service-mesh-hub/cli/pkg/common/kube"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type KubeConfigLookup interface {
	// get various pieces of config corresponding to a registered kube cluster
	FromCluster(ctx context.Context, clusterName string) (config *kube.ConvertedConfigs, err error)
}
