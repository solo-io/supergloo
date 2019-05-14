package common_test

import (
	"context"

	kubev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg/stats/istio"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/prometheus/config"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/memory"
	v1 "github.com/solo-io/supergloo/pkg/api/external/prometheus/v1"
	sgv1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/solo-io/supergloo/pkg/stats/common"
)

var _ = Describe("PrometheusSyncer", func() {
	var client v1.PrometheusConfigClient
	var kube kubernetes.Interface
	BeforeEach(func() {
		var err error
		client, err = v1.NewPrometheusConfigClient(&factory.MemoryResourceClientFactory{Cache: memory.NewInMemoryResourceCache()})
		Expect(err).NotTo(HaveOccurred())
		kube = fake.NewSimpleClientset()
	})
	Context("when mesh is chosen", func() {
		It("adds scrape configs to that mesh's target prometheus and bounces any pods using those configs", func() {
			s := NewPrometheusSyncer(
				"my-prometheus-syncer",
				client,
				kube,
				func(mesh *sgv1.Mesh) bool {
					return true
				},
				func(mesh *sgv1.Mesh) (configs []*config.ScrapeConfig, e error) {
					return inputs.InputIstioPrometheusScrapeConfigs(), nil
				},
				false,
			)

			cfg1 := inputs.PrometheusConfig("cfg1", "observability-ns-1")
			cfg2 := inputs.PrometheusConfig("cfg2", "observability-ns-2")

			_, err := client.Write(cfg1, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())
			_, err = client.Write(cfg2, clients.WriteOpts{})
			Expect(err).NotTo(HaveOccurred())

			pod1, err := kube.CoreV1().Pods(cfg1.Metadata.Namespace).Create(prometheusPod("pod1", cfg1.Metadata.Ref()))
			Expect(err).NotTo(HaveOccurred())
			pod2, err := kube.CoreV1().Pods(cfg2.Metadata.Namespace).Create(prometheusPod("pod2", cfg2.Metadata.Ref()))
			Expect(err).NotTo(HaveOccurred())
			pod3, err := kube.CoreV1().Pods("ns").Create(prometheusPod("pod3", core.ResourceRef{"non-updated", "ns"}))
			Expect(err).NotTo(HaveOccurred())

			ns1, ns2, ns3 := "ns1", "ns2", "ns3"
			onlyCfg1 := []core.ResourceRef{cfg1.Metadata.Ref()}
			bothCfgs := []core.ResourceRef{cfg1.Metadata.Ref(), cfg2.Metadata.Ref()}

			// it merges together scrape configs for same prefix + namespace
			mesh1, mesh2, mesh3 := inputs.IstioMeshWithInstallNsPrometheus(ns1, ns1, nil, bothCfgs),
				inputs.IstioMeshWithInstallNsPrometheus(ns2, ns2, nil, bothCfgs),
				inputs.IstioMeshWithInstallNsPrometheus(ns3, ns3, nil, onlyCfg1)
			err = s.Sync(context.TODO(), &sgv1.RegistrationSnapshot{
				Meshes: sgv1.MeshList{mesh1, mesh2, mesh3},
			})
			Expect(err).NotTo(HaveOccurred())

			cfg1Updated, err := client.Read(cfg1.Metadata.Namespace, cfg1.Metadata.Name, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())
			cfg2Updated, err := client.Read(cfg2.Metadata.Namespace, cfg2.Metadata.Name, clients.ReadOpts{})
			Expect(err).NotTo(HaveOccurred())

			promCfg1 := cfg1Updated.Config
			promCfg2 := cfg2Updated.Config

			scsIstio1, _ := istio.PrometheusScrapeConfigs(ns1)
			scsIstio2, _ := istio.PrometheusScrapeConfigs(ns2)
			scsIstio3, _ := istio.PrometheusScrapeConfigs(ns3)
			for _, sc := range scsIstio1 {
				for _, prefix := range []string{"supergloo-my-prometheus-syncer-ns1.fancy-istio-"} {
					Expect(ScContainsJob(promCfg1.ScrapeConfigs, prefix+sc.JobName)).To(BeTrue())
					Expect(ScContainsJob(promCfg2.ScrapeConfigs, prefix+sc.JobName)).To(BeTrue())
				}
			}
			for _, sc := range scsIstio2 {
				for _, prefix := range []string{"supergloo-my-prometheus-syncer-ns2.fancy-istio-"} {
					Expect(ScContainsJob(promCfg1.ScrapeConfigs, prefix+sc.JobName)).To(BeTrue())
					Expect(ScContainsJob(promCfg2.ScrapeConfigs, prefix+sc.JobName)).To(BeTrue())
				}
			}
			for _, sc := range scsIstio3 {
				for _, prefix := range []string{"supergloo-my-prometheus-syncer-ns3.fancy-istio-"} {
					Expect(ScContainsJob(promCfg1.ScrapeConfigs, prefix+sc.JobName)).To(BeTrue())
					Expect(ScContainsJob(promCfg2.ScrapeConfigs, prefix+sc.JobName)).To(BeFalse())
				}
			}

			_, err = kube.CoreV1().Pods(pod1.Namespace).Get(pod1.Name, metav1.GetOptions{})
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())

			_, err = kube.CoreV1().Pods(pod2.Namespace).Get(pod2.Name, metav1.GetOptions{})
			Expect(err).To(HaveOccurred())
			Expect(errors.IsNotFound(err)).To(BeTrue())

			_, err = kube.CoreV1().Pods(pod3.Namespace).Get(pod3.Name, metav1.GetOptions{})
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func ScContainsJob(scs []*config.ScrapeConfig, jobName string) bool {
	for _, sc := range scs {
		if sc.JobName == jobName {
			return true
		}
	}
	return false
}

func prometheusPod(podName string, cfgRef core.ResourceRef) *kubev1.Pod {
	return &kubev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cfgRef.Namespace,
			Name:      podName,
		},
		Spec: kubev1.PodSpec{
			Volumes: []kubev1.Volume{
				{
					VolumeSource: kubev1.VolumeSource{
						ConfigMap: &kubev1.ConfigMapVolumeSource{
							LocalObjectReference: kubev1.LocalObjectReference{
								Name: cfgRef.Name,
							},
						},
					},
				},
			},
		},
	}
}
