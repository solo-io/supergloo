package apply_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	certificatesv1 "github.com/solo-io/gloo-mesh/pkg/api/certificates.mesh.gloo.solo.io/v1"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	discoveryv1 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/reporting"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	skv2corev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/solo-io/skv2/test/matchers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/solo-io/gloo-mesh/pkg/mesh-networking/apply"
)

var _ = Describe("Applier", func() {

	Context("applied traffic policies", func() {
		var (
			destination = &discoveryv1.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
				Spec: discoveryv1.DestinationSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "svc-name",
								Namespace:   "svc-namespace",
								ClusterName: "svc-cluster",
							},
						},
					},
				},
			}
			workload = &discoveryv1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wkld1",
					Namespace: "ns",
				},
				Spec: discoveryv1.WorkloadSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
				},
			}
			mesh = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh1",
					Namespace: "ns",
				},
			}
			trafficPolicy1 = &networkingv1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
				Spec: networkingv1.TrafficPolicySpec{
					Policy: &networkingv1.TrafficPolicySpec_Policy{
						// fill an arbitrary part of the spec
						Mirror: &networkingv1.TrafficPolicySpec_Policy_Mirror{},
					},
				},
			}
			trafficPolicy2 = &networkingv1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp2",
					Namespace: "ns",
				},
				Spec: networkingv1.TrafficPolicySpec{
					Policy: &networkingv1.TrafficPolicySpec_Policy{
						// fill an arbitrary part of the spec
						FaultInjection: &networkingv1.TrafficPolicySpec_Policy_FaultInjection{},
					},
				},
			}

			snap = input.NewInputLocalSnapshotManualBuilder("").
				AddDestinations(discoveryv1.DestinationSlice{destination}).
				AddTrafficPolicies(networkingv1.TrafficPolicySlice{trafficPolicy1, trafficPolicy2}).
				AddWorkloads(discoveryv1.WorkloadSlice{workload}).
				AddMeshes(discoveryv1.MeshSlice{mesh}).
				Build()
		)

		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap, nil, nil)
		})
		It("updates status on input traffic policies", func() {
			Expect(trafficPolicy1.Status.Destinations).To(HaveKey(sets.Key(destination)))
			Expect(trafficPolicy1.Status.Destinations[sets.Key(destination)]).To(Equal(&networkingv1.ApprovalStatus{
				AcceptanceOrder: 0,
				State:           commonv1.ApprovalState_ACCEPTED,
			}))
			Expect(trafficPolicy1.Status.Workloads).To(HaveLen(1))
			Expect(trafficPolicy1.Status.Workloads[0]).To(Equal(sets.Key(workload)))
			Expect(trafficPolicy2.Status.Destinations).To(HaveKey(sets.Key(destination)))
			Expect(trafficPolicy2.Status.Destinations[sets.Key(destination)]).To(Equal(&networkingv1.ApprovalStatus{
				AcceptanceOrder: 1,
				State:           commonv1.ApprovalState_ACCEPTED,
			}))
			Expect(trafficPolicy2.Status.Workloads).To(HaveLen(1))
			Expect(trafficPolicy2.Status.Workloads[0]).To(Equal(sets.Key(workload)))

		})
		It("updates status on input Destination policies", func() {
			Expect(destination.Status.AppliedTrafficPolicies).To(HaveLen(2))
			Expect(destination.Status.AppliedTrafficPolicies[0].Ref).To(Equal(ezkube.MakeObjectRef(trafficPolicy1)))
			Expect(destination.Status.AppliedTrafficPolicies[0].Spec).To(Equal(&trafficPolicy1.Spec))
			Expect(destination.Status.AppliedTrafficPolicies[1].Ref).To(Equal(ezkube.MakeObjectRef(trafficPolicy2)))
			Expect(destination.Status.AppliedTrafficPolicies[1].Spec).To(Equal(&trafficPolicy2.Spec))
			Expect(destination.Status.LocalFqdn).To(Equal("svc-name.svc-namespace.svc.cluster.local"))
		})
	})

	Context("VirtualMesh status", func() {

		It("retains the conditions on a Virtual Mesh", func() {

			virtualMesh := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm",
					Namespace: "test",
				},
				Status: networkingv1.VirtualMeshStatus{
					Conditions: []*certificatesv1.CertificateRotationCondition{
						{
							State: certificatesv1.CertificateRotationState_FAILED,
						},
						{
							State: certificatesv1.CertificateRotationState_ADDING_NEW_ROOT,
						},
					},
				},
			}

			snap := input.NewInputLocalSnapshotManualBuilder("").
				AddVirtualMeshes([]*networkingv1.VirtualMesh{virtualMesh}).
				Build()

			vmExpectCopy := virtualMesh.DeepCopy()
			vmExpectCopy.Status.State = commonv1.ApprovalState_ACCEPTED
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap, nil, nil)
			Expect(&vmExpectCopy.Status).To(matchers.MatchProto(&virtualMesh.Status))
		})
	})

	Context("invalid traffic policies", func() {
		var (
			destination = &discoveryv1.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
			}
			trafficPolicy = &networkingv1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
			}

			snap = input.NewInputLocalSnapshotManualBuilder("").
				AddDestinations(discoveryv1.DestinationSlice{destination}).
				AddTrafficPolicies(networkingv1.TrafficPolicySlice{trafficPolicy}).
				Build()
		)

		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// report = reject
				reporter.ReportTrafficPolicyToDestination(destination, trafficPolicy, errors.New("did an oopsie"))
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap, nil, nil)
		})
		It("updates status on input traffic policies", func() {
			Expect(trafficPolicy.Status.Destinations).To(HaveKey(sets.Key(destination)))
			Expect(trafficPolicy.Status.Destinations[sets.Key(destination)]).To(Equal(&networkingv1.ApprovalStatus{
				AcceptanceOrder: 0,
				State:           commonv1.ApprovalState_INVALID,
				Errors:          []string{"did an oopsie"},
			}))
		})
		It("does not add the policy to the Destination status", func() {
			Expect(destination.Status.AppliedTrafficPolicies).To(HaveLen(0))
		})
	})

	Context("setting workloads status", func() {
		var (
			destination1 = &discoveryv1.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
				Spec: discoveryv1.DestinationSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							WorkloadSelectorLabels: map[string]string{"istio": "ingressgateway"},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
								{
									Port: 1234,
									Name: defaults.IstioGatewayTlsPortName,
								},
							},
						},
					},
				},
			}
			destination2 = &discoveryv1.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms2",
					Namespace: "ns",
				},
				Spec: discoveryv1.DestinationSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh2",
						Namespace: "ns",
					},
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							WorkloadSelectorLabels: map[string]string{"istio": "ingressgateway"},
							Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
								{
									Port: 1234,
									Name: defaults.IstioGatewayTlsPortName,
								},
							},
						},
					},
				},
			}
			workload1 = &discoveryv1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wkld1",
					Namespace: "ns",
				},
				Spec: discoveryv1.WorkloadSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
				},
			}
			workload2 = &discoveryv1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wkld2",
					Namespace: "ns",
				},
				Spec: discoveryv1.WorkloadSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh2",
						Namespace: "ns",
					},
				},
			}
			mesh1 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh1",
					Namespace: "ns",
				},
			}
			mesh2 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh2",
					Namespace: "ns",
				},
			}
			virtualMesh = &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vmesh1",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						{Name: "mesh1", Namespace: "ns"},
						{Name: "mesh2", Namespace: "ns"},
					},
				},
			}
			trafficPolicy = &networkingv1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
			}
			accessPolicy = &networkingv1.AccessPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ap1",
					Namespace: "ns",
				},
			}
		)

		It("sets policy workloads using mesh", func() {
			snap := input.NewInputLocalSnapshotManualBuilder("").
				AddTrafficPolicies(networkingv1.TrafficPolicySlice{trafficPolicy}).
				AddAccessPolicies(networkingv1.AccessPolicySlice{accessPolicy}).
				AddDestinations(discoveryv1.DestinationSlice{destination1}).
				AddWorkloads(discoveryv1.WorkloadSlice{workload1, workload2}).
				AddMeshes(discoveryv1.MeshSlice{mesh1, mesh2}).
				Build()
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap, nil, nil)

			// destination1 and workload1 are both in mesh1
			Expect(trafficPolicy.Status.Workloads).To(HaveLen(1))
			Expect(trafficPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
			Expect(accessPolicy.Status.Workloads).To(HaveLen(1))
			Expect(accessPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
		})
		It("sets policy workloads using VirtualMesh", func() {
			snap := input.NewInputLocalSnapshotManualBuilder("").
				AddTrafficPolicies(networkingv1.TrafficPolicySlice{trafficPolicy}).
				AddAccessPolicies(networkingv1.AccessPolicySlice{accessPolicy}).
				AddDestinations(discoveryv1.DestinationSlice{destination1, destination2}).
				AddWorkloads(discoveryv1.WorkloadSlice{workload1, workload2}).
				AddMeshes(discoveryv1.MeshSlice{mesh1, mesh2}).
				AddVirtualMeshes(networkingv1.VirtualMeshSlice{virtualMesh}).
				Build()
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap, nil, nil)

			// destination1 is in mesh1, workload1 is in mesh1, and workload2 is in mesh2.
			// since mesh1 and mesh2 are in the same VirtualMesh, both workloads are returned
			Expect(trafficPolicy.Status.Workloads).To(HaveLen(2))
			Expect(trafficPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
			Expect(trafficPolicy.Status.Workloads[1]).To(Equal(sets.Key(workload2)))
			Expect(accessPolicy.Status.Workloads).To(HaveLen(2))
			Expect(accessPolicy.Status.Workloads[0]).To(Equal(sets.Key(workload1)))
			Expect(accessPolicy.Status.Workloads[1]).To(Equal(sets.Key(workload2)))
		})
		It("sets no policy workloads when there is no matching mesh", func() {
			workload1.Spec.Mesh.Name = "mesh2"
			snap := input.NewInputLocalSnapshotManualBuilder("").
				AddTrafficPolicies(networkingv1.TrafficPolicySlice{trafficPolicy}).
				AddAccessPolicies(networkingv1.AccessPolicySlice{accessPolicy}).
				AddDestinations(discoveryv1.DestinationSlice{destination1}).
				AddWorkloads(discoveryv1.WorkloadSlice{workload1, workload2}).
				AddMeshes(discoveryv1.MeshSlice{mesh1, mesh2}).
				Build()
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier := NewApplier(translator)
			applier.Apply(context.TODO(), snap, nil, nil)

			// destination1 is in mesh1, but both workloads are in mesh2
			Expect(trafficPolicy.Status.Workloads).To(BeNil())
			Expect(accessPolicy.Status.Workloads).To(BeNil())
		})
	})

	Context("applied federation", func() {
		var (
			applier Applier

			destination = &discoveryv1.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
				Spec: discoveryv1.DestinationSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "svc-name",
								Namespace:   "svc-namespace",
								ClusterName: "svc-cluster",
							},
						},
					},
				},
			}

			mesh1 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh1",
					Namespace: "ns",
				},
			}
			mesh2 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh2",
					Namespace: "ns",
				},
			}
			mesh3 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh3",
					Namespace: "ns",
				},
			}
			mesh4 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh4",
					Namespace: "ns",
				},
			}

			snap = input.NewInputLocalSnapshotManualBuilder("").
				AddDestinations(discoveryv1.DestinationSlice{destination}).
				AddMeshes(discoveryv1.MeshSlice{mesh1, mesh2, mesh3, mesh4})
		)

		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier = NewApplier(translator)
		})

		It("applies VirtualMesh with no federation", func() {
			permissiveVirtualMesh := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm1",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh1),
						ezkube.MakeObjectRef(mesh2),
						ezkube.MakeObjectRef(mesh3),
						ezkube.MakeObjectRef(mesh4),
					},
				},
			}

			snap.AddVirtualMeshes([]*networkingv1.VirtualMesh{permissiveVirtualMesh})

			applier.Apply(context.TODO(), snap.Build(), nil, nil)

			Expect(destination.Status.AppliedFederation).To(BeNil())
		})

		It("applies VirtualMesh with permissive federation using deprecated field", func() {
			permissiveVirtualMesh := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm1",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh1),
						ezkube.MakeObjectRef(mesh2),
						ezkube.MakeObjectRef(mesh3),
						ezkube.MakeObjectRef(mesh4),
					},
					Federation: &networkingv1.VirtualMeshSpec_Federation{
						Mode: &networkingv1.VirtualMeshSpec_Federation_Permissive{},
					},
				},
			}

			snap.AddVirtualMeshes([]*networkingv1.VirtualMesh{permissiveVirtualMesh})

			expectedAppliedFederation := &discoveryv1.DestinationStatus_AppliedFederation{
				FederatedHostname: "svc-name.svc-namespace.svc.svc-cluster.global",
				FederatedToMeshes: []*skv2corev1.ObjectRef{
					ezkube.MakeObjectRef(mesh2),
					ezkube.MakeObjectRef(mesh3),
					ezkube.MakeObjectRef(mesh4),
				},
				VirtualMeshRef: ezkube.MakeObjectRef(permissiveVirtualMesh),
			}

			applier.Apply(context.TODO(), snap.Build(), nil, nil)

			Expect(destination.Status.AppliedFederation).To(Equal(expectedAppliedFederation))
		})

		It("restrictive federation with empty selectors should have permissive semantics", func() {
			restrictiveVirtualMesh := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm1",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh1),
						ezkube.MakeObjectRef(mesh2),
						ezkube.MakeObjectRef(mesh3),
						ezkube.MakeObjectRef(mesh4),
					},
					Federation: &networkingv1.VirtualMeshSpec_Federation{
						Selectors: []*networkingv1.VirtualMeshSpec_Federation_FederationSelector{
							{},
						},
					},
				},
			}

			snap.AddVirtualMeshes([]*networkingv1.VirtualMesh{restrictiveVirtualMesh})

			expectedAppliedFederation := &discoveryv1.DestinationStatus_AppliedFederation{
				FederatedHostname: "svc-name.svc-namespace.svc.svc-cluster.global",
				FederatedToMeshes: []*skv2corev1.ObjectRef{
					ezkube.MakeObjectRef(mesh2),
					ezkube.MakeObjectRef(mesh3),
					ezkube.MakeObjectRef(mesh4),
				},
				VirtualMeshRef: ezkube.MakeObjectRef(restrictiveVirtualMesh),
			}

			applier.Apply(context.TODO(), snap.Build(), nil, nil)

			Expect(destination.Status.AppliedFederation).To(Equal(expectedAppliedFederation))
		})

		It("restrictive federation with defined selectors should selectively federate Destinations", func() {
			destination3 := &discoveryv1.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "d3",
					Namespace: "ns",
				},
				Spec: discoveryv1.DestinationSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "svc-name2",
								Namespace:   "svc-namespace",
								ClusterName: "svc-cluster",
							},
						},
					},
				},
			}

			restrictiveVirtualMesh := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm1",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh1),
						ezkube.MakeObjectRef(mesh2),
						ezkube.MakeObjectRef(mesh3),
						ezkube.MakeObjectRef(mesh4),
					},
					Federation: &networkingv1.VirtualMeshSpec_Federation{
						Selectors: []*networkingv1.VirtualMeshSpec_Federation_FederationSelector{
							{
								DestinationSelectors: []*commonv1.DestinationSelector{
									{
										KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
											Services: []*skv2corev1.ClusterObjectRef{
												{
													Name:        destination.Spec.GetKubeService().GetRef().GetName(),
													Namespace:   destination.Spec.GetKubeService().GetRef().GetNamespace(),
													ClusterName: destination.Spec.GetKubeService().GetRef().GetClusterName(),
												},
											},
										},
									},
								},
								Meshes: []*skv2corev1.ObjectRef{
									ezkube.MakeObjectRef(mesh2),
									ezkube.MakeObjectRef(mesh4),
								},
							},
							{
								DestinationSelectors: []*commonv1.DestinationSelector{
									{
										KubeServiceMatcher: &commonv1.DestinationSelector_KubeServiceMatcher{
											Namespaces: []string{destination.Spec.GetKubeService().GetRef().GetNamespace()},
											Clusters:   []string{destination.Spec.GetKubeService().GetRef().GetClusterName()},
										},
									},
								},
								// multiple references to the same mesh across different federation selectors should be deduplicated
								Meshes: []*skv2corev1.ObjectRef{
									ezkube.MakeObjectRef(mesh2),
									ezkube.MakeObjectRef(mesh3),
								},
							},
						},
					},
				},
			}

			snap.AddDestinations([]*discoveryv1.Destination{destination3})
			snap.AddVirtualMeshes([]*networkingv1.VirtualMesh{restrictiveVirtualMesh})

			expectedAppliedFederation1 := &discoveryv1.DestinationStatus_AppliedFederation{
				FederatedHostname: "svc-name.svc-namespace.svc.svc-cluster.global",
				FederatedToMeshes: []*skv2corev1.ObjectRef{
					ezkube.MakeObjectRef(mesh2),
					ezkube.MakeObjectRef(mesh3),
					ezkube.MakeObjectRef(mesh4),
				},
				VirtualMeshRef: ezkube.MakeObjectRef(restrictiveVirtualMesh),
			}

			expectedAppliedFederation2 := &discoveryv1.DestinationStatus_AppliedFederation{
				FederatedHostname: "svc-name2.svc-namespace.svc.svc-cluster.global",
				FederatedToMeshes: []*skv2corev1.ObjectRef{
					ezkube.MakeObjectRef(mesh2),
					ezkube.MakeObjectRef(mesh3),
				},
				VirtualMeshRef: ezkube.MakeObjectRef(restrictiveVirtualMesh),
			}

			applier.Apply(context.TODO(), snap.Build(), nil, nil)

			Expect(destination.Status.AppliedFederation).To(Equal(expectedAppliedFederation1))
			Expect(destination3.Status.AppliedFederation).To(Equal(expectedAppliedFederation2))
		})
	})

	Context("applies east west ingress gateways for Meshes configured by VirtualMesh", func() {
		var (
			applier Applier

			ingressDestinationForMesh = func(svcName, meshName string, serviceType discoveryv1.DestinationSpec_KubeService_ServiceType) *discoveryv1.Destination {
				return &discoveryv1.Destination{
					ObjectMeta: metav1.ObjectMeta{
						Name:        svcName,
						Namespace:   "ns",
						ClusterName: "cluster-name",
					},
					Spec: discoveryv1.DestinationSpec{
						Mesh: &skv2corev1.ObjectRef{
							Name:      meshName,
							Namespace: "ns",
						},
						Type: &discoveryv1.DestinationSpec_KubeService_{
							KubeService: &discoveryv1.DestinationSpec_KubeService{
								Ref: &skv2corev1.ClusterObjectRef{
									Name:        svcName,
									Namespace:   "svc-namespace",
									ClusterName: "cluster-name",
								},
								ServiceType:            serviceType,
								WorkloadSelectorLabels: defaults.DefaultGatewayWorkloadLabels,
								Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
									{
										Port: 1234,
										TargetPort: &discoveryv1.DestinationSpec_KubeService_KubeServicePort_TargetPortNumber{
											TargetPortNumber: 91234,
										},
										NodePort: 11234,
										Name:     defaults.IstioGatewayTlsPortName,
									},
									{
										Port: 5678,
										TargetPort: &discoveryv1.DestinationSpec_KubeService_KubeServicePort_TargetPortName{
											TargetPortName: "tls-port-name",
										},
										NodePort: 15678,
										Name:     "non-default-port",
									},
								},
								EndpointSubsets: []*discoveryv1.DestinationSpec_KubeService_EndpointsSubset{
									{
										Ports: []*discoveryv1.DestinationSpec_KubeService_EndpointPort{
											{
												Port: 95678,
												Name: "tls-port-name",
											},
										},
									},
								},
								ExternalAddresses: []*discoveryv1.DestinationSpec_KubeService_ExternalAddress{
									{
										ExternalAddressType: &discoveryv1.DestinationSpec_KubeService_ExternalAddress_DnsName{
											DnsName: "external-dns-name",
										},
									},
									{
										ExternalAddressType: &discoveryv1.DestinationSpec_KubeService_ExternalAddress_Ip{
											Ip: "external-ip",
										},
									},
								},
							},
						},
					},
				}
			}

			destination1 = ingressDestinationForMesh("svc-name1", "mesh1", discoveryv1.DestinationSpec_KubeService_NODE_PORT)
			destination2 = ingressDestinationForMesh("svc-name2", "mesh2", discoveryv1.DestinationSpec_KubeService_NODE_PORT)
			destination3 = ingressDestinationForMesh("svc-name3", "mesh3", discoveryv1.DestinationSpec_KubeService_LOAD_BALANCER)
			destination4 = ingressDestinationForMesh("svc-name4", "mesh4", discoveryv1.DestinationSpec_KubeService_LOAD_BALANCER)
			destination5 = ingressDestinationForMesh("svc-name5", "mesh4", discoveryv1.DestinationSpec_KubeService_LOAD_BALANCER)

			mesh1 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh1",
					Namespace: "ns",
				},
			}
			mesh2 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh2",
					Namespace: "ns",
				},
			}
			mesh3 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh3",
					Namespace: "ns",
				},
			}
			mesh4 = &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh4",
					Namespace: "ns",
				},
			}

			snap = input.NewInputLocalSnapshotManualBuilder("").
				AddMeshes(discoveryv1.MeshSlice{mesh1, mesh2, mesh3, mesh4})
		)

		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {
				// no report = accept
			}}
			applier = NewApplier(translator)
		})

		It("when ingress selectors are omitted, should first fallback on deprecated Mesh ingress gateway info", func() {

			mesh := &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh",
					Namespace: "ns",
				},
				Spec: discoveryv1.MeshSpec{
					Type: &discoveryv1.MeshSpec_Istio_{
						Istio: &discoveryv1.MeshSpec_Istio{
							Installation: &discoveryv1.MeshInstallation{
								Namespace: "mesh-ns",
								Cluster:   "cluster-name",
								Version:   "latest",
								PodLabels: map[string]string{"app": "istiod"},
							},
							IngressGateways: []*discoveryv1.MeshSpec_Istio_IngressGatewayInfo{
								{
									Name:             "svc-name1",
									Namespace:        "svc-namespace",
									ExternalTlsPort:  5678,
									TlsContainerPort: 95678,
								},
							},
						},
					},
				},
			}

			snap.AddDestinations(discoveryv1.DestinationSlice{destination1, destination2, destination3, destination4})
			snap.AddMeshes(discoveryv1.MeshSlice{mesh})

			defaultVirtualMesh := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm1",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh),
					},
				},
			}

			mesh.Status.AppliedVirtualMesh = &discoveryv1.MeshStatus_AppliedVirtualMesh{
				Ref: &skv2corev1.ObjectRef{
					Name:      "vm1",
					Namespace: "ns",
				},
			}

			snap.AddVirtualMeshes([]*networkingv1.VirtualMesh{defaultVirtualMesh})

			applier.Apply(context.TODO(), snap.Build(), nil, nil)

			Expect(len(mesh.Status.AppliedEastWestIngressGateways)).To(Equal(1))

			Expect(mesh.Status.AppliedEastWestIngressGateways[0]).To(Equal(&discoveryv1.MeshStatus_AppliedIngressGateway{
				DestinationRef:    ezkube.MakeObjectRef(destination1),
				ExternalAddresses: []string{"external-dns-name", "external-ip"},
				DestinationPort:   5678,
				ContainerPort:     95678,
			}))
		})

		It("when ingress selectors are omitted, should fallback on default east west ingress gateway selection", func() {

			snap.AddDestinations(discoveryv1.DestinationSlice{destination1, destination2, destination3, destination4})

			defaultVirtualMesh := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm1",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh1),
						ezkube.MakeObjectRef(mesh2),
						ezkube.MakeObjectRef(mesh3),
						ezkube.MakeObjectRef(mesh4),
					},
				},
			}
			for _, mesh := range []*discoveryv1.Mesh{mesh1, mesh2, mesh3, mesh4} {
				mesh.Status.AppliedVirtualMesh = &discoveryv1.MeshStatus_AppliedVirtualMesh{
					Ref: &skv2corev1.ObjectRef{
						Name:      "vm1",
						Namespace: "ns",
					},
				}
			}

			snap.AddVirtualMeshes([]*networkingv1.VirtualMesh{defaultVirtualMesh})

			applier.Apply(context.TODO(), snap.Build(), nil, nil)

			Expect(len(mesh1.Status.AppliedEastWestIngressGateways)).To(Equal(1))
			Expect(len(mesh2.Status.AppliedEastWestIngressGateways)).To(Equal(1))
			Expect(len(mesh3.Status.AppliedEastWestIngressGateways)).To(Equal(1))
			Expect(len(mesh4.Status.AppliedEastWestIngressGateways)).To(Equal(1))

			Expect(mesh1.Status.AppliedEastWestIngressGateways[0]).To(Equal(&discoveryv1.MeshStatus_AppliedIngressGateway{
				DestinationRef:    ezkube.MakeObjectRef(destination1),
				ExternalAddresses: []string{"external-dns-name", "external-ip"},
				DestinationPort:   11234,
				ContainerPort:     91234,
			}))
			Expect(mesh2.Status.AppliedEastWestIngressGateways[0]).To(Equal(&discoveryv1.MeshStatus_AppliedIngressGateway{
				DestinationRef:    ezkube.MakeObjectRef(destination2),
				ExternalAddresses: []string{"external-dns-name", "external-ip"},
				DestinationPort:   11234,
				ContainerPort:     91234,
			}))
			Expect(mesh3.Status.AppliedEastWestIngressGateways[0]).To(Equal(&discoveryv1.MeshStatus_AppliedIngressGateway{
				DestinationRef:    ezkube.MakeObjectRef(destination3),
				ExternalAddresses: []string{"external-dns-name", "external-ip"},
				DestinationPort:   1234,
				ContainerPort:     91234,
			}))
			Expect(mesh4.Status.AppliedEastWestIngressGateways[0]).To(Equal(&discoveryv1.MeshStatus_AppliedIngressGateway{
				DestinationRef:    ezkube.MakeObjectRef(destination4),
				ExternalAddresses: []string{"external-dns-name", "external-ip"},
				DestinationPort:   1234,
				ContainerPort:     91234,
			}))
		})

		It("selects east west ingress gateways on a VirtualMesh with east west ingress gateway selectors", func() {
			snap.AddDestinations(discoveryv1.DestinationSlice{destination1, destination2, destination3, destination4, destination5})

			virtualMeshWithSelector1 := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm1",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh1),
					},
					Federation: &networkingv1.VirtualMeshSpec_Federation{
						EastWestIngressGatewaySelectors: []*commonv1.IngressGatewaySelector{
							{
								DestinationSelectors: []*commonv1.DestinationSelector{
									{
										KubeServiceMatcher: &commonv1.DestinationSelector_KubeServiceMatcher{
											Namespaces: []string{"svc-namespace"},
										},
									},
								},
							},
						},
					},
				},
			}
			mesh1.Status.AppliedVirtualMesh = &discoveryv1.MeshStatus_AppliedVirtualMesh{
				Ref: &skv2corev1.ObjectRef{
					Name:      "vm1",
					Namespace: "ns",
				},
			}

			virtualMeshWithSelector2 := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm2",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh2),
					},
					Federation: &networkingv1.VirtualMeshSpec_Federation{
						EastWestIngressGatewaySelectors: []*commonv1.IngressGatewaySelector{
							{
								DestinationSelectors: []*commonv1.DestinationSelector{
									{
										KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
											Services: []*skv2corev1.ClusterObjectRef{
												ezkube.MakeClusterObjectRef(destination2.Spec.GetKubeService().Ref),
											},
										},
									},
								},
								PortName: "non-default-port",
							},
						},
					},
				},
			}
			mesh2.Status.AppliedVirtualMesh = &discoveryv1.MeshStatus_AppliedVirtualMesh{
				Ref: &skv2corev1.ObjectRef{
					Name:      "vm2",
					Namespace: "ns",
				},
			}

			virtualMeshWithSelector3 := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm3",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh3),
						ezkube.MakeObjectRef(mesh4),
					},
					Federation: &networkingv1.VirtualMeshSpec_Federation{
						EastWestIngressGatewaySelectors: []*commonv1.IngressGatewaySelector{
							{
								DestinationSelectors: []*commonv1.DestinationSelector{
									{
										KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
											Services: []*skv2corev1.ClusterObjectRef{
												ezkube.MakeClusterObjectRef(destination3.Spec.GetKubeService().Ref),
											},
										},
									},
								},
								PortName: "non-default-port",
							},
							{
								DestinationSelectors: []*commonv1.DestinationSelector{
									{
										KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
											Services: []*skv2corev1.ClusterObjectRef{
												ezkube.MakeClusterObjectRef(destination4.Spec.GetKubeService().GetRef()),
											},
										},
									},
								},
								PortName: defaults.IstioGatewayTlsPortName,
							},
							// should deduplicate "svc-name" and add "svc-name2"
							{
								DestinationSelectors: []*commonv1.DestinationSelector{
									{
										KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
											Services: []*skv2corev1.ClusterObjectRef{
												ezkube.MakeClusterObjectRef(destination4.Spec.GetKubeService().GetRef()),
												ezkube.MakeClusterObjectRef(destination5.Spec.GetKubeService().GetRef()),
											},
										},
									},
								},
								PortName: defaults.IstioGatewayTlsPortName,
							},
						},
					},
				},
			}

			snap.AddVirtualMeshes([]*networkingv1.VirtualMesh{
				virtualMeshWithSelector1,
				virtualMeshWithSelector2,
				virtualMeshWithSelector3,
			})

			applier.Apply(context.TODO(), snap.Build(), nil, nil)

			Expect(len(mesh1.Status.AppliedEastWestIngressGateways)).To(Equal(1))
			Expect(len(mesh2.Status.AppliedEastWestIngressGateways)).To(Equal(1))
			Expect(len(mesh3.Status.AppliedEastWestIngressGateways)).To(Equal(1))
			Expect(len(mesh4.Status.AppliedEastWestIngressGateways)).To(Equal(2))

			Expect(mesh1.Status.AppliedEastWestIngressGateways[0]).To(Equal(&discoveryv1.MeshStatus_AppliedIngressGateway{
				DestinationRef:    ezkube.MakeObjectRef(destination1),
				ExternalAddresses: []string{"external-dns-name", "external-ip"},
				DestinationPort:   11234,
				ContainerPort:     91234,
			}))
			Expect(mesh2.Status.AppliedEastWestIngressGateways[0]).To(Equal(&discoveryv1.MeshStatus_AppliedIngressGateway{
				DestinationRef:    ezkube.MakeObjectRef(destination2),
				ExternalAddresses: []string{"external-dns-name", "external-ip"},
				DestinationPort:   15678,
				ContainerPort:     95678,
			}))
			Expect(mesh3.Status.AppliedEastWestIngressGateways[0]).To(Equal(&discoveryv1.MeshStatus_AppliedIngressGateway{
				DestinationRef:    ezkube.MakeObjectRef(destination3),
				ExternalAddresses: []string{"external-dns-name", "external-ip"},
				DestinationPort:   5678,
				ContainerPort:     95678,
			}))
			Expect(mesh4.Status.AppliedEastWestIngressGateways).To(ConsistOf([]*discoveryv1.MeshStatus_AppliedIngressGateway{
				{
					DestinationRef:    ezkube.MakeObjectRef(destination4),
					ExternalAddresses: []string{"external-dns-name", "external-ip"},
					DestinationPort:   1234,
					ContainerPort:     91234,
				},
				{
					DestinationRef:    ezkube.MakeObjectRef(destination5),
					ExternalAddresses: []string{"external-dns-name", "external-ip"},
					DestinationPort:   1234,
					ContainerPort:     91234,
				},
			}))
		})

		It("put status errors on a VirtualMesh with bad selectors", func() {
			snap.AddDestinations([]*discoveryv1.Destination{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bad-destination-no-workload-labels",
						Namespace: "ns",
					},
					Spec: discoveryv1.DestinationSpec{
						Mesh: &skv2corev1.ObjectRef{
							Name:      "mesh1",
							Namespace: "ns",
						},
						Type: &discoveryv1.DestinationSpec_KubeService_{
							KubeService: &discoveryv1.DestinationSpec_KubeService{
								Ref: &skv2corev1.ClusterObjectRef{
									Name:        "svc-name-no-labels",
									Namespace:   "svc-namespace-no-labels",
									ClusterName: "svc-cluster",
								},
								Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
									{
										Port: 1234,
										Name: defaults.IstioGatewayTlsPortName,
									},
									{
										Port: 5678,
										Name: "non-default-port",
									},
								},
							},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "bad-destination-no-tls-port",
						Namespace: "ns",
					},
					Spec: discoveryv1.DestinationSpec{
						Mesh: &skv2corev1.ObjectRef{
							Name:      "mesh2",
							Namespace: "ns",
						},
						Type: &discoveryv1.DestinationSpec_KubeService_{
							KubeService: &discoveryv1.DestinationSpec_KubeService{
								Ref: &skv2corev1.ClusterObjectRef{
									Name:        "svc-name-no-tls-port",
									Namespace:   "svc-namespace-no-tls-port",
									ClusterName: "svc-cluster",
								},
								WorkloadSelectorLabels: defaults.DefaultGatewayWorkloadLabels,
								Ports: []*discoveryv1.DestinationSpec_KubeService_KubeServicePort{
									{
										Port: 5678,
										Name: "non-tls-port",
									},
								},
							},
						},
					},
				},
			})
			virtualMeshWithBadSelector1 := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm1",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh1),
					},
					Federation: &networkingv1.VirtualMeshSpec_Federation{
						EastWestIngressGatewaySelectors: []*commonv1.IngressGatewaySelector{
							{
								DestinationSelectors: []*commonv1.DestinationSelector{
									{
										KubeServiceMatcher: &commonv1.DestinationSelector_KubeServiceMatcher{
											Namespaces: []string{"svc-namespace-no-labels"},
										},
									},
								},
								PortName: "non-default-port",
							},
						},
					},
				},
			}
			mesh1.Status.AppliedVirtualMesh = &discoveryv1.MeshStatus_AppliedVirtualMesh{
				Ref: &skv2corev1.ObjectRef{
					Name:      "vm1",
					Namespace: "ns",
				},
			}

			virtualMeshWithBadSelector2 := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm2",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh2),
					},
					Federation: &networkingv1.VirtualMeshSpec_Federation{
						EastWestIngressGatewaySelectors: []*commonv1.IngressGatewaySelector{
							{
								DestinationSelectors: []*commonv1.DestinationSelector{
									{
										KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
											Services: []*skv2corev1.ClusterObjectRef{
												{
													Name:        "svc-name-no-tls-port",
													Namespace:   "svc-namespace-no-tls-port",
													ClusterName: "svc-cluster",
												},
											},
										},
									},
								},
								PortName: defaults.IstioGatewayTlsPortName,
							},
						},
					},
				},
			}
			mesh2.Status.AppliedVirtualMesh = &discoveryv1.MeshStatus_AppliedVirtualMesh{
				Ref: &skv2corev1.ObjectRef{
					Name:      "vm2",
					Namespace: "ns",
				},
			}

			virtualMeshWithBadSelector3 := &networkingv1.VirtualMesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vm3",
					Namespace: "ns",
				},
				Spec: networkingv1.VirtualMeshSpec{
					Meshes: []*skv2corev1.ObjectRef{
						ezkube.MakeObjectRef(mesh3),
					},
					Federation: &networkingv1.VirtualMeshSpec_Federation{
						EastWestIngressGatewaySelectors: []*commonv1.IngressGatewaySelector{
							{
								DestinationSelectors: []*commonv1.DestinationSelector{
									{
										KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
											Services: []*skv2corev1.ClusterObjectRef{
												{
													Name:        "nonexistent-svc",
													Namespace:   "nonexistent-svc",
													ClusterName: "svc-cluster",
												},
											},
										},
									},
								},
								PortName: defaults.IstioGatewayTlsPortName,
							},
						},
					},
				},
			}
			mesh3.Status.AppliedVirtualMesh = &discoveryv1.MeshStatus_AppliedVirtualMesh{
				Ref: &skv2corev1.ObjectRef{
					Name:      "vm3",
					Namespace: "ns",
				},
			}

			snap.AddVirtualMeshes([]*networkingv1.VirtualMesh{virtualMeshWithBadSelector1, virtualMeshWithBadSelector2,
				virtualMeshWithBadSelector3})

			applier.Apply(context.TODO(), snap.Build(), nil, nil)

			Expect(virtualMeshWithBadSelector1.Status.GetState()).To(Equal(commonv1.ApprovalState_INVALID))
			Expect(len(virtualMeshWithBadSelector1.Status.Errors)).To(Equal(1))
			Expect(virtualMeshWithBadSelector1.Status.Errors[0]).
				To(Equal("attempting to select ingress gateway destination bad-destination-no-workload-labels.ns. with no selector labels"))

			Expect(virtualMeshWithBadSelector2.Status.GetState()).To(Equal(commonv1.ApprovalState_INVALID))
			Expect(len(virtualMeshWithBadSelector2.Status.Errors)).To(Equal(1))
			Expect(virtualMeshWithBadSelector2.Status.Errors[0]).
				To(Equal("ingress gateway destination port info could not be determined for tls port name: tls"))

			Expect(virtualMeshWithBadSelector3.Status.GetState()).To(Equal(commonv1.ApprovalState_INVALID))
			Expect(len(virtualMeshWithBadSelector3.Status.Errors)).To(Equal(1))
			Expect(virtualMeshWithBadSelector3.Status.Errors[0]).
				To(Equal("Invalid Destination selector: Destination nonexistent-svc.nonexistent-svc.svc-cluster not found"))
		})

	})

	Context("required subsets", func() {
		var applier Applier

		BeforeEach(func() {
			translator := testIstioTranslator{callReporter: func(reporter reporting.Reporter) {}}
			applier = NewApplier(translator)
		})

		It("computes a Destination's required subsets", func() {
			destination := &discoveryv1.Destination{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ms1",
					Namespace: "ns",
				},
				Spec: discoveryv1.DestinationSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
					Type: &discoveryv1.DestinationSpec_KubeService_{
						KubeService: &discoveryv1.DestinationSpec_KubeService{
							Ref: &skv2corev1.ClusterObjectRef{
								Name:        "svc-name",
								Namespace:   "svc-namespace",
								ClusterName: "svc-cluster",
							},
						},
					},
				},
			}
			workload := &discoveryv1.Workload{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "wkld1",
					Namespace: "ns",
				},
				Spec: discoveryv1.WorkloadSpec{
					Mesh: &skv2corev1.ObjectRef{
						Name:      "mesh1",
						Namespace: "ns",
					},
				},
			}
			mesh := &discoveryv1.Mesh{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mesh1",
					Namespace: "ns",
				},
			}
			trafficPolicy1 := &networkingv1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp1",
					Namespace: "ns",
				},
				Spec: networkingv1.TrafficPolicySpec{
					Policy: &networkingv1.TrafficPolicySpec_Policy{
						TrafficShift: &networkingv1.TrafficPolicySpec_Policy_MultiDestination{
							Destinations: []*networkingv1.WeightedDestination{
								{
									Weight: 50,
									DestinationType: &networkingv1.WeightedDestination_KubeService{
										KubeService: &networkingv1.WeightedDestination_KubeDestination{
											Name:        destination.Spec.GetKubeService().Ref.Name,
											Namespace:   destination.Spec.GetKubeService().Ref.Namespace,
											ClusterName: destination.Spec.GetKubeService().Ref.ClusterName,
											Subset: map[string]string{
												"version": "v1",
											},
										},
									},
								},
								{
									Weight: 50,
									DestinationType: &networkingv1.WeightedDestination_KubeService{
										KubeService: &networkingv1.WeightedDestination_KubeDestination{
											Name:        "ignored",
											Namespace:   "ignored",
											ClusterName: "ignored",
											Subset: map[string]string{
												"version": "v2",
											},
										},
									},
								},
							},
						},
					},
				},
			}
			trafficPolicy2 := &networkingv1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp2",
					Namespace: "ns",
				},
				Spec: networkingv1.TrafficPolicySpec{
					Policy: &networkingv1.TrafficPolicySpec_Policy{
						TrafficShift: &networkingv1.TrafficPolicySpec_Policy_MultiDestination{
							Destinations: []*networkingv1.WeightedDestination{
								{
									Weight: 50,
									DestinationType: &networkingv1.WeightedDestination_KubeService{
										KubeService: &networkingv1.WeightedDestination_KubeDestination{
											Name:        destination.Spec.GetKubeService().Ref.Name,
											Namespace:   destination.Spec.GetKubeService().Ref.Namespace,
											ClusterName: destination.Spec.GetKubeService().Ref.ClusterName,
											Subset: map[string]string{
												"version": "v2",
											},
										},
									},
								},
								{
									Weight: 50,
									DestinationType: &networkingv1.WeightedDestination_KubeService{
										KubeService: &networkingv1.WeightedDestination_KubeDestination{
											Name:        "ignored",
											Namespace:   "ignored",
											ClusterName: "ignored",
											Subset: map[string]string{
												"version": "v2",
											},
										},
									},
								},
							},
						},
					},
				},
			}
			// should not be in required subsets
			invalidTrafficPolicy := &networkingv1.TrafficPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "tp3",
					Namespace: "ns",
				},
				Spec: networkingv1.TrafficPolicySpec{
					DestinationSelector: []*commonv1.DestinationSelector{
						{
							KubeServiceRefs: &commonv1.DestinationSelector_KubeServiceRefs{
								Services: []*skv2corev1.ClusterObjectRef{
									{
										Name:        "nonexistent",
										Namespace:   "namespace",
										ClusterName: "cluster",
									},
								},
							},
						},
					},
					Policy: &networkingv1.TrafficPolicySpec_Policy{
						TrafficShift: &networkingv1.TrafficPolicySpec_Policy_MultiDestination{
							Destinations: []*networkingv1.WeightedDestination{
								{
									DestinationType: &networkingv1.WeightedDestination_KubeService{
										KubeService: &networkingv1.WeightedDestination_KubeDestination{
											Name:        destination.Spec.GetKubeService().Ref.Name,
											Namespace:   destination.Spec.GetKubeService().Ref.Namespace,
											ClusterName: destination.Spec.GetKubeService().Ref.ClusterName,
											Subset: map[string]string{
												"version": "v3",
											},
										},
									},
								},
							},
						},
					},
				},
			}

			snap := input.NewInputLocalSnapshotManualBuilder("").
				AddDestinations(discoveryv1.DestinationSlice{destination}).
				AddTrafficPolicies(networkingv1.TrafficPolicySlice{trafficPolicy1, trafficPolicy2, invalidTrafficPolicy}).
				AddWorkloads(discoveryv1.WorkloadSlice{workload}).
				AddMeshes(discoveryv1.MeshSlice{mesh}).
				Build()

			applier.Apply(context.TODO(), snap, nil, nil)

			expectedRequiredSubsets := []*discoveryv1.RequiredSubsets{
				{
					TrafficPolicyRef:   ezkube.MakeObjectRef(trafficPolicy1),
					ObservedGeneration: trafficPolicy1.Generation,
					TrafficShift:       trafficPolicy1.Spec.Policy.TrafficShift,
				},
				{
					TrafficPolicyRef:   ezkube.MakeObjectRef(trafficPolicy2),
					ObservedGeneration: trafficPolicy2.Generation,
					TrafficShift:       trafficPolicy2.Spec.Policy.TrafficShift,
				},
			}

			Expect(destination.Status.RequiredSubsets).To(Equal(expectedRequiredSubsets))
		})

	})
})

// NOTE(ilackarms): we implement a test translator here instead of using a mock because
// we need to call methods on the reporter which is passed as an argument to the translator
type testIstioTranslator struct {
	callReporter func(reporter reporting.Reporter)
}

func (t testIstioTranslator) Translate(
	ctx context.Context,
	in input.LocalSnapshot,
	_ input.RemoteSnapshot,
	_ input.RemoteSnapshot,
	reporter reporting.Reporter,
) (*translation.Outputs, error) {
	t.callReporter(reporter)
	return &translation.Outputs{}, nil
}
