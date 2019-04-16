package config

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
)

type EventLoop interface {
	Run(namespaces []string, opts clients.WatchOpts) (<-chan error, error)
}
