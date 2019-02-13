package appmesh_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/kubernetes"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/pkg2/api/external/istio/networking/v1alpha3"
	v1 "github.com/solo-io/supergloo/pkg2/api/v1"
	"github.com/solo-io/supergloo/test/utils"

	. "github.com/solo-io/supergloo/pkg2/translator/appmesh"
)

var _ = Describe("Appmesh", func() {
	It("works", func() {
		namespace := "placeholder-ns"
		secret, mesh := utils.MakeAppMeshResources(namespace)

		s := NewSyncer()
		err := s.Sync(context.TODO(), &v1.TranslatorSnapshot{
			Secrets: map[string]gloov1.SecretList{
				"": {secret},
			},
			Meshes: map[string]v1.MeshList{
				"ignored-at-this-point": {mesh},
			},
			Upstreams: map[string]gloov1.UpstreamList{
				"also gets ignored": {
					{
						Metadata: core.Metadata{
							Name:      "default-reviews-9080",
							Namespace: "gloo-system",
						},
						UpstreamSpec: &gloov1.UpstreamSpec{
							UpstreamType: &gloov1.UpstreamSpec_Kube{
								Kube: &kubernetes.UpstreamSpec{
									ServiceName:      "reviews",
									ServiceNamespace: "default",
									ServicePort:      9080,
									Selector:         map[string]string{"app": "reviews"},
								},
							},
						},
					},
					{
						Metadata: core.Metadata{
							Name:      "default-reviews-9080-version-v2",
							Namespace: namespace,
						},
						UpstreamSpec: &gloov1.UpstreamSpec{
							UpstreamType: &gloov1.UpstreamSpec_Kube{
								Kube: &kubernetes.UpstreamSpec{
									ServiceName:      "reviews",
									ServiceNamespace: "default",
									ServicePort:      9080,
									Selector:         map[string]string{"app": "reviews", "version": "v2"},
								},
							},
						},
					},
				},
			},
			Routingrules: map[string]v1.RoutingRuleList{
				"": {
					{
						Metadata:   core.Metadata{Name: "fault", Namespace: namespace},
						TargetMesh: &core.ResourceRef{Name: "name", Namespace: namespace},
						FaultInjection: &v1alpha3.HTTPFaultInjection{
							Abort: &v1alpha3.HTTPFaultInjection_Abort{
								ErrorType: &v1alpha3.HTTPFaultInjection_Abort_HttpStatus{
									HttpStatus: 566,
								},
								Percent: 100,
							},
						},
					},
					{
						Metadata:   core.Metadata{Name: "trafficshifting", Namespace: namespace},
						TargetMesh: &core.ResourceRef{Name: "name", Namespace: namespace},
						TrafficShifting: &v1.TrafficShifting{
							Destinations: []*v1.WeightedDestination{
								{
									Upstream: &core.ResourceRef{
										Name:      "default-reviews-9080-version-v2",
										Namespace: namespace,
									},
									Weight: 100,
								},
							},
						},
					},
				},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(err).NotTo(HaveOccurred())

	})
})
