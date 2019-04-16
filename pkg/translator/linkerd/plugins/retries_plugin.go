package plugins

import (
	"time"

	"github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/vektah/gqlgen/neelance/errors"
)

type retriesPlugin struct{}

func NewRetriesPlugin() *retriesPlugin {}

func (r *retriesPlugin) Init(params InitParams) error {
	return nil
}

func (r *retriesPlugin) ProcessRoutes(params Params, in v1.RoutingRuleSpec, out []*v1alpha1.RouteSpec) error {
	retryPolicy := in.GetRetries()
	if retryPolicy == nil || retryPolicy.LinkerdRetries == nil {
		return nil
	}
	for _, route := range out {
		route.IsRetryable = true
	}
	return nil
}

func (r *retriesPlugin) ProcessServiceProfile(params Params, in v1.RoutingRuleSpec, out *v1alpha1.ServiceProfileSpec) error {
	retryPolicy := in.GetRetries()
	if retryPolicy == nil || retryPolicy.LinkerdRetries == nil || retryPolicy.LinkerdRetries.RetryBudget == nil {
		return nil
	}
	retryBudget := retryPolicy.LinkerdRetries.RetryBudget
	if retryBudget.RetryRatio < 0 || retryBudget.RetryRatio > 1 {
		return errors.Errorf("retryRatio must be a percentage value between 0 and 1")
	}
	_, err := time.ParseDuration(retryBudget.Ttl)
	if err != nil {
		return err
	}

	out.RetryBudget = &v1alpha1.RetryBudget{
		MinRetriesPerSecond: retryBudget.MinRetriesPerSecond,
		RetryRatio:          retryBudget.RetryRatio,
		TTL:                 retryBudget.Ttl,
	}

	return nil
}
