package appmesh

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v12 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/meshdiscovery/clientset"
	"github.com/solo-io/supergloo/test/inputs/appmesh"
	"github.com/solo-io/supergloo/test/inputs/appmesh/scenarios"
)

var _ = Describe("appmesh config syncer", func() {
	var (
		cs     *clientset.Clientset
		ctx    context.Context
		syncer *appmeshDiscoveryConfigSyncer
		input  appmesh.TestResourceSet
	)

	BeforeEach(func() {
		var err error
		ctx = context.TODO()
		cs, err = clientset.ClientsetFromContext(ctx)
		Expect(err).NotTo(HaveOccurred())
		syncer = newAppmeshDiscoveryConfigSyncer(cs)
		input = scenarios.GetAllResources()
	})

	It("returns nil with 0 meshes", func() {
		snap := &v1.AppmeshDiscoverySnapshot{}
		Expect(syncer.Sync(ctx, snap)).NotTo(HaveOccurred())
	})

	It("works with multiple meshes", func() {

	})

	It("properly grabs config and applies it", func() {
		config, ok := input["other"]
		Expect(ok).To(BeTrue())
		snap := &v1.AppmeshDiscoverySnapshot{
			Pods: kubernetes.PodsByNamespace{
				"": config.MustGetPodList(),
			},
			Upstreams: v12.UpstreamsByNamespace{
				"": config.MustGetUpstreamList(),
			},
		}
		err := syncer.Sync(ctx, snap)
		Expect(err).NotTo(HaveOccurred())

	})

})
