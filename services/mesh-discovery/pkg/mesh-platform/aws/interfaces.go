package aws

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

type AppMeshParser interface {
	ScanPodForAppMesh(pod *k8s_core_types.Pod) (*AppMeshPod, error)
}

type ArnParser interface {
	ParseAccountID(arn string) (string, error)
}
