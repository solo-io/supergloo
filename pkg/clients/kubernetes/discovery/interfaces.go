package kubernetes_discovery

import "k8s.io/apimachinery/pkg/version"

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type ServerVersionClient interface {
	Get() (*version.Info, error)
}
