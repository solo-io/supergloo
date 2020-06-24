package validation_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover/validation"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Validation", func() {
	var (
		ctrl      *gomock.Controller
		validator validation.FailoverServiceValidator
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		validator = validation.NewFailoverServiceValidator()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	var meshServiceStatusWithOutlierDetection = func() smh_discovery_types.MeshServiceStatus {
		return smh_discovery_types.MeshServiceStatus{
			ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
				{
					TrafficPolicySpec: &smh_networking_types.TrafficPolicySpec{
						OutlierDetection: &smh_networking_types.TrafficPolicySpec_OutlierDetection{},
					},
				},
			},
		}
	}

	// Snapshot with valid FailoverService.
	var validInputSnapshot = func() failover.InputSnapshot {
		return failover.InputSnapshot{
			FailoverServices: []*smh_networking.FailoverService{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Generation: 1},
					Spec: smh_networking_types.FailoverServiceSpec{
						Hostname:  "valid.dns.hostname",
						Namespace: "namespace",
						Port: &smh_networking_types.FailoverServiceSpec_Port{
							Port:     9080,
							Name:     "valid-name",
							Protocol: "TCP",
						},
						Cluster: "cluster1",
						Services: []*smh_core_types.ResourceRef{
							{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster1",
							},
							{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster2",
							},
						},
					},
				},
			},
			MeshServices: []*smh_discovery.MeshService{
				{
					Spec: smh_discovery_types.MeshServiceSpec{
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster1",
							},
						},
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh1",
							Namespace: "namespace1",
							Cluster:   "cluster1",
						},
					},
					Status: meshServiceStatusWithOutlierDetection(),
				},
				{
					Spec: smh_discovery_types.MeshServiceSpec{
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster2",
							},
						},
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh2",
							Namespace: "namespace1",
							Cluster:   "cluster2",
						},
					},
					Status: meshServiceStatusWithOutlierDetection(),
				},
			},
			KubeClusters: []*smh_discovery.KubernetesCluster{
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "cluster1"}},
				{ObjectMeta: k8s_meta_types.ObjectMeta{Name: "cluster2"}},
			},
			Meshes: []*smh_discovery.Mesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh1",
						Namespace: "namespace1",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
						Cluster:  &smh_core_types.ResourceRef{Name: "cluster1"},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh2",
						Namespace: "namespace1",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
						Cluster:  &smh_core_types.ResourceRef{Name: "cluster2"},
					},
				},
			},
			VirtualMeshes: []*smh_networking.VirtualMesh{
				{
					Spec: smh_networking_types.VirtualMeshSpec{
						Meshes: []*smh_core_types.ResourceRef{
							{
								Name:      "mesh1",
								Namespace: "namespace1",
							},
							{
								Name:      "mesh2",
								Namespace: "namespace1",
							},
						},
					},
				},
			},
		}
	}

	// Snapshot with invalid FailoverService.
	var invalidInputSnapshot = func() failover.InputSnapshot {
		return failover.InputSnapshot{
			FailoverServices: []*smh_networking.FailoverService{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{Generation: 1},
					Spec: smh_networking_types.FailoverServiceSpec{
						Hostname: "!@#$",
						Cluster:  "cluster1",
						Services: []*smh_core_types.ResourceRef{
							{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster1",
							},
							{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster2",
							},
							{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster3",
							},
							{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster4",
							},
						},
					},
				},
			},
			MeshServices: []*smh_discovery.MeshService{
				{
					Spec: smh_discovery_types.MeshServiceSpec{
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster1",
							},
						},
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh1",
							Namespace: "namespace1",
							Cluster:   "cluster1",
						},
					},
					Status: meshServiceStatusWithOutlierDetection(),
				},
				{
					Spec: smh_discovery_types.MeshServiceSpec{
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster2",
							},
						},
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh2",
							Namespace: "namespace1",
							Cluster:   "cluster2",
						},
					},
					Status: meshServiceStatusWithOutlierDetection(),
				},
				{
					Spec: smh_discovery_types.MeshServiceSpec{
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster3",
							},
						},
						Mesh: &smh_core_types.ResourceRef{
							Name:      "mesh3",
							Namespace: "namespace1",
							Cluster:   "cluster3",
						},
					},
					Status: meshServiceStatusWithOutlierDetection(),
				},
				{
					Spec: smh_discovery_types.MeshServiceSpec{
						KubeService: &smh_discovery_types.MeshServiceSpec_KubeService{
							Ref: &smh_core_types.ResourceRef{
								Name:      "service1",
								Namespace: "namespace1",
								Cluster:   "cluster4",
							},
						},
					},
				},
			},
			KubeClusters: []*smh_discovery.KubernetesCluster{},
			Meshes: []*smh_discovery.Mesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh1",
						Namespace: "namespace1",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
						Cluster:  &smh_core_types.ResourceRef{Name: "cluster1"},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh2",
						Namespace: "namespace1",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
						Cluster:  &smh_core_types.ResourceRef{Name: "cluster2"},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name:      "mesh3",
						Namespace: "namespace1",
					},
					Spec: smh_discovery_types.MeshSpec{
						MeshType: &smh_discovery_types.MeshSpec_Istio1_5_{},
						Cluster:  &smh_core_types.ResourceRef{Name: "cluster3"},
					},
				},
			},
			VirtualMeshes: []*smh_networking.VirtualMesh{
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "vm1",
					},
					Spec: smh_networking_types.VirtualMeshSpec{
						Meshes: []*smh_core_types.ResourceRef{
							{
								Name:      "mesh1",
								Namespace: "namespace1",
							},
						},
					},
				},
				{
					ObjectMeta: k8s_meta_types.ObjectMeta{
						Name: "vm2",
					},
					Spec: smh_networking_types.VirtualMeshSpec{
						Meshes: []*smh_core_types.ResourceRef{
							{
								Name:      "mesh2",
								Namespace: "namespace1",
							},
						},
					},
				},
			},
		}
	}

	It("should set validation status on valid FailoverService", func() {
		inputSnapshot := validInputSnapshot()
		validator.Validate(inputSnapshot)
		expectedFailoverServiceStatus := smh_networking_types.FailoverServiceStatus{
			ObservedGeneration: inputSnapshot.FailoverServices[0].GetGeneration(),
			ValidationStatus: &smh_core_types.Status{
				State: smh_core_types.Status_ACCEPTED,
			},
		}
		Expect(inputSnapshot.FailoverServices[0].Status).To(Equal(expectedFailoverServiceStatus))
	})

	It("should set validation status on invalid FailoverService", func() {
		inputSnapshot := invalidInputSnapshot()
		validator.Validate(inputSnapshot)
		failoverService := inputSnapshot.FailoverServices[0]
		actualStatus := failoverService.Status
		Expect(actualStatus.ObservedGeneration).To(Equal(failoverService.GetGeneration()))
		Expect(actualStatus.GetValidationStatus().GetState()).To(Equal(smh_core_types.Status_INVALID))
		// Invalid hostname
		Expect(actualStatus.GetValidationStatus().GetMessage()).To(ContainSubstring("DNS-1123 subdomain must consist of"))
		// Cluster not found
		Expect(actualStatus.GetValidationStatus().GetMessage()).To(ContainSubstring(validation.ClusterNotFound(failoverService.Spec.GetCluster()).Error()))
		// Missing namespace
		Expect(actualStatus.GetValidationStatus().GetMessage()).To(ContainSubstring(validation.MissingNamespace.Error()))
		// Missing port
		Expect(actualStatus.GetValidationStatus().GetMessage()).To(ContainSubstring(validation.MissingPort.Error()))
		// Service without OutlierDetection
		Expect(actualStatus.GetValidationStatus().GetMessage()).To(ContainSubstring(validation.MissingOutlierDetection(inputSnapshot.MeshServices[3]).Error()))
		// Mesh without parent VirtualMesh
		Expect(actualStatus.GetValidationStatus().GetMessage()).To(ContainSubstring(
			validation.ServiceWithoutParentVM(
				inputSnapshot.MeshServices[2].Spec.GetKubeService().GetRef(),
				inputSnapshot.Meshes[2],
			).Error()))
		Expect(actualStatus.GetValidationStatus().GetMessage()).To(ContainSubstring(validation.MultipleParentVirtualMeshes(inputSnapshot.VirtualMeshes).Error()))
	})
})
