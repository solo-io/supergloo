package utils_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/solo-io/supergloo/pkg/translator/utils"
)

var _ = Describe("UpstreamsForSelector", func() {
	Context("PodSelector_UpstreamSelector", func() {
		It("returns each kube upstream for each ref", func() {
			refs := []core.ResourceRef{
				{Name: "default-details-v1-9080", Namespace: "default"},
				{Name: "default-reviews-v2-9080", Namespace: "default"},
				{Name: "default-reviews-9080", Namespace: "default"},
			}
			upstreams, err := UpstreamsForSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: refs,
					},
				},
			}, inputs.BookInfoUpstreams("default"))
			Expect(err).NotTo(HaveOccurred())
			Expect(upstreams).To(HaveLen(3))
			for _, ref := range refs {
				us, err := upstreams.Find(ref.Strings())
				Expect(err).NotTo(HaveOccurred())
				Expect(us).NotTo(BeNil())
			}

		})
	})
	Context("PodSelector_LabelSelector", func() {
		It("returns each kube upstream with matching labels", func() {
			refs := []core.ResourceRef{
				{Name: "default-details-9080", Namespace: "default"},
				{Name: "default-details-v1-9080", Namespace: "default"},
			}
			upstreams, err := UpstreamsForSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{"app": "details"},
					},
				},
			}, inputs.BookInfoUpstreams("default"))
			Expect(err).NotTo(HaveOccurred())
			Expect(upstreams).To(HaveLen(2))
			for _, ref := range refs {
				us, err := upstreams.Find(ref.Strings())
				Expect(err).NotTo(HaveOccurred())
				Expect(us).NotTo(BeNil())
			}

		})
	})
	Context("PodSelector_NamespaceSelector_", func() {
		It("returns each kube upstream in the namespaces", func() {
			refs1 := []core.ResourceRef{
				{Name: "default-details-v1-9080", Namespace: "default1"},
				{Name: "default-reviews-v2-9080", Namespace: "default1"},
				{Name: "default-reviews-9080", Namespace: "default1"},
			}
			refs2 := []core.ResourceRef{
				{Name: "default-details-v1-9080", Namespace: "default2"},
				{Name: "default-reviews-v2-9080", Namespace: "default2"},
				{Name: "default-reviews-9080", Namespace: "default2"},
			}
			refs := append(refs1, refs2...)
			allUs := inputs.BookInfoUpstreams("default1")
			allUs = append(allUs, inputs.BookInfoUpstreams("default2")...)
			upstreams, err := UpstreamsForSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_NamespaceSelector_{
					NamespaceSelector: &v1.PodSelector_NamespaceSelector{
						Namespaces: []string{"default1", "default2"},
					},
				},
			}, allUs)
			Expect(err).NotTo(HaveOccurred())
			Expect(upstreams).To(HaveLen(len(allUs)))
			for _, ref := range refs {
				us, err := upstreams.Find(ref.Strings())
				Expect(err).NotTo(HaveOccurred())
				Expect(us).NotTo(BeNil())
			}
			upstreams, err = UpstreamsForSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_NamespaceSelector_{
					NamespaceSelector: &v1.PodSelector_NamespaceSelector{
						Namespaces: []string{"default1"},
					},
				},
			}, allUs)
			Expect(err).NotTo(HaveOccurred())
			Expect(upstreams).To(HaveLen(len(allUs) / 2))
			for _, ref := range refs1 {
				us, err := upstreams.Find(ref.Strings())
				Expect(err).NotTo(HaveOccurred())
				Expect(us).NotTo(BeNil())
			}
			upstreams, err = UpstreamsForSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_NamespaceSelector_{
					NamespaceSelector: &v1.PodSelector_NamespaceSelector{
						Namespaces: []string{"default2"},
					},
				},
			}, allUs)
			Expect(err).NotTo(HaveOccurred())
			Expect(upstreams).To(HaveLen(len(allUs) / 2))
			for _, ref := range refs2 {
				us, err := upstreams.Find(ref.Strings())
				Expect(err).NotTo(HaveOccurred())
				Expect(us).NotTo(BeNil())
			}

		})
	})
})

var _ = Describe("PodsForSelector", func() {
	Context("PodSelector_UpstreamSelector", func() {
		It("returns the pod for specified upstream refs", func() {
			refs := []core.ResourceRef{
				{Name: "default-reviews-9080", Namespace: "default"},
			}
			pods := inputs.BookInfoPods("default")
			selectedPods, err := PodsForSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_UpstreamSelector_{
					UpstreamSelector: &v1.PodSelector_UpstreamSelector{
						Upstreams: refs,
					},
				},
			}, inputs.BookInfoUpstreams("default"),
				pods)
			Expect(err).NotTo(HaveOccurred())
			Expect(selectedPods).To(HaveLen(3))
			for _, ref := range []core.ResourceRef{
				{Namespace: "default", Name: "reviews-v1-pod-1"},
				{Namespace: "default", Name: "reviews-v2-pod-1"},
				{Namespace: "default", Name: "reviews-v3-pod-1"},
			} {
				pod, err := pods.Find(ref.Strings())
				Expect(err).NotTo(HaveOccurred())
				Expect(pod).NotTo(BeNil())
			}
		})
	})
	Context("PodSelector_ServiceSelector", func() {
		It("returns the pod for specified service refs", func() {
			refs := []core.ResourceRef{
				{Name: "reviews", Namespace: "default"},
				{Name: "details", Namespace: "default"},
			}
			svcs := inputs.BookInfoServices("default")
			selectedPods, err := ServicesForSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_ServiceSelector_{
					ServiceSelector: &v1.PodSelector_ServiceSelector{
						Services: refs,
					},
				},
			}, inputs.BookInfoUpstreams("default"),
				svcs)
			Expect(err).NotTo(HaveOccurred())
			Expect(selectedPods).To(HaveLen(2))
			for _, ref := range []core.ResourceRef{
				{Namespace: "default", Name: "reviews"},
				{Namespace: "default", Name: "details"},
			} {
				pod, err := svcs.Find(ref.Strings())
				Expect(err).NotTo(HaveOccurred())
				Expect(pod).NotTo(BeNil())
			}
		})
	})
	Context("PodSelector_LabelSelector", func() {
		It("returns each pod with matching labels", func() {
			pods := inputs.BookInfoPods("default")
			selectedPods, err := PodsForSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_LabelSelector_{
					LabelSelector: &v1.PodSelector_LabelSelector{
						LabelsToMatch: map[string]string{"app": "reviews"},
					},
				},
			}, inputs.BookInfoUpstreams("default"),
				pods)
			Expect(err).NotTo(HaveOccurred())
			Expect(selectedPods).To(HaveLen(3))
			for _, ref := range []core.ResourceRef{
				{Namespace: "default", Name: "reviews-v1-pod-1"},
				{Namespace: "default", Name: "reviews-v2-pod-1"},
				{Namespace: "default", Name: "reviews-v3-pod-1"},
			} {
				pod, err := pods.Find(ref.Strings())
				Expect(err).NotTo(HaveOccurred())
				Expect(pod).NotTo(BeNil())
			}
		})
	})
	Context("PodSelector_NamespaceSelector_", func() {
		It("returns each pod in the namespaces", func() {
			pods := inputs.BookInfoPods("default1")
			pods = append(pods, inputs.BookInfoPods("default2")...)
			selectedPods, err := PodsForSelector(&v1.PodSelector{
				SelectorType: &v1.PodSelector_NamespaceSelector_{
					NamespaceSelector: &v1.PodSelector_NamespaceSelector{
						Namespaces: []string{"default1", "default2"},
					},
				},
			}, nil, pods)
			Expect(err).NotTo(HaveOccurred())
			Expect(selectedPods).To(HaveLen(len(pods)))
		})
	})
})
