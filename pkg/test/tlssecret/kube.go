package tlssecret

import (
	testk8s "github.com/solo-io/gloo-mesh/pkg/test/kube"
	v1 "k8s.io/api/core/v1"
	kubeApiMeta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"istio.io/istio/pkg/test/framework/resource"
)

var (
	_ Instance = &kubeComponent{}
)

type kubeComponent struct {
	id       resource.ID
	settings *Config
	secret   *v1.Secret
}

func newKube(ctx resource.Context, cfg *Config) (Instance, error) {
	c := &kubeComponent{}
	c.id = ctx.TrackResource(c)
	c.settings = cfg

	secret := &v1.Secret{
		ObjectMeta: kubeApiMeta.ObjectMeta{
			Name:      cfg.Name,
			Namespace: cfg.Namespace,
		},
		StringData: map[string]string{
			v1.ServiceAccountRootCAKey: cfg.CACrt,
			v1.TLSCertKey:              cfg.TLSCert,
			v1.TLSPrivateKeyKey:        cfg.TLSKey,
		},
		Type: v1.SecretTypeTLS,
	}

	s, err := testk8s.CreateSecret(cfg.Cluster, secret)
	if err != nil {
		return nil, err
	}
	c.secret = s
	return c, nil
}

func (c *kubeComponent) Secret() (*v1.Secret, error) {
	s, err := testk8s.GetSecret(c.settings.Cluster, c.settings.Name, c.settings.Namespace)

	return s, err
}

func (c *kubeComponent) ID() resource.ID {
	return c.id
}
