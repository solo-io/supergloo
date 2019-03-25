package istio

import (
	"github.com/solo-io/supergloo/pkg/stats/common"

	"github.com/prometheus/prometheus/config"
	"github.com/solo-io/go-utils/errors"
	prometheusv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

// registration-level syncer

func NewIstioPrometheusSyncer(client prometheusv1.PrometheusConfigClient) v1.RegistrationSyncer {
	return common.NewPrometheusSyncer("istio", client, chooseMesh, getScrapeConfigs)
}

func chooseMesh(mesh *v1.Mesh) bool {
	return mesh.GetIstio() != nil
}

func getScrapeConfigs(mesh *v1.Mesh) ([]*config.ScrapeConfig, error) {
	istio := mesh.GetIstio()
	if istio == nil {
		return nil, errors.Errorf("internal error: mesh %v was expected to be type istio", mesh.Metadata.Ref())
	}
	return PrometheusScrapeConfigs(istio.InstallationNamespace)
}
