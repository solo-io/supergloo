package routerule

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	superglooV1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func EnsureHeaderManipulation(irOpts *options.InputHeaderManipulation, opts *options.Options) error {
	staging := &superglooV1.HeaderManipulation{
		RemoveResponseHeaders: []string{},
		AppendResponseHeaders: make(map[string]string),
		RemoveRequestHeaders:  []string{},
		AppendRequestHeaders:  make(map[string]string),
	}

	// Response
	if err := ensureCsv("Please specify headers to remove from the response", irOpts.RemoveResponseHeaders, &staging.RemoveResponseHeaders, opts.Top.Static); err != nil {
		return nil
	}
	if err := ensureKVCsv("Please specify headers to append to the response", irOpts.AppendResponseHeaders, &staging.AppendResponseHeaders, opts.Top.Static); err != nil {
		return nil
	}

	// Request
	if err := ensureCsv("Please specify headers to remove from the request", irOpts.RemoveRequestHeaders, &staging.RemoveRequestHeaders, opts.Top.Static); err != nil {
		return nil
	}
	if err := ensureKVCsv("Please specify headers to append to the request", irOpts.AppendRequestHeaders, &staging.AppendRequestHeaders, opts.Top.Static); err != nil {
		return nil
	}

	opts.MeshTool.RoutingRule.HeaderManipulaition = staging
	return nil
}
