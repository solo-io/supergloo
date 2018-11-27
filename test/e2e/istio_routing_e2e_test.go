package e2e

import (
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/kubeutils"
	"github.com/solo-io/solo-kit/test/helpers"
	testsetup "github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	"github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/setup"
	"github.com/solo-io/supergloo/test/utils"
)

var _ = Describe("istio routing E2e", func() {
	var namespace, releaseName string
	path := os.Getenv("HELM_CHART_PATH_ISTIO")
	if path == "" {
		Skip("Set environment variable HELM_CHART_PATH")
	}
	BeforeEach(func() {
		releaseName = "istio-release-test-" + helpers.RandString(8)
		namespace = "istio-routing-test-" + helpers.RandString(8)
		err := testsetup.SetupKubeForTest(namespace)
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		exec.Command("helm", "delete", releaseName, "--purge").Run()
		testsetup.TeardownKube(namespace)
	})

	It("works", func() {
		go setup.Main(namespace)

		// start discovery
		cmd := exec.Command(PathToUds, "--namespace", namespace)
		cmd.Env = os.Environ()
		_, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		meshes, routingRules, installClient, err := run()
		Expect(err).NotTo(HaveOccurred())

		installClient.Register()
		// wait for supergloo to register crds
		Eventually(func() error {
			_, err := installClient.List(namespace, clients.ListOpts{})
			return err
		}, time.Second*5).Should(Not(HaveOccurred()))

		installName := setupInstall(installClient, namespace, releaseName, path)

		var ref *core.ResourceRef
		Eventually(func() (*core.ResourceRef, error) {
			mesh, err := meshes.Read(namespace, installName, clients.ReadOpts{})
			if err != nil {
				return nil, err
			}
			r := mesh.Metadata.Ref()
			ref = &r
			return ref, nil
		}, time.Second*120).Should(Not(BeNil()))

		cfg, err := kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		err = utils.DeployBookinfo(cfg, namespace)
		Expect(err).NotTo(HaveOccurred())
		err = utils.DeployTestRunner(cfg, namespace)
		Expect(err).NotTo(HaveOccurred())

		setupV1RoutingRule(routingRules, namespace, ref)

		drClient, vsClient, err := istioClients()
		Expect(err).NotTo(HaveOccurred())
		// we want to see that the appropriate istio crds have been written

		// get the destination rule for
		var reviewsDestinationRule *v1alpha3.DestinationRule
		Eventually(func() []*v1alpha3.Subset {
			drs, err := drClient.List(namespace, clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())
			for _, dr := range drs {
				if dr.Host == "reviews."+namespace+".svc.cluster.local" {
					reviewsDestinationRule = dr
					return reviewsDestinationRule.Subsets
				}
			}
			return nil
		}, time.Second*45).Should(Equal([]*v1alpha3.Subset{
			{Labels: map[string]string{"app": "reviews"}, Name: "app-reviews"},
			{Labels: map[string]string{"app": "reviews", "version": "v1"}, Name: "app-reviews-version-v1"},
			{Labels: map[string]string{"app": "reviews", "version": "v2"}, Name: "app-reviews-version-v2"},
			{Labels: map[string]string{"app": "reviews", "version": "v3"}, Name: "app-reviews-version-v3"},
		}))

		var testVirtualService *v1alpha3.VirtualService
		Eventually(func() *v1alpha3.VirtualService {
			vss, err := vsClient.List(namespace, clients.ListOpts{})
			Expect(err).NotTo(HaveOccurred())
			for _, vs := range vss {
				for _, host := range vs.Hosts {
					if host == "reviews."+namespace+".svc.cluster.local" {
						testVirtualService = vs
						return testVirtualService
					}
				}
			}
			return nil
		}, time.Second*2).Should(Not(BeNil()))
		Expect(testVirtualService.Gateways).To(Equal([]string{"mesh"}))
		Expect(testVirtualService.Http).To(HaveLen(1))

		// reviews v1
		Eventually(func() string {
			resp, err := testsetup.Curl(testsetup.CurlOpts{
				Method:  "GET",
				Path:    "/reviews/1",
				Service: "reviews",
				Port:    9080,
			})
			Expect(err).NotTo(HaveOccurred())
			return resp
		}, time.Second*10).Should(ContainSubstring(`{"id": "1","reviews": [{  "reviewer": "Reviewer1",  "text": "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!"},{  "reviewer": "Reviewer2",  "text": "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare."}]}`))

		setupV2RoutingRule(routingRules, namespace, ref)

		// reviews v2
		Eventually(func() string {
			resp, err := testsetup.Curl(testsetup.CurlOpts{
				Method:  "GET",
				Path:    "/reviews/1",
				Service: "reviews",
				Port:    9080,
			})
			Expect(err).NotTo(HaveOccurred())
			return resp
		}, time.Second*10).Should(ContainSubstring(`{"id": "1","reviews": [{  "reviewer": "Reviewer1",  "text": "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!", "rating": {"stars": 5, "color": "black"}},{  "reviewer": "Reviewer2",  "text": "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare.", "rating": {"stars": 4, "color": "black"}}]}`))

		setupFaultRoutingRule(routingRules, namespace, ref)

		Eventually(func() string {
			resp, err := testsetup.Curl(testsetup.CurlOpts{
				Method:  "GET",
				Path:    "/reviews/1",
				Service: "reviews",
				Port:    9080,
			})
			Expect(err).NotTo(HaveOccurred())
			return resp
		}, time.Second*20).Should(Not(ContainSubstring(`HTTP/1.1 566 Unknown`)))

	})
})

func setupInstall(installClient v1.InstallClient, namespace, releaseName string, chartPath string) string {
	installMeta := core.Metadata{Name: releaseName, Namespace: namespace}
	installClient.Delete(installMeta.Namespace, installMeta.Name, clients.DeleteOpts{})

	install1, err := installClient.Write(&v1.Install{
		Metadata: installMeta,
		MeshType: &v1.Install_Istio{
			Istio: &v1.Istio{
				InstallationNamespace: namespace,
			},
		},
		ChartLocator: &v1.HelmChartLocator{
			Kind: &v1.HelmChartLocator_ChartPath{
				ChartPath: &v1.HelmChartPath{
					Path: chartPath,
				},
			},
		},
		Encryption: &v1.Encryption{
			TlsEnabled: true,
		},
	}, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())
	Expect(install1).NotTo(BeNil())
	return install1.Metadata.Name
}

func setupV1RoutingRule(routingRules v1.RoutingRuleClient, namespace string, targetMesh *core.ResourceRef) {
	rrMeta := core.Metadata{Name: "reviews-trafficshifting", Namespace: namespace}
	routingRules.Delete(rrMeta.Namespace, rrMeta.Name, clients.DeleteOpts{})
	rr1, err := routingRules.Write(&v1.RoutingRule{
		Metadata:   rrMeta,
		TargetMesh: targetMesh,
		Destinations: []*core.ResourceRef{{
			Name:      namespace + "-reviews-9080",
			Namespace: namespace,
		}},
		TrafficShifting: &v1.TrafficShifting{
			Destinations: []*v1.WeightedDestination{
				{
					Upstream: &core.ResourceRef{
						Name:      namespace + "-reviews-v1-9080",
						Namespace: namespace,
					},
					Weight: 100,
				},
			},
		},
	}, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())
	Expect(rr1).NotTo(BeNil())
}

func setupV2RoutingRule(routingRules v1.RoutingRuleClient, namespace string, targetMesh *core.ResourceRef) {
	rrMeta := core.Metadata{Name: "reviews-trafficshifting", Namespace: namespace}
	routingRules.Delete(rrMeta.Namespace, rrMeta.Name, clients.DeleteOpts{})
	rr1, err := routingRules.Write(&v1.RoutingRule{
		Metadata:   rrMeta,
		TargetMesh: targetMesh,
		Destinations: []*core.ResourceRef{{
			Name:      namespace + "-reviews-9080",
			Namespace: namespace,
		}},
		TrafficShifting: &v1.TrafficShifting{
			Destinations: []*v1.WeightedDestination{
				{
					Upstream: &core.ResourceRef{
						Name:      namespace + "-reviews-v2-9080",
						Namespace: namespace,
					},
					Weight: 100,
				},
			},
		},
	}, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())
	Expect(rr1).NotTo(BeNil())
}

func setupFaultRoutingRule(routingRules v1.RoutingRuleClient, namespace string, targetMesh *core.ResourceRef) {
	rrMeta := core.Metadata{Name: "reviews-fault", Namespace: namespace}
	routingRules.Delete(rrMeta.Namespace, rrMeta.Name, clients.DeleteOpts{})
	rr1, err := routingRules.Write(&v1.RoutingRule{
		Metadata:   rrMeta,
		TargetMesh: targetMesh,
		Destinations: []*core.ResourceRef{{
			Name:      namespace + "-reviews-9080",
			Namespace: namespace,
		}},
		FaultInjection: &v1alpha3.HTTPFaultInjection{
			Abort: &v1alpha3.HTTPFaultInjection_Abort{
				ErrorType: &v1alpha3.HTTPFaultInjection_Abort_HttpStatus{
					HttpStatus: 566,
				},
				Percent: 100,
			},
		},
	}, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())
	Expect(rr1).NotTo(BeNil())
}

func run() (v1.MeshClient, v1.RoutingRuleClient, v1.InstallClient, error) {
	kubeCache := kube.NewKubeCache()
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, nil, nil, err
	}
	meshClient, err := v1.NewMeshClient(&factory.KubeResourceClientFactory{
		Crd:         v1.MeshCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	if err := meshClient.Register(); err != nil {
		return nil, nil, nil, err
	}

	routingRuleClient, err := v1.NewRoutingRuleClient(&factory.KubeResourceClientFactory{
		Crd:         v1.RoutingRuleCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	if err := routingRuleClient.Register(); err != nil {
		return nil, nil, nil, err
	}

	installClient, err := v1.NewInstallClient(&factory.KubeResourceClientFactory{
		Crd:         v1.InstallCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	if err := routingRuleClient.Register(); err != nil {
		return nil, nil, nil, err
	}
	return meshClient, routingRuleClient, installClient, nil
}

func istioClients() (v1alpha3.DestinationRuleClient, v1alpha3.VirtualServiceClient, error) {
	kubeCache := kube.NewKubeCache()
	restConfig, err := kubeutils.GetConfig("", "")
	if err != nil {
		return nil, nil, err
	}
	drClient, err := v1alpha3.NewDestinationRuleClient(&factory.KubeResourceClientFactory{
		Crd:         v1alpha3.DestinationRuleCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, nil, err
	}
	if err := drClient.Register(); err != nil {
		return nil, nil, err
	}

	vsClient, err := v1alpha3.NewVirtualServiceClient(&factory.KubeResourceClientFactory{
		Crd:         v1alpha3.VirtualServiceCrd,
		Cfg:         restConfig,
		SharedCache: kubeCache,
	})
	if err != nil {
		return nil, nil, err
	}
	if err := vsClient.Register(); err != nil {
		return nil, nil, err
	}

	return drClient, vsClient, nil
}
