package aws_utils

import (
	k8s_core_types "k8s.io/api/core/v1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type AppMeshPod struct {
	AwsAccountID    string
	Region          string
	AppMeshName     string
	VirtualNodeName string
}

type AppMeshScanner interface {
	ScanPodForAppMesh(
		pod *k8s_core_types.Pod,
		configMap *k8s_core_types.ConfigMap,
	) (*AppMeshPod, error)
}
type ArnParser interface {
	ParseAccountID(arn string) (string, error)
	ParseRegion(arn string) (string, error)
}
