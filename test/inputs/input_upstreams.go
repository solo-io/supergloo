package inputs

import (
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/kubernetes"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/rest"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/transformation"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

func BookInfoUpstreams() v1.UpstreamList {
	return v1.UpstreamList{
		&v1.Upstream{
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "details",
						ServiceNamespace: "default",
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "details",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-details-9080",
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "productpage",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-productpage-9080",
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "ratings",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-ratings-9080",
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "reviews",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-reviews-9080",
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
				Labels: map[string]string{
					"discovered_by": "kubernetesplugin",
					"app":           "reviews",
				},
			},
		},
	}
}

func InputUpstreams() v1.UpstreamList {
	return v1.UpstreamList{
		&v1.Upstream{
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "details",
						ServiceNamespace: "default",
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "details",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-details-9080",
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
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
						ServiceName:      "kubernetes",
						ServiceNamespace: "default",
						ServicePort:      0x000001bb,
						Selector:         map[string]string{},
						ServiceSpec:      (*plugins.ServiceSpec)(nil),
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-kubernetes-443",
				Namespace: "gloo-system",
				Labels: map[string]string{
					"component":     "apiserver",
					"discovered_by": "kubernetesplugin",
					"provider":      "kubernetes",
				},
			},
		},
		&v1.Upstream{
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "petstore",
						ServiceNamespace: "default",
						ServicePort:      0x00001f90,
						Selector: map[string]string{
							"app": "petstore",
						},
						ServiceSpec: &plugins.ServiceSpec{
							PluginType: &plugins.ServiceSpec_Rest{
								Rest: &rest.ServiceSpec{
									Transformations: map[string]*transformation.TransformationTemplate{
										"deletePet": &transformation.TransformationTemplate{
											AdvancedTemplates: false,
											Extractors:        map[string]*transformation.Extraction{},
											Headers: map[string]*transformation.InjaTemplate{
												":method": &transformation.InjaTemplate{
													Text: "DELETE",
												},
												":path": &transformation.InjaTemplate{
													Text: "/api/pets/{{ default(id, \"\") }}",
												},
												"content-type": &transformation.InjaTemplate{
													Text: "application/json",
												},
											},
											BodyTransformation: nil,
										},
										"findPetById": &transformation.TransformationTemplate{
											AdvancedTemplates: false,
											Extractors:        map[string]*transformation.Extraction{},
											Headers: map[string]*transformation.InjaTemplate{
												":method": &transformation.InjaTemplate{
													Text: "GET",
												},
												":path": &transformation.InjaTemplate{
													Text: "/api/pets/{{ default(id, \"\") }}",
												},
												"content-length": &transformation.InjaTemplate{
													Text: "0",
												},
												"content-type": &transformation.InjaTemplate{
													Text: "",
												},
												"transfer-encoding": &transformation.InjaTemplate{
													Text: "",
												},
											},
											BodyTransformation: &transformation.TransformationTemplate_Body{
												Body: &transformation.InjaTemplate{
													Text: "",
												},
											},
										},
										"findPets": &transformation.TransformationTemplate{
											AdvancedTemplates: false,
											Extractors:        map[string]*transformation.Extraction{},
											Headers: map[string]*transformation.InjaTemplate{
												":method": &transformation.InjaTemplate{
													Text: "GET",
												},
												":path": &transformation.InjaTemplate{
													Text: "/api/pets?tags={{default(tags, \"\")}}&limit={{default(limit, \"\")}}",
												},
												"content-length": &transformation.InjaTemplate{
													Text: "0",
												},
												"content-type": &transformation.InjaTemplate{
													Text: "",
												},
												"transfer-encoding": &transformation.InjaTemplate{
													Text: "",
												},
											},
											BodyTransformation: &transformation.TransformationTemplate_Body{
												Body: &transformation.InjaTemplate{
													Text: "",
												},
											},
										},
										"addPet": &transformation.TransformationTemplate{
											AdvancedTemplates: false,
											Extractors:        map[string]*transformation.Extraction{},
											Headers: map[string]*transformation.InjaTemplate{
												":method": &transformation.InjaTemplate{
													Text: "POST",
												},
												":path": &transformation.InjaTemplate{
													Text: "/api/pets",
												},
												"content-type": &transformation.InjaTemplate{
													Text: "application/json",
												},
											},
											BodyTransformation: &transformation.TransformationTemplate_Body{
												Body: &transformation.InjaTemplate{
													Text: "{\"id\": {{ default(id, \"\") }},\"name\": \"{{ default(name, \"\")}}\",\"tag\": \"{{ default(tag, \"\")}}\"}",
												},
											},
										},
									},
									SwaggerInfo: &rest.ServiceSpec_SwaggerInfo{
										SwaggerSpec: &rest.ServiceSpec_SwaggerInfo_Url{
											Url: "http://petstore.default.svc.cluster.local:8080/swagger.json",
										},
									},
								},
							},
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-petstore-8080",
				Namespace: "gloo-system",
				Labels: map[string]string{
					"discovered_by": "kubernetesplugin",
					"service":       "petstore",
				},
			},
		},
		&v1.Upstream{
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "productpage",
						ServiceNamespace: "default",
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "productpage",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-productpage-9080",
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "ratings",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-ratings-9080",
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
						ServicePort:      0x00002378,
						Selector: map[string]string{
							"app": "reviews",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "default-reviews-9080",
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
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
						ServiceNamespace: "default",
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
				Namespace: "gloo-system",
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
						ServiceName:      "gateway-proxy",
						ServiceNamespace: "gloo-system",
						ServicePort:      0x00001f90,
						Selector: map[string]string{
							"gloo": "gateway-proxy",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "gloo-system-gateway-proxy-8080",
				Namespace: "gloo-system",
				Labels: map[string]string{
					"app":           "gloo",
					"discovered_by": "kubernetesplugin",
					"gloo":          "gateway-proxy",
				},
			},
		},
		&v1.Upstream{
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "gloo",
						ServiceNamespace: "gloo-system",
						ServicePort:      0x000026f9,
						Selector: map[string]string{
							"gloo": "gloo",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "gloo-system-gloo-9977",
				Namespace: "gloo-system",
				Labels: map[string]string{
					"discovered_by": "kubernetesplugin",
					"gloo":          "gloo",
					"app":           "gloo",
				},
			},
		},
		&v1.Upstream{
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "kube-dns",
						ServiceNamespace: "kube-system",
						ServicePort:      0x00000035,
						Selector: map[string]string{
							"k8s-app": "kube-dns",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "kube-system-kube-dns-53",
				Namespace: "gloo-system",
				Labels: map[string]string{
					"kubernetes.io/cluster-service": "true",
					"kubernetes.io/name":            "KubeDNS",
					"discovered_by":                 "kubernetesplugin",
					"k8s-app":                       "kube-dns",
				},
			},
		},
		&v1.Upstream{
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "kubernetes-dashboard",
						ServiceNamespace: "kube-system",
						ServicePort:      0x00000050,
						Selector: map[string]string{
							"app": "kubernetes-dashboard",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "kube-system-kubernetes-dashboard-80",
				Namespace: "gloo-system",
				Labels: map[string]string{
					"addonmanager.kubernetes.io/mode":        "Reconcile",
					"app":                                    "kubernetes-dashboard",
					"discovered_by":                          "kubernetesplugin",
					"kubernetes.io/minikube-addons":          "dashboard",
					"kubernetes.io/minikube-addons-endpoint": "dashboard",
				},
			},
		},
		&v1.Upstream{
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "kubernetes-dashboard",
						ServiceNamespace: "kube-system",
						ServicePort:      0x00000050,
						Selector: map[string]string{
							"app":                             "kubernetes-dashboard",
							"version":                         "v1.10.1",
							"addonmanager.kubernetes.io/mode": "Reconcile",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "kube-system-kubernetes-dashboard-reconcile-v1-10-1-80",
				Namespace: "gloo-system",
				Labels: map[string]string{
					"kubernetes.io/minikube-addons":          "dashboard",
					"kubernetes.io/minikube-addons-endpoint": "dashboard",
					"addonmanager.kubernetes.io/mode":        "Reconcile",
					"app":                                    "kubernetes-dashboard",
					"discovered_by":                          "kubernetesplugin",
				},
			},
		},
		&v1.Upstream{
			UpstreamSpec: &v1.UpstreamSpec{
				UpstreamType: &v1.UpstreamSpec_Kube{
					Kube: &kubernetes.UpstreamSpec{
						ServiceName:      "tiller-deploy",
						ServiceNamespace: "kube-system",
						ServicePort:      0x0000ac66,
						Selector: map[string]string{
							"app":  "helm",
							"name": "tiller",
						},
					},
				},
			},
			Metadata: core.Metadata{
				Name:      "kube-system-tiller-deploy-44134",
				Namespace: "gloo-system",
				Labels: map[string]string{
					"name":          "tiller",
					"app":           "helm",
					"discovered_by": "kubernetesplugin",
				},
			},
		},
	}
}
