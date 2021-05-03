package tlssecret

import (
	"istio.io/istio/pkg/test/framework/components/cluster"
	v1 "k8s.io/api/core/v1"

	"istio.io/istio/pkg/test/framework/resource"
)

type Instance interface {
	resource.Resource

	Secret() (*v1.Secret, error)
}

// Config represents the configuration for frontend.
type Config struct {
	Namespace string
	Name      string
	CACrt     string
	TLSKey    string
	TLSCert   string
	Cluster   cluster.Cluster
}

// New returns a new instance of Certificate.
func New(ctx resource.Context, cfg *Config) (i Instance, err error) {
	return newKube(ctx, cfg)
}
