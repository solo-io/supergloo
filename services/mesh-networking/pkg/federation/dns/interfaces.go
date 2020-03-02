package dns

import "context"

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type IpAssigner interface {
	// generate a unique IP on the given cluster
	AssignIPOnCluster(ctx context.Context, clusterName string) (clusterUniqueIp string, err error)

	// this method may block if the storage medium has not yet appeared
	UnAssignIPOnCluster(ctx context.Context, clusterName, ip string) error
}
