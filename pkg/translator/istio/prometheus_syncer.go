package istio

import (
	prometheusv1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/shared"
	"k8s.io/client-go/kubernetes"
)

func NewPrometheusSyncer(kube kubernetes.Interface, prometheusClient prometheusv1.ConfigClient) v1.TranslatorSyncer {
	return &shared.PrometheusSyncer{
		Kube:                 kube,
		PrometheusClient:     prometheusClient,
		DesiredScrapeConfigs: IstioScrapeConfigs,
		MeshType:             v1.MeshType_ISTIO,
	}
}
