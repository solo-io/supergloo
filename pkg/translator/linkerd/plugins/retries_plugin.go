package plugins

import (
	"github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/vektah/gqlgen/neelance/errors"
)

type retriesPlugin struct{}

func NewRetriesPlugin() *retriesPlugin {
	return &retriesPlugin{}
}

func (r *retriesPlugin) Init(params InitParams) error {
	return nil
}

func (r *retriesPlugin) ProcessRoutes(params Params, in v1.RoutingRuleSpec, out []*v1alpha1.RouteSpec) error {
	retryPolicy := in.GetRetries()
	if retryPolicy == nil || retryPolicy.RetryBudget == nil {
		return nil
	}

	for _, route := range out {
		route.IsRetryable = true
	}
	return nil
}

func (r *retriesPlugin) ProcessServiceProfile(params Params, in v1.RoutingRuleSpec, out *v1alpha1.ServiceProfileSpec) error {
	retryPolicy := in.GetRetries()
	if retryPolicy == nil || retryPolicy.RetryBudget == nil {
		return nil
	}
	retryBudget := retryPolicy.RetryBudget
	if retryBudget.RetryRatio < 0 || retryBudget.RetryRatio > 1 {
		return errors.Errorf("retryRatio must be a percentage value between 0 and 1")
	}
	out.RetryBudget = &v1alpha1.RetryBudget{
		MinRetriesPerSecond: retryBudget.MinRetriesPerSecond,
		RetryRatio:          retryBudget.RetryRatio,
		TTL:                 retryBudget.Ttl.String(),
	}

	// user specified no request matchers, create a default route
	if len(out.Routes) == 0 {
		out.Routes = []*v1alpha1.RouteSpec{{
			Name:        "default",
			IsRetryable: true,
			Condition: &v1alpha1.RequestMatch{
				PathRegex: ".*",
			},
		}}
	}

	return nil
}
