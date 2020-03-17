package dns

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type IpAssigner interface {
	// generate a unique IP on the given cluster
	AssignIPOnCluster(ctx context.Context, clusterName string) (clusterUniqueIp string, err error)

	// this method may block if the storage medium has not yet appeared
	UnAssignIPOnCluster(ctx context.Context, clusterName, ip string) error
}

type ExternalAccessPoint struct {
	Address string
	Port    uint32
}

type ExternalAccessPointGetter interface {
	GetExternalAccessPointForService(
		ctx context.Context,
		svc *corev1.Service,
		portName, clusterName string,
	) (eap ExternalAccessPoint, err error)
}
