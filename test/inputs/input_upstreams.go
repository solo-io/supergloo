package inputs

import (
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/kubernetes"
	"github.com/solo-io/solo-kit/api/external/kubernetes/service"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	kubev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func BookInfoUpstreams(namespace string) v1.UpstreamList {
	return v1.UpstreamList{
		&v1.Upstream{
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "details",
						ServiceNamespace: namespace,
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "details",
						},
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
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
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
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "productpage",
						ServiceNamespace: namespace,
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "productpage",
						},
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
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
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
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "ratings",
						ServiceNamespace: namespace,
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "ratings",
						},
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
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
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
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "reviews",
						ServiceNamespace: namespace,
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "reviews",
						},
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
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
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
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
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
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
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

func BookInfoServices(namespace string) skkube.ServiceList {
	svc := func(name string) *skkube.Service {
		return &skkube.Service{
			Service: service.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: namespace,
					Labels: map[string]string{
						"app": name,
					},
				},
				Spec: kubev1.ServiceSpec{
					Ports: []kubev1.ServicePort{
						{
							Name:       "http",
							Port:       9080,
							Protocol:   kubev1.ProtocolTCP,
							TargetPort: intstr.IntOrString{IntVal: 9080},
						},
					},
					Selector: map[string]string{"app": name},
				},
			},
		}
	}
	return skkube.ServiceList{
		svc("details"),
		svc("productpage"),
		svc("ratings"),
		svc("reviews"),
	}
}
