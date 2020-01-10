package inputs

import (
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/options/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

func BookInfoUpstreams(namespace string) v1.UpstreamList {
	return v1.UpstreamList{
		&v1.Upstream{
			UpstreamType: &v1.Upstream_Kube{
				Kube: &kubernetes.UpstreamSpec{
					ServiceName:      "details",
					ServiceNamespace: namespace,
					ServicePort:      0x00002378,
					Selector: map[string]string{
						"app": "details",
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-details-9080",
				Namespace: namespace,
				Labels: map[string]string{
					"app":           "details",
					"discovered_by": "kubernetesplugin",
				},
			},
		},
		&v1.Upstream{
			UpstreamType: &v1.Upstream_Kube{
				Kube: &kubernetes.UpstreamSpec{
					ServiceName:      "details",
					ServiceNamespace: namespace,
					ServicePort:      0x00002378,
					Selector: map[string]string{
						"version": "v1",
						"app":     "details",
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-details-v1-9080",
				Namespace: namespace,
				Labels: map[string]string{
					"app":           "details",
					"discovered_by": "kubernetesplugin",
				},
			},
		},
		&v1.Upstream{
			UpstreamType: &v1.Upstream_Kube{
				Kube: &kubernetes.UpstreamSpec{
					ServiceName:      "productpage",
					ServiceNamespace: namespace,
					ServicePort:      0x00002378,
					Selector: map[string]string{
						"app": "productpage",
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-productpage-9080",
				Namespace: namespace,
				Labels: map[string]string{
					"app":           "productpage",
					"discovered_by": "kubernetesplugin",
				},
			},
		},
		&v1.Upstream{
			UpstreamType: &v1.Upstream_Kube{
				Kube: &kubernetes.UpstreamSpec{
					ServiceName:      "productpage",
					ServiceNamespace: namespace,
					ServicePort:      0x00002378,
					Selector: map[string]string{
						"app":     "productpage",
						"version": "v1",
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-productpage-v1-9080",
				Namespace: namespace,
				Labels: map[string]string{
					"app":           "productpage",
					"discovered_by": "kubernetesplugin",
				},
			},
		},
		&v1.Upstream{
			UpstreamType: &v1.Upstream_Kube{
				Kube: &kubernetes.UpstreamSpec{
					ServiceName:      "ratings",
					ServiceNamespace: namespace,
					ServicePort:      0x00002378,
					Selector: map[string]string{
						"app": "ratings",
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-ratings-9080",
				Namespace: namespace,
				Labels: map[string]string{
					"app":           "ratings",
					"discovered_by": "kubernetesplugin",
				},
			},
		},
		&v1.Upstream{
			UpstreamType: &v1.Upstream_Kube{
				Kube: &kubernetes.UpstreamSpec{
					ServiceName:      "ratings",
					ServiceNamespace: namespace,
					ServicePort:      0x00002378,
					Selector: map[string]string{
						"app":     "ratings",
						"version": "v1",
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-ratings-v1-9080",
				Namespace: namespace,
				Labels: map[string]string{
					"app":           "ratings",
					"discovered_by": "kubernetesplugin",
				},
			},
		},
		&v1.Upstream{
			UpstreamType: &v1.Upstream_Kube{
				Kube: &kubernetes.UpstreamSpec{
					ServiceName:      "reviews",
					ServiceNamespace: namespace,
					ServicePort:      0x00002378,
					Selector: map[string]string{
						"app": "reviews",
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-reviews-9080",
				Namespace: namespace,
				Labels: map[string]string{
					"app":           "reviews",
					"discovered_by": "kubernetesplugin",
				},
			},
		},
		&v1.Upstream{
			UpstreamType: &v1.Upstream_Kube{
				Kube: &kubernetes.UpstreamSpec{
					ServiceName:      "reviews",
					ServiceNamespace: namespace,
					ServicePort:      0x00002378,
					Selector: map[string]string{
						"version": "v1",
						"app":     "reviews",
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-reviews-v1-9080",
				Namespace: namespace,
				Labels: map[string]string{
					"app":           "reviews",
					"discovered_by": "kubernetesplugin",
				},
			},
		},
		&v1.Upstream{
			UpstreamType: &v1.Upstream_Kube{
				Kube: &kubernetes.UpstreamSpec{
					ServiceName:      "reviews",
					ServiceNamespace: namespace,
					ServicePort:      0x00002378,
					Selector: map[string]string{
						"app":     "reviews",
						"version": "v2",
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-reviews-v2-9080",
				Namespace: namespace,
				Labels: map[string]string{
					"discovered_by": "kubernetesplugin",
					"app":           "reviews",
				},
			},
		},
		&v1.Upstream{
			UpstreamType: &v1.Upstream_Kube{
				Kube: &kubernetes.UpstreamSpec{
					ServiceName:      "reviews",
					ServiceNamespace: namespace,
					ServicePort:      0x00002378,
					Selector: map[string]string{
						"version": "v3",
						"app":     "reviews",
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-reviews-v3-9080",
				Namespace: namespace,
				Labels: map[string]string{
					"discovered_by": "kubernetesplugin",
					"app":           "reviews",
				},
			},
		},
	}
}
