package trafficpolicy


import "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"

// the Validator validates TrafficPolicies.
type Validator interface {
	// returns an error if the policy is invalid or conflicts with an existing (Accepted) traffic policy.
	ValidateTrafficPolicy(policy *v1alpha1.TrafficPolicy) error
}

