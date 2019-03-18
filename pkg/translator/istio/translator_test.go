package istio

import (
	"context"

	"github.com/solo-io/supergloo/pkg/translator/utils"

	istiorbac "github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"

	"github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"

	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/translator/istio/plugins"

	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/supergloo/test/inputs"

	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

var _ = Describe("appliesToDestination", func() {
	Context("upstream selector match", func() {
		It("returns true", func() {
			applies, err := appliesToDestination("details.default.svc.cluster.local", &v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-details-v1-9080", Namespace: "default"},
						},
					},
				},
			}, inputs.BookInfoUpstreams("default"))
			Expect(err).NotTo(HaveOccurred())
			Expect(applies).To(BeTrue())
		})
	})
	Context("namespace selector match", func() {
		It("returns true", func() {
			applies, err := appliesToDestination("details.default.svc.cluster.local", &v1.PodSelector{
				SelectorType: &v1.PodSelector_NamespaceSelector_{
					NamespaceSelector: &v1.PodSelector_NamespaceSelector{
						Namespaces: []string{"default"},
					},
				},
			}, inputs.BookInfoUpstreams("default"))
			Expect(err).NotTo(HaveOccurred())
			Expect(applies).To(BeTrue())
		})
	})
	Context("label selector match", func() {
		It("returns true", func() {
			applies, err := appliesToDestination("details.default.svc.cluster.local", &v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{"version": "v1", "app": "details"},
					},
				},
			}, inputs.BookInfoUpstreams("default"))
			Expect(err).NotTo(HaveOccurred())
			Expect(applies).To(BeTrue())
		})
	})
})

var _ = Describe("labelSetsFromSelector", func() {
	Context("PodSelector_UpstreamSelector", func() {
		It("returns labels for each upstream found", func() {
			labelSets, err := labelSetsFromSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-details-v1-9080", Namespace: "default"},
							{Name: "default-reviews-v2-9080", Namespace: "default"},
							{Name: "default-reviews-9080", Namespace: "default"},
						},
					},
				},
			}, inputs.BookInfoUpstreams("default"))
			Expect(err).NotTo(HaveOccurred())
			Expect(labelSets).To(Equal([]map[string]string{
				{"version": "v1", "app": "details"},
				{"app": "reviews", "version": "v2"},
				{"app": "reviews"},
			}))
		})
	})
	Context("PodSelector_NamespaceSelector", func() {
		It("returns labels for each upstream in the namespace", func() {
			labelSets, err := labelSetsFromSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_NamespaceSelector_{
					NamespaceSelector: &v1.PodSelector_NamespaceSelector{
						Namespaces: []string{"default"},
					},
				},
			}, inputs.BookInfoUpstreams("default"))
			Expect(err).NotTo(HaveOccurred())
			Expect(labelSets).To(Equal([]map[string]string{
				{"app": "details"},
				{"version": "v1", "app": "details"},
				{"app": "productpage"},
				{"app": "productpage", "version": "v1"},
				{"app": "ratings"},
				{"app": "ratings", "version": "v1"},
				{"app": "reviews"},
				{"version": "v1", "app": "reviews"},
				{"app": "reviews", "version": "v2"},
				{"version": "v3", "app": "reviews"},
			}))
		})
	})
})

var _ = Describe("makeDestinationRule", func() {
	Context("input mesh has encryption enabled", func() {
		It("creates a destination rule with tls setting ISTIO_MUTUAL", func() {
			dr := makeDestinationRule(context.TODO(), "ns", "host", nil, true)
			Expect(dr.TrafficPolicy).NotTo(BeNil())
			Expect(dr.TrafficPolicy.Tls).NotTo(BeNil())
			Expect(dr.TrafficPolicy.Tls.Mode).To(Equal(v1alpha3.TLSSettings_ISTIO_MUTUAL))
		})
	})
	Context("input mesh has encryption enabled, input host is the kube apiserver", func() {
		It("creates a destination rule with tls setting DISABLED", func() {
			dr := makeDestinationRule(context.TODO(), "ns", "kubernetes.default.svc.cluster.local", nil, true)
			Expect(dr.TrafficPolicy).NotTo(BeNil())
			Expect(dr.TrafficPolicy.Tls).NotTo(BeNil())
			Expect(dr.TrafficPolicy.Tls.Mode).To(Equal(v1alpha3.TLSSettings_DISABLE))
		})
	})
	Context("input mesh has encryption disabled", func() {
		It("creates a destination rule with no tls setting", func() {
			dr := makeDestinationRule(context.TODO(), "ns", "host", nil, false)
			Expect(dr.TrafficPolicy).To(BeNil())
		})
	})
})

var _ = Describe("convertMatcher", func() {
	It("converts a gloo match to an istio match", func() {
		istioMatch := convertMatcher(map[string]string{"app": "details", "version": "v1"}, 1234, &gloov1.Matcher{
			PathSpecifier: &gloov1.Matcher_Exact{
				Exact: "hi",
			},
			Methods: []string{"GET", "ME", "OUTTA", "HERE"},
			Headers: []*gloov1.HeaderMatcher{
				{Name: "k", Value: "v", Regex: true},
				{Name: "a", Value: "z", Regex: false},
			},
		})
		Expect(istioMatch).To(Equal(&v1alpha3.HTTPMatchRequest{
			Uri: &v1alpha3.StringMatch{
				MatchType: &v1alpha3.StringMatch_Exact{Exact: "hi"},
			},
			Method: &v1alpha3.StringMatch{
				MatchType: &v1alpha3.StringMatch_Regex{Regex: "GET|ME|OUTTA|HERE"},
			},
			Headers: map[string]*v1alpha3.StringMatch{
				"a": {MatchType: &v1alpha3.StringMatch_Exact{Exact: "z"}},
				"k": {MatchType: &v1alpha3.StringMatch_Regex{Regex: "v"}},
			},
			Port:         1234,
			SourceLabels: map[string]string{"app": "details", "version": "v1"},
		}))
	})
})

var _ = Describe("createIstioMatcher", func() {
	Context("yes matcher, yes source labels", func() {
		It("creates a copy of each matcher for each set of source labels", func() {
			istioMatchers := createIstioMatcher(
				[]map[string]string{
					{"app": "details"},
					{"app": "reviews"},
				}, 1234, []*gloov1.Matcher{
					{
						PathSpecifier: &gloov1.Matcher_Exact{
							Exact: "hi",
						},
					},
					{
						PathSpecifier: &gloov1.Matcher_Exact{
							Exact: "bye",
						},
					},
				})
			Expect(istioMatchers).To(Equal([]*v1alpha3.HTTPMatchRequest{
				{
					Uri: &v1alpha3.StringMatch{
						MatchType: &v1alpha3.StringMatch_Exact{Exact: "hi"},
					},
					Port:         1234,
					SourceLabels: map[string]string{"app": "details"},
				},
				{
					Uri: &v1alpha3.StringMatch{
						MatchType: &v1alpha3.StringMatch_Exact{Exact: "hi"},
					},
					Port:         1234,
					SourceLabels: map[string]string{"app": "reviews"},
				},
				{
					Uri: &v1alpha3.StringMatch{
						MatchType: &v1alpha3.StringMatch_Exact{Exact: "bye"},
					},
					Port:         1234,
					SourceLabels: map[string]string{"app": "details"},
				},
				{
					Uri: &v1alpha3.StringMatch{
						MatchType: &v1alpha3.StringMatch_Exact{Exact: "bye"},
					},
					Port:         1234,
					SourceLabels: map[string]string{"app": "reviews"},
				},
			}))
		})
	})
})

type testRoutingPlugin struct {
	collectedRoutes []*v1alpha3.HTTPRoute
}

func (t *testRoutingPlugin) Init(params plugins.InitParams) error {
	return nil
}

func (t *testRoutingPlugin) ProcessRoute(params plugins.Params, in v1.RoutingRuleSpec, out *v1alpha3.HTTPRoute) error {
	t.collectedRoutes = append(t.collectedRoutes, out)
	return nil
}

var _ = Describe("createRoute", func() {
	Context("with a route plugin", func() {
		It("creates an http route with the corresponding destination, and calls the plugin for each route", func() {
			resourceErrs := make(reporter.ResourceErrors)
			plug := testRoutingPlugin{}
			t := NewTranslator([]plugins.Plugin{&plug}).(*translator)
			upstreams := inputs.BookInfoUpstreams("default")
			route := t.createRoute(
				plugins.Params{Ctx: context.TODO(), Upstreams: upstreams},
				"details.default.svc.cluster.local",
				inputs.BookInfoRoutingRules("namespace-where-rules-crds-live", nil),
				createIstioMatcher(
					[]map[string]string{
						{"app": "details"},
						{"app": "reviews"},
					}, 1234, []*gloov1.Matcher{
						{
							PathSpecifier: &gloov1.Matcher_Exact{
								Exact: "hi",
							},
						},
						{
							PathSpecifier: &gloov1.Matcher_Exact{
								Exact: "bye",
							},
						},
					}),
				upstreams,
				resourceErrs,
			)
			Expect(route.Route).To(HaveLen(1))
			Expect(route.Route[0].Destination.Host).To(Equal("details.default.svc.cluster.local"))
			Expect(plug.collectedRoutes).To(HaveLen(1))
			Expect(plug.collectedRoutes[0]).To(Equal(route))
		})
	})
})

type destinationRule struct {
	host    string
	subsets []subset
}

type subset struct {
	name   string
	labels map[string]string
}
type virtualService struct {
	host   string
	routes []route
}

type route struct {
	matches []match
}

type match struct {
	sourceLabels map[string]string
	port         uint32
}

var _ = Describe("Translator", func() {
	It("translates a snapshot into a corresponding meshconfig, returns ResourceErrors", func() {
		plug := testRoutingPlugin{}
		istioMesh := inputs.IstioMesh("fresh", nil)
		ref := istioMesh.Metadata.Ref()
		inputRoutingRules := inputs.BookInfoRoutingRules("namespace-where-rules-crds-live", &ref)

		t := NewTranslator([]plugins.Plugin{&plug}).(*translator)
		configPerMesh, resourceErrs, err := t.Translate(context.TODO(), &v1.ConfigSnapshot{
			Meshes:       map[string]v1.MeshList{"": {istioMesh}},
			Upstreams:    map[string]gloov1.UpstreamList{"": inputs.BookInfoUpstreams("default")},
			Routingrules: map[string]v1.RoutingRuleList{"": inputRoutingRules},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(configPerMesh).To(HaveKey(istioMesh))
		meshConfig := configPerMesh[istioMesh]
		Expect(meshConfig).NotTo(BeNil())
		Expect(meshConfig.DestinationRules).To(HaveLen(4))
		// we should have 1 subset for the service selector followed by 1 subset for each set of tags
		for i, expected := range []destinationRule{
			{"details.default.svc.cluster.local",
				[]subset{
					{"app-details",
						map[string]string{"app": "details"}},
					{"app-details-version-v1",
						map[string]string{"app": "details", "version": "v1"}},
				}},
			{"productpage.default.svc.cluster.local",
				[]subset{
					{"app-productpage",
						map[string]string{"app": "productpage"}},
					{"app-productpage-version-v1",
						map[string]string{"app": "productpage", "version": "v1"}},
				}},
			{"ratings.default.svc.cluster.local",
				[]subset{
					{"app-ratings",
						map[string]string{"app": "ratings"}},
					{"app-ratings-version-v1",
						map[string]string{"app": "ratings", "version": "v1"}},
				}},
			{"reviews.default.svc.cluster.local",
				[]subset{
					{"app-reviews",
						map[string]string{"app": "reviews"}},
					{"app-reviews-version-v1",
						map[string]string{"app": "reviews", "version": "v1"}},
					{"app-reviews-version-v2",
						map[string]string{"app": "reviews", "version": "v2"}},
					{"app-reviews-version-v3",
						map[string]string{"app": "reviews", "version": "v3"}},
				}},
		} {
			Expect(meshConfig.DestinationRules[i].Host).To(Equal(expected.host))
			for j, sub := range expected.subsets {
				Expect(meshConfig.DestinationRules[i].Subsets[j].Name).To(Equal(sub.name))
				Expect(meshConfig.DestinationRules[i].Subsets[j].Labels).To(Equal(sub.labels))
			}

			Expect(meshConfig.VirtualServices).To(HaveLen(4))
			for i, expected := range []virtualService{
				{"details.default.svc.cluster.local",
					[]route{{[]match{{
						map[string]string{"app": "reviews"}, 9080,
					}}}}},
				{"productpage.default.svc.cluster.local",
					[]route{{[]match{{
						map[string]string{"app": "reviews"}, 9080,
					}}}}},
				{"ratings.default.svc.cluster.local",
					[]route{{[]match{{
						map[string]string{"app": "reviews"}, 9080,
					}}}}},
				{"reviews.default.svc.cluster.local",
					[]route{{[]match{{
						map[string]string{"app": "reviews"}, 9080,
					}}}}},
			} {
				Expect(meshConfig.VirtualServices[i].Metadata.Name).To(Equal(utils.SanitizeName(expected.host)))
				Expect(meshConfig.VirtualServices[i].Hosts).To(Equal([]string{expected.host}))
				Expect(meshConfig.VirtualServices[i].Gateways).To(Equal([]string{"mesh"}))
				Expect(meshConfig.VirtualServices[i].Http).To(HaveLen(len(inputRoutingRules)))
				for j, rr := range inputRoutingRules {
					if reqMatchers := len(rr.RequestMatchers); reqMatchers > 0 {
						Expect(meshConfig.VirtualServices[i].Http[j].Match).To(HaveLen(reqMatchers))
					} else {
						Expect(meshConfig.VirtualServices[i].Http[j].Match[0].Port).To(Equal(expected.routes[j].matches[0].port))
						Expect(meshConfig.VirtualServices[i].Http[j].Match[0].SourceLabels).To(Equal(expected.routes[j].matches[0].sourceLabels))
					}
				}
			}
		}
		Expect(meshConfig.MeshPolicy).NotTo(BeNil())
		Expect(meshConfig.MeshPolicy.Metadata.Name).To(Equal("default"))
		Expect(meshConfig.MeshPolicy.Peers).To(HaveLen(1))
		Expect(meshConfig.MeshPolicy.Peers[0].Params).To(Equal(&v1alpha1.PeerAuthenticationMethod_Mtls{
			Mtls: &v1alpha1.MutualTls{
				Mode: v1alpha1.MutualTls_STRICT,
			},
		}))

		Expect(resourceErrs).NotTo(BeNil())
	})
	It("sets the root cert on the meshconfig if specified on the mesh", func() {
		plug := testRoutingPlugin{}
		tlsSecret := &v1.TlsSecret{
			Metadata:  core.Metadata{Namespace: "mynamespace", Name: "some-tls-secret"},
			RootCert:  "RootCert",
			CertChain: "CertChain",
			CaCert:    "CaCert",
			CaKey:     "CaKey",
		}
		secretRef := tlsSecret.Metadata.Ref()
		istioMesh := inputs.IstioMesh("fresh", &secretRef)

		t := NewTranslator([]plugins.Plugin{&plug}).(*translator)
		configPerMesh, _, err := t.Translate(context.TODO(), &v1.ConfigSnapshot{
			Meshes:     map[string]v1.MeshList{"": {istioMesh}},
			Tlssecrets: map[string]v1.TlsSecretList{"": {tlsSecret}},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(configPerMesh).To(HaveKey(istioMesh))
		meshConfig := configPerMesh[istioMesh]
		Expect(meshConfig).NotTo(BeNil())
		Expect(meshConfig.RootCert).NotTo(BeNil())
		Expect(*meshConfig.RootCert).To(Equal(v1.TlsSecret{
			Metadata: core.Metadata{
				Namespace: istioMesh.MeshType.(*v1.Mesh_Istio_).Istio.InstallationNamespace,
				Name:      "cacerts",
			},
			RootCert:  tlsSecret.RootCert,
			CertChain: tlsSecret.CertChain,
			CaCert:    tlsSecret.CaCert,
			CaKey:     tlsSecret.CaKey,
		}))
	})
	It("errors if the root cert is not found", func() {
		plug := testRoutingPlugin{}
		tlsSecret := &v1.TlsSecret{
			Metadata:  core.Metadata{Namespace: "mynamespace", Name: "some-tls-secret"},
			RootCert:  "RootCert",
			CertChain: "CertChain",
			CaCert:    "CaCert",
			CaKey:     "CaKey",
		}
		secretRef := tlsSecret.Metadata.Ref()
		istioMesh := inputs.IstioMesh("fresh", &secretRef)

		t := NewTranslator([]plugins.Plugin{&plug}).(*translator)
		_, errs, err := t.Translate(context.TODO(), &v1.ConfigSnapshot{
			Meshes:     map[string]v1.MeshList{"": {istioMesh}},
			Tlssecrets: map[string]v1.TlsSecretList{"": {}},
		})
		Expect(err).NotTo(HaveOccurred())
		err = errs.Validate()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("finding tls secret for mesh root cert"))
	})
})

var _ = Describe("hostsForSelector", func() {
	It("returns the list of hostnames for the selected upstreams", func() {
		hosts, err := hostsForSelector(&v1.PodSelector{
			SelectorType: &v1.PodSelector_UpstreamSelector_{
				UpstreamSelector: &v1.PodSelector_UpstreamSelector{
					Upstreams: []core.ResourceRef{
						{Name: "default-details-v1-9080", Namespace: "default"},
						{Name: "default-reviews-v3-9080", Namespace: "default"},
					},
				},
			},
		}, inputs.BookInfoUpstreams("default"))
		Expect(err).NotTo(HaveOccurred())
		Expect(hosts).To(Equal([]string{
			"details.default.svc.cluster.local",
			"reviews.default.svc.cluster.local",
		}))
	})
})

var _ = Describe("createServiceRoleFromRule", func() {
	It("creates a service role allowing access to the destinations on the specified paths and methods", func() {
		rule := &v1.SecurityRule{
			Metadata: core.Metadata{
				Namespace: "should-inherit-this-namespace",
				Name:      "should-inherit-this-name",
			},
			SourceSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-details-v1-9080", Namespace: "default"},
							{Name: "default-reviews-v3-9080", Namespace: "default"},
						},
					},
				},
			},
			DestinationSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-productpage-v1-9080", Namespace: "default"},
							{Name: "default-ratings-9080", Namespace: "default"},
						},
					},
				},
			},
			AllowedMethods: []string{"GET", "POST"},
			AllowedPaths:   []string{"/a", "/b"},
		}
		serviceRole, err := createServiceRoleFromRule(
			"somenamespace",
			rule,
			inputs.BookInfoUpstreams("default"))
		Expect(err).NotTo(HaveOccurred())
		Expect(serviceRole.Metadata.Name).To(Equal(rule.Metadata.Namespace + "-" + rule.Metadata.Name))
		Expect(serviceRole.Rules).To(HaveLen(1))
		Expect(serviceRole.Rules[0].Paths).To(Equal(rule.AllowedPaths))
		Expect(serviceRole.Rules[0].Methods).To(Equal(rule.AllowedMethods))
		Expect(serviceRole.Rules[0].Services).To(Equal([]string{
			"productpage.default.svc.cluster.local",
			"ratings.default.svc.cluster.local",
		}))
	})
})
var _ = Describe("getSubjectsForSelector", func() {
	It("extracts the service account name from each pod which is selected", func() {
		subjects, err := getSubjectsForSelector(
			&v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-details-v1-9080", Namespace: "default"},
							{Name: "default-reviews-v3-9080", Namespace: "default"},
						},
					},
				},
			},
			inputs.BookInfoUpstreams("default"),
			inputs.BookInfoPods("default"),
		)
		Expect(err).NotTo(HaveOccurred())
		Expect(subjects).To(Equal([]*istiorbac.Subject{
			{Properties: map[string]string{"source.principal": "cluster.local/ns/default/sa/details-v1-pod-1"}},
			{Properties: map[string]string{"source.principal": "cluster.local/ns/default/sa/reviews-v3-pod-1"}},
		}))
	})
})

var _ = Describe("createServiceRoleBinding", func() {
	It("creates a service role allowing access to the destinations on the specified paths and methods", func() {
		serviceRoleBinding, err := createServiceRoleBinding(
			"somename",
			"somenamespace",
			&v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-details-v1-9080", Namespace: "default"},
							{Name: "default-reviews-v3-9080", Namespace: "default"},
						},
					},
				},
			},
			inputs.BookInfoUpstreams("default"),
			inputs.BookInfoPods("default"))
		Expect(err).NotTo(HaveOccurred())
		Expect(serviceRoleBinding.Metadata.Name).To(Equal("somename"))
		Expect(serviceRoleBinding.RoleRef.Kind).To(Equal("ServiceRole"))
		Expect(serviceRoleBinding.RoleRef.Name).To(Equal("somename"))
		Expect(serviceRoleBinding.Subjects).To(Equal([]*istiorbac.Subject{
			{Properties: map[string]string{"source.principal": "cluster.local/ns/default/sa/details-v1-pod-1"}},
			{Properties: map[string]string{"source.principal": "cluster.local/ns/default/sa/reviews-v3-pod-1"}},
		}))
	})
})

var _ = Describe("createSecurityConfig", func() {
	It("creates a complete config with serviceroles, servicerolebindings, and rbacconfig", func() {
		rule11 := &v1.SecurityRule{
			Metadata: core.Metadata{
				Namespace: "ns1",
				Name:      "rule1",
			},
			SourceSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-details-v1-9080", Namespace: "default"},
							{Name: "default-reviews-v3-9080", Namespace: "default"},
						},
					},
				},
			},
			DestinationSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-productpage-v1-9080", Namespace: "default"},
							{Name: "default-ratings-9080", Namespace: "default"},
						},
					},
				},
			},
			AllowedMethods: []string{"GET", "POST"},
			AllowedPaths:   []string{"/a", "/b"},
		}

		rule12 := &v1.SecurityRule{
			Metadata: core.Metadata{
				Namespace: "ns1",
				Name:      "rule2",
			},
			SourceSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-productpage-v1-9080", Namespace: "default"},
							{Name: "default-ratings-v1-9080", Namespace: "default"},
						},
					},
				},
			},
			DestinationSelector: &v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: []core.ResourceRef{
							{Name: "default-details-v1-9080", Namespace: "default"},
							{Name: "default-reviews-9080", Namespace: "default"},
						},
					},
				},
			},
			AllowedMethods: []string{"GET", "POST"},
			AllowedPaths:   []string{"/a", "/b"},
		}

		rules := v1.SecurityRuleList{rule11, rule12}

		resourceErrs := make(reporter.ResourceErrors)

		securityConfig := createSecurityConfig(
			"somenamespace",
			rules,
			inputs.BookInfoUpstreams("default"),
			inputs.BookInfoPods("default"),
			resourceErrs,
		)
		Expect(securityConfig.RbacConfig).NotTo(BeNil())
		Expect(securityConfig.RbacConfig.Metadata.Name).To(Equal("default"))
		Expect(securityConfig.RbacConfig.Mode).To(Equal(istiorbac.RbacConfig_ON))
		Expect(securityConfig.RbacConfig.EnforcementMode).To(Equal(istiorbac.EnforcementMode_ENFORCED))
		Expect(securityConfig.ServiceRoles).To(HaveLen(2))
		Expect(securityConfig.ServiceRoles[0].Metadata.Name).To(Equal(rule11.Metadata.Namespace + "-" + rule11.Metadata.Name))
		Expect(securityConfig.ServiceRoles[0].Rules[0].Services).To(Equal([]string{
			"productpage.default.svc.cluster.local",
			"ratings.default.svc.cluster.local",
		}))
		Expect(securityConfig.ServiceRoles[1].Metadata.Name).To(Equal(rule12.Metadata.Namespace + "-" + rule12.Metadata.Name))
		Expect(securityConfig.ServiceRoles[1].Rules[0].Services).To(Equal([]string{
			"details.default.svc.cluster.local",
			"reviews.default.svc.cluster.local",
		}))
		Expect(securityConfig.ServiceRoleBindings).To(HaveLen(2))
		Expect(securityConfig.ServiceRoleBindings[0].Metadata.Name).To(Equal(rule11.Metadata.Namespace + "-" + rule11.Metadata.Name))
		Expect(securityConfig.ServiceRoleBindings[0].RoleRef.Name).To(Equal(securityConfig.ServiceRoles[0].Metadata.Name))
		Expect(securityConfig.ServiceRoleBindings[1].Metadata.Name).To(Equal(rule12.Metadata.Namespace + "-" + rule12.Metadata.Name))
		Expect(securityConfig.ServiceRoleBindings[1].RoleRef.Name).To(Equal(securityConfig.ServiceRoles[1].Metadata.Name))
	})
})
