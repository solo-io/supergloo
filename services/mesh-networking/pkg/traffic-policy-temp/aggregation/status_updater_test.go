package traffic_policy_aggregation_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	traffic_policy_aggregation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/traffic-policy-temp/aggregation"
)

var _ = Describe("StatusUpdater", func() {
	statusUpdater := traffic_policy_aggregation.NewInMemoryStatusMutator()

	Context("when updating service policies", func() {
		Context("if the length of the status policies is different than the length of the new policies", func() {
			It("returns true when going from nonzero -> nil", func() {
				ms := &smh_discovery.MeshService{
					Status: smh_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
							Ref: &smh_core_types.ResourceRef{Name: "foo"},
						}},
					},
				}
				newPolicies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy(nil)

				Expect(statusUpdater.MutateServicePolicies(ms, newPolicies)).To(BeTrue())
				Expect(ms.Status.ValidatedTrafficPolicies).To(BeNil())
			})

			It("returns true when going from nonzero -> non-nil and empty", func() {
				ms := &smh_discovery.MeshService{
					Status: smh_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
							Ref: &smh_core_types.ResourceRef{Name: "foo"},
						}},
					},
				}
				newPolicies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{}

				Expect(statusUpdater.MutateServicePolicies(ms, newPolicies)).To(BeTrue())
				Expect(ms.Status.ValidatedTrafficPolicies).To(BeEmpty())
			})

			It("returns true when the lengths are different", func() {
				ms := &smh_discovery.MeshService{
					Status: smh_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
							Ref: &smh_core_types.ResourceRef{Name: "foo"},
						}},
					},
				}
				newPolicies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{
					{
						Ref: &smh_core_types.ResourceRef{Name: "bar"},
					},
					{
						Ref: &smh_core_types.ResourceRef{Name: "bar-2.0"},
					},
				}

				Expect(statusUpdater.MutateServicePolicies(ms, newPolicies)).To(BeTrue())
				Expect(ms.Status.ValidatedTrafficPolicies).To(Equal(newPolicies))
			})
		})

		Context("if the length of the status has not changed compared to the incoming length", func() {
			It("returns true if an entry in the list has changed", func() {
				ms := &smh_discovery.MeshService{
					Status: smh_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
							Ref: &smh_core_types.ResourceRef{Name: "foo"},
						}},
					},
				}
				newPolicies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
					Ref: &smh_core_types.ResourceRef{Name: "bar"},
				}}

				Expect(statusUpdater.MutateServicePolicies(ms, newPolicies)).To(BeTrue())
				Expect(ms.Status.ValidatedTrafficPolicies).To(Equal(newPolicies))
			})

			It("returns false if the list has not changed", func() {
				ms := &smh_discovery.MeshService{
					Status: smh_discovery_types.MeshServiceStatus{
						ValidatedTrafficPolicies: []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
							Ref: &smh_core_types.ResourceRef{Name: "foo"},
						}},
					},
				}
				newPolicies := []*smh_discovery_types.MeshServiceStatus_ValidatedTrafficPolicy{{
					Ref: &smh_core_types.ResourceRef{Name: "foo"},
				}}

				Expect(statusUpdater.MutateServicePolicies(ms, newPolicies)).To(BeFalse())
				Expect(ms.Status.ValidatedTrafficPolicies).To(Equal(newPolicies))
			})
		})
	})

	Context("when updating policy statuses", func() {
		Context("when only the conflict errors have changed", func() {
			It("returns true if going from zero -> nonzero", func() {
				policy := &smh_networking.TrafficPolicy{}
				conflictErrors := []*smh_networking_types.TrafficPolicyStatus_ConflictError{{
					ErrorMessage: "whoops",
				}}

				Expect(statusUpdater.MutateConflictAndTranslatorErrors(policy, conflictErrors, nil)).To(BeTrue())
				Expect(policy.Status.ConflictErrors).To(Equal(conflictErrors))
			})

			It("returns true if going from nonzero -> zero", func() {
				conflictErrors := []*smh_networking_types.TrafficPolicyStatus_ConflictError{{
					ErrorMessage: "whoops",
				}}
				policy := &smh_networking.TrafficPolicy{
					Status: smh_networking_types.TrafficPolicyStatus{
						ConflictErrors: conflictErrors,
					},
				}
				Expect(statusUpdater.MutateConflictAndTranslatorErrors(policy, nil, nil)).To(BeTrue())
				Expect(policy.Status.ConflictErrors).To(BeNil())
			})

			It("can compare items in the list for changes", func() {
				conflictErrors := []*smh_networking_types.TrafficPolicyStatus_ConflictError{{
					ErrorMessage: "whoops",
				}}
				policy := &smh_networking.TrafficPolicy{
					Status: smh_networking_types.TrafficPolicyStatus{
						ConflictErrors: conflictErrors,
					},
				}

				newConflictErrors := []*smh_networking_types.TrafficPolicyStatus_ConflictError{{
					ErrorMessage: "new message",
				}}
				Expect(statusUpdater.MutateConflictAndTranslatorErrors(policy, newConflictErrors, nil)).To(BeTrue())
				Expect(policy.Status.ConflictErrors).To(Equal(newConflictErrors))
			})
		})
	})
})
