package routerule

import (
	"github.com/solo-io/supergloo/cli/pkg/common"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
)

func CreateRoutingRule(ruleTypeID string, opts *options.Options) error {

	rrOpts := &(opts.Create).InputRoutingRule
	if err := AssembleRoutingRule(ruleTypeID, &rrOpts.ActiveTypes, opts); err != nil {
		return err
	}

	rrClient, err := common.GetRoutingRuleClient()
	if err != nil {
		return err
	}

	_, err = (*rrClient).Write(&(opts.MeshTool).RoutingRule, clients.WriteOpts{OverwriteExisting: rrOpts.OverrideExisting})
	return err
}
