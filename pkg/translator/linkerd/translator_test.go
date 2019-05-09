package linkerd_test

import (
	"context"

	"github.com/solo-io/supergloo/api/external/linkerd"
	linkerdv1 "github.com/solo-io/supergloo/pkg/api/external/linkerd/v1"

	"github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/translator/linkerd/plugins"
	"github.com/solo-io/supergloo/test/inputs"

	. "github.com/solo-io/supergloo/pkg/translator/linkerd"
)

var _ = Describe("Translator", func() {
	It("translates routing rules to service profiles for the given upstreams", func() {
		t := NewTranslator(plugins.Plugins(nil))
		bookinfoNs := "bookinfo-installed-here"
		inputMesh := inputs.LinkerdMesh("linkerd-installed-here", nil)
		inputMeshRef := inputMesh.Metadata.Ref()
		meshConfig, resourceErrs, err := t.Translate(context.TODO(), &v1.ConfigSnapshot{
			Meshes:       v1.MeshList{inputMesh},
			Upstreams:    inputs.BookInfoUpstreams(bookinfoNs),
			Routingrules: inputs.AdvancedBookInfoRoutingRules(bookinfoNs, &inputMeshRef),
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(resourceErrs.Validate()).NotTo(HaveOccurred())
		Expect(meshConfig).To(HaveLen(1))
		Expect(meshConfig).To(HaveKey(inputMesh))
		cfg := meshConfig[inputMesh]
		Expect(cfg.ServiceProfiles).To(HaveLen(2))
		Expect(cfg.ServiceProfiles[0]).To(Equal(&linkerdv1.ServiceProfile{
			ServiceProfile: linkerd.ServiceProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ratings.bookinfo-installed-here.svc.cluster.local",
					Namespace: "bookinfo-installed-here",
				},
				Spec: v1alpha1.ServiceProfileSpec{
					Routes: []*v1alpha1.RouteSpec{{
						Name:        "default",
						Condition:   &v1alpha1.RequestMatch{PathRegex: ".*"},
						IsRetryable: true,
					}},
					RetryBudget: &v1alpha1.RetryBudget{
						RetryRatio:          0,
						MinRetriesPerSecond: 5,
						TTL:                 "0s",
					},
				},
			},
		}))
		Expect(cfg.ServiceProfiles[1]).To(Equal(&linkerdv1.ServiceProfile{
			ServiceProfile: linkerd.ServiceProfile{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "reviews.bookinfo-installed-here.svc.cluster.local",
					Namespace: "bookinfo-installed-here",
				},
				Spec: v1alpha1.ServiceProfileSpec{
					Routes: []*v1alpha1.RouteSpec{
						{
							Name: "GET,POST /users/.*",
							Condition: &v1alpha1.RequestMatch{
								Any: []*v1alpha1.RequestMatch{
									{
										PathRegex: "/users/.*",
										Method:    "GET",
									},
									{
										PathRegex: "/users/.*",
										Method:    "POST",
									},
								},
							},
							IsRetryable: true,
						},
						{
							Name: "GET,POST /posts/.*",
							Condition: &v1alpha1.RequestMatch{
								Any: []*v1alpha1.RequestMatch{
									{
										PathRegex: "/posts/.*",
										Method:    "GET",
									},
									{
										PathRegex: "/posts/.*",
										Method:    "POST",
									},
								},
							},
							ResponseClasses: nil,
							IsRetryable:     true,
						}},
					RetryBudget: &v1alpha1.RetryBudget{
						RetryRatio:          0,
						MinRetriesPerSecond: 5,
						TTL:                 "0s",
					},
				},
			},
		}))
	})
})
