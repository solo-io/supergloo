package dns_test

import (
	"context"
	"encoding/json"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/clients"
	mock_kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core/mocks"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	"github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/federation/dns"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Federation Decider", func() {
	var (
		ctrl *gomock.Controller
		ctx  context.Context

		clusterName1 = "cluster-1"
		clusterName2 = "cluster-2"
		clusterName3 = "cluster-3"
		cmRef        = &core_types.ResourceRef{
			Name:      dns.IpRecordName,
			Namespace: env.DefaultWriteNamespace,
		}
		mustMarshal = func(obj interface{}) string {
			bytes, err := json.Marshal(obj)
			Expect(err).NotTo(HaveOccurred(), "Unexpected error while marshaling in test")
			return string(bytes)
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		ctx = context.TODO()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("does not assign an IP with all zeroes in the host component to a brand new cluster", func() {
		// the IP 240.0.0.0 is invalid since its host component (the bottom 28 bits, based on the /4 CIDR) are all zero,
		// so that one should not be issued first
		expectedIp := "240.0.0.1"

		cmClient := mock_kubernetes_core.NewMockConfigMapClient(ctrl)
		cmClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(cmRef)).
			Return(&corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
			}, nil)
		cmClient.EXPECT().
			Update(ctx, &corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
				Data: map[string]string{
					clusterName1: mustMarshal([]string{expectedIp}),
				},
			})

		ipAssigner := dns.NewIpAssigner(cmClient)

		newIp, err := ipAssigner.AssignIPOnCluster(ctx, clusterName1)
		Expect(err).NotTo(HaveOccurred())
		Expect(newIp).To(Equal(expectedIp))
	})

	It("can assign an IP to a cluster that has already had IPs generated for it", func() {
		expectedIp := "240.0.0.3"

		cmClient := mock_kubernetes_core.NewMockConfigMapClient(ctrl)
		cmClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(cmRef)).
			Return(&corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
				Data: map[string]string{
					clusterName1: mustMarshal([]string{"240.0.0.1", "240.0.0.2"}),
				},
			}, nil)
		cmClient.EXPECT().
			Update(ctx, &corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
				Data: map[string]string{
					clusterName1: mustMarshal([]string{"240.0.0.1", "240.0.0.2", expectedIp}),
				},
			})

		ipAssigner := dns.NewIpAssigner(cmClient)

		newIp, err := ipAssigner.AssignIPOnCluster(ctx, clusterName1)
		Expect(err).NotTo(HaveOccurred())
		Expect(newIp).To(Equal(expectedIp))
	})

	It("can un-assign issued IPs", func() {
		cmClient := mock_kubernetes_core.NewMockConfigMapClient(ctrl)
		cmClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(cmRef)).
			Return(&corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
				Data: map[string]string{
					clusterName1: mustMarshal([]string{"240.0.0.1", "240.0.0.2", "240.0.0.3"}),
					clusterName2: mustMarshal([]string{"a.b.c.d"}),
				},
			}, nil)

		firstIpRemovedCM := &corev1.ConfigMap{
			ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
			Data: map[string]string{
				clusterName1: mustMarshal([]string{"240.0.0.1", "", "240.0.0.3"}),
				clusterName2: mustMarshal([]string{"a.b.c.d"}), // should be unchanged
			},
		}
		cmClient.EXPECT().
			Update(ctx, firstIpRemovedCM).
			Return(nil)

		cmClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(cmRef)).
			Return(firstIpRemovedCM, nil)

		cmClient.EXPECT().
			Update(ctx, &corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
				Data: map[string]string{
					clusterName1: mustMarshal([]string{"240.0.0.1", "", ""}),
					clusterName2: mustMarshal([]string{"a.b.c.d"}), // should be unchanged
				},
			}).
			Return(nil)

		ipAssigner := dns.NewIpAssigner(cmClient)
		err := ipAssigner.UnAssignIPOnCluster(ctx, clusterName1, "240.0.0.2")
		Expect(err).NotTo(HaveOccurred())
		err = ipAssigner.UnAssignIPOnCluster(ctx, clusterName1, "240.0.0.3")
		Expect(err).NotTo(HaveOccurred())
	})

	It("can re-use un-assigned IPs", func() {
		cmClient := mock_kubernetes_core.NewMockConfigMapClient(ctrl)
		cmClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(cmRef)).
			Return(&corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
				Data: map[string]string{
					clusterName1: mustMarshal([]string{"240.0.0.1", "", "240.0.0.3"}),
					clusterName2: mustMarshal([]string{"", "240.0.0.2", "240.0.0.3"}),
					clusterName3: mustMarshal([]string{"240.0.0.1", "240.0.0.2", ""}),
				},
			}, nil)

		cmClient.EXPECT().
			Update(ctx, &corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
				Data: map[string]string{
					clusterName1: mustMarshal([]string{"240.0.0.1", "240.0.0.2", "240.0.0.3"}),
					clusterName2: mustMarshal([]string{"", "240.0.0.2", "240.0.0.3"}),
					clusterName3: mustMarshal([]string{"240.0.0.1", "240.0.0.2", ""}),
				},
			}).
			Return(nil)

		ipAssigner := dns.NewIpAssigner(cmClient)

		newIp, err := ipAssigner.AssignIPOnCluster(ctx, clusterName1)
		Expect(err).NotTo(HaveOccurred())
		Expect(newIp).To(Equal("240.0.0.2"))

		cmClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(cmRef)).
			Return(&corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
				Data: map[string]string{
					clusterName1: mustMarshal([]string{"240.0.0.1", "240.0.0.2", "240.0.0.3"}),
					clusterName2: mustMarshal([]string{"", "240.0.0.2", "240.0.0.3"}),
					clusterName3: mustMarshal([]string{"240.0.0.1", "240.0.0.2", ""}),
				},
			}, nil)

		cmClient.EXPECT().
			Update(ctx, &corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
				Data: map[string]string{
					clusterName1: mustMarshal([]string{"240.0.0.1", "240.0.0.2", "240.0.0.3"}),
					clusterName2: mustMarshal([]string{"240.0.0.1", "240.0.0.2", "240.0.0.3"}),
					clusterName3: mustMarshal([]string{"240.0.0.1", "240.0.0.2", ""}),
				},
			}).
			Return(nil)

		newIp, err = ipAssigner.AssignIPOnCluster(ctx, clusterName2)
		Expect(err).NotTo(HaveOccurred())
		Expect(newIp).To(Equal("240.0.0.1"))

		cmClient.EXPECT().
			Get(ctx, clients.ResourceRefToObjectKey(cmRef)).
			Return(&corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
				Data: map[string]string{
					clusterName1: mustMarshal([]string{"240.0.0.1", "240.0.0.2", "240.0.0.3"}),
					clusterName2: mustMarshal([]string{"240.0.0.1", "240.0.0.2", "240.0.0.3"}),
					clusterName3: mustMarshal([]string{"240.0.0.1", "240.0.0.2", ""}),
				},
			}, nil)

		cmClient.EXPECT().
			Update(ctx, &corev1.ConfigMap{
				ObjectMeta: clients.ResourceRefToObjectMeta(cmRef),
				Data: map[string]string{
					clusterName1: mustMarshal([]string{"240.0.0.1", "240.0.0.2", "240.0.0.3"}),
					clusterName2: mustMarshal([]string{"240.0.0.1", "240.0.0.2", "240.0.0.3"}),
					clusterName3: mustMarshal([]string{"240.0.0.1", "240.0.0.2", "240.0.0.3"}),
				},
			}).
			Return(nil)

		newIp, err = ipAssigner.AssignIPOnCluster(ctx, clusterName3)
		Expect(err).NotTo(HaveOccurred())
		Expect(newIp).To(Equal("240.0.0.3"))
	})
})
