package apply

import (
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/cli/pkg/options"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func ConvertSelector(in options.Selector) (*v1.PodSelector, error) {
	useLabels := len(in.SelectedLabels) > 0
	useUpstreams := len(in.SelectedUpstreams) > 0
	useNamespaces := len(in.SelectedNamespaces) > 0
	switch {
	case !(useLabels || useUpstreams || useNamespaces):
		return nil, nil
	case (useLabels && useUpstreams) || (useLabels && useNamespaces) || (useUpstreams && useNamespaces):
		return nil, errors.Errorf("you may only use one type of selector: upstreams, namespaces, or labels")
	case useLabels:
		return &v1.PodSelector{
			SelectorType: &v1.PodSelector_LabelSelector_{
				LabelSelector: &v1.PodSelector_LabelSelector{
					LabelsToMatch: in.SelectedLabels,
				},
			},
		}, nil
	case useUpstreams:
		return &v1.PodSelector{
			SelectorType: &v1.PodSelector_UpstreamSelector_{
				UpstreamSelector: &v1.PodSelector_UpstreamSelector{
					Upstreams: in.SelectedUpstreams,
				},
			},
		}, nil
	case useNamespaces:
		return &v1.PodSelector{
			SelectorType: &v1.PodSelector_NamespaceSelector_{
				NamespaceSelector: &v1.PodSelector_NamespaceSelector{
					Namespaces: in.SelectedNamespaces,
				},
			},
		}, nil
	}
	return nil, nil
}
