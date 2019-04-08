package appmesh

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
	"github.com/solo-io/supergloo/pkg/api/custom/clients/kubernetes"
	kubecustom "github.com/solo-io/supergloo/pkg/api/external/kubernetes/core/v1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/webhook/test"
)

var appMesh = func(name string) *v1.Mesh {
	if name == "" {
		name = "appmesh"
	}
	return &v1.Mesh{
		Metadata: core.Metadata{
			Name:      name,
			Namespace: "supergloo-system",
		},
		MeshType: &v1.Mesh_AwsAppMesh{
			AwsAppMesh: &v1.AwsAppMesh{
				Region:           "us-east-1",
				VirtualNodeLabel: "app",
				EnableAutoInject: true,
				SidecarPatchConfigMap: &core.ResourceRef{
					Name:      "sidecar-injector-webhook-configmap",
					Namespace: "supergloo-system",
				},
				InjectionSelector: &v1.PodSelector{
					SelectorType: &v1.PodSelector_NamespaceSelector_{
						NamespaceSelector: &v1.PodSelector_NamespaceSelector{
							Namespaces: []string{"namespace-with-inject"}}}}}}}
}

var _ = Describe("config translator", func() {
	var (
		injectedPodList kubecustom.PodList
		upstreamList    gloov1.UpstreamList
	)
	//
	// var defaultConfig = func() *awsAppMeshConfiguration {
	// 	mesh := appMesh("")
	// 	config, err := NewAwsAppMeshConfiguration(mesh, injectedPodList, upstreamList)
	// 	Expect(err).NotTo(HaveOccurred())
	// 	err = config.AllowAll()
	// 	Expect(err).NotTo(HaveOccurred())
	// 	typedConfig, ok := config.(*awsAppMeshConfiguration)
	// 	Expect(ok).To(BeTrue())
	// 	return typedConfig
	// }
	BeforeEach(func() {
		clients.UseMemoryClients()
		injectedPodList = test.MustGetInjectedPodList()
		upstreamList = test.MustGetUpstreamList()
	})

	Context("get pod info", func() {
		It("can get valid virtual node name", func() {
			kubePod := injectedPodList[0]
			pod, err := kubernetes.ToKube(kubePod)
			Expect(err).NotTo(HaveOccurred())
			info, err := getPodInfo(appMesh(""), pod)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.virtualNodeName).To(Equal("productpage"))
			Expect(info.ports).To(Equal([]uint32{9080}))
		})
		It("will return nil, if meshname doesn't match virtual node path", func() {
			kubePod := injectedPodList[0]
			pod, err := kubernetes.ToKube(kubePod)
			Expect(err).NotTo(HaveOccurred())
			info, err := getPodInfo(appMesh("123"), pod)
			Expect(err).NotTo(HaveOccurred())
			Expect(info).To(BeNil())
		})
	})
	Context("get pods for mesh", func() {
		It("can filter valid mesh pods", func() {
			_, podList, err := getPodsForMesh(appMesh(""), injectedPodList)
			Expect(err).NotTo(HaveOccurred())
			Expect(podList).To(HaveLen(6))
		})
	})

	Context("get upstreams for mesh", func() {
		It("can get valid upstreams for the mesh", func() {
			info, podList, err := getPodsForMesh(appMesh(""), injectedPodList)
			Expect(err).NotTo(HaveOccurred())
			_, usList, err := getUpstreamsForMesh(upstreamList, info, podList)
			Expect(err).NotTo(HaveOccurred())
			Expect(usList).To(HaveLen(10))
		})
	})
	Context("allow all", func() {
		It("can create the proper config for allow all", func() {
			mesh := appMesh("")
			config, err := NewAwsAppMeshConfiguration(mesh, injectedPodList, upstreamList)
			Expect(err).NotTo(HaveOccurred())
			err = config.AllowAll()
			Expect(err).NotTo(HaveOccurred())
			typedConfig, ok := config.(*awsAppMeshConfiguration)
			Expect(ok).To(BeTrue())
			Expect(typedConfig.MeshName).To(Equal(mesh.Metadata.Name))
		})
	})

	Context("Routing Rules", func() {
		// Context("matchers", func() {
		// 	It("cannot have 0 matchers", func() {
		// 		routingRule := &v1.RoutingRule{
		// 			Spec: &v1.RoutingRuleSpec{
		// 				RuleType: &v1.RoutingRuleSpec_FaultInjection{},
		// 			},
		// 		}
		// 		typedConfig := defaultConfig()
		// 		err := typedConfig.HandleRoutingRule(routingRule)
		// 		Expect(err).To(HaveOccurred())
		// 		Expect(err.Error()).To(ContainSubstring("appmesh requires exactly one matcher, 0 found"))
		// 	})
		//
		// 	It("Cannot have > 2 matchers", func() {
		// 		routingRule := &v1.RoutingRule{
		// 			Spec: &v1.RoutingRuleSpec{
		// 				RuleType: &v1.RoutingRuleSpec_FaultInjection{},
		// 			},
		// 			RequestMatchers: []*gloov1.Matcher{
		// 				{}, {},
		// 			},
		// 		}
		// 		typedConfig := defaultConfig()
		// 		err := typedConfig.HandleRoutingRule(routingRule)
		// 		Expect(err).To(HaveOccurred())
		// 		Expect(err.Error()).To(ContainSubstring("appmesh requires exactly one matcher, 2 found"))
		// 	})
		// 	It("must be a prefix matcher", func() {
		// 		routingRule := &v1.RoutingRule{
		// 			Spec: &v1.RoutingRuleSpec{
		// 				RuleType: &v1.RoutingRuleSpec_FaultInjection{},
		// 			},
		// 			RequestMatchers: []*gloov1.Matcher{
		// 				{
		// 					PathSpecifier: &gloov1.Matcher_Exact{},
		// 				},
		// 			},
		// 		}
		// 		typedConfig := defaultConfig()
		// 		err := typedConfig.HandleRoutingRule(routingRule)
		// 		Expect(err).To(HaveOccurred())
		// 		Expect(err.Error()).To(ContainSubstring("unsupported matcher type found"))
		// 	})
		// })
		//
		// It("can only handle traffic shifting", func() {
		// 	routingRule := &v1.RoutingRule{
		// 		Spec: &v1.RoutingRuleSpec{
		// 			RuleType: &v1.RoutingRuleSpec_FaultInjection{},
		// 		},
		// 		RequestMatchers: []*gloov1.Matcher{
		// 			{
		// 				PathSpecifier: &gloov1.Matcher_Prefix{
		// 					Prefix: "/",
		// 				},
		// 			},
		// 		},
		// 	}
		// 	typedConfig := defaultConfig()
		// 	err := typedConfig.HandleRoutingRule(routingRule)
		// 	Expect(err).To(HaveOccurred())
		// 	Expect(err.Error()).To(ContainSubstring("currently only traffic shifting rules are supported by appmesh"))
		// })
		// It("can handle traffic shifting", func() {
		// 	routingRule := &v1.RoutingRule{
		// 		Spec: &v1.RoutingRuleSpec{
		// 			RuleType: &v1.RoutingRuleSpec_TrafficShifting{},
		// 		},
		// 		RequestMatchers: []*gloov1.Matcher{
		// 			{
		// 				PathSpecifier: &gloov1.Matcher_Prefix{
		// 					Prefix: "/",
		// 				},
		// 			},
		// 		},
		// 	}
		// 	typedConfig := defaultConfig()
		// 	err := typedConfig.HandleRoutingRule(routingRule)
		// 	Expect(err).NotTo(HaveOccurred())
		// })

	})
})
