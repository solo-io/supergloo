package istio

import (
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
)

// creates a virtual service rule for every host
// then a subset for every unique set of labels therein
func virtualServicesFromUpstreams(writeNamespace string, upstreams gloov1.UpstreamList) (v1alpha3.VirtualServiceList, error) {
	var virtualServices v1alpha3.VirtualServiceList

	labelsByHost, err := labelsByHost(upstreams)
	if err != nil {
		return nil, errors.Wrapf(err, "getting hostnames and labels for upstreams")
	}

	for host, labelSets := range labelsByHost {
		var subsets []*v1alpha3.Subset
		for _, labels := range labelSets {
			if len(labels) == 0 {
				continue
			}
			subsets = append(subsets, &v1alpha3.Subset{
				Name:   subsetName(labels),
				Labels: labels,
			})
		}
		virtualServices = append(virtualServices, &v1alpha3.VirtualService{
			Metadata: core.Metadata{
				Namespace: writeNamespace,
				Name:      host,
			},
			Hosts:    []string{host},
			Gateways: []string{"mesh"},
		})
	}

	return virtualServices.Sort(), nil
}
