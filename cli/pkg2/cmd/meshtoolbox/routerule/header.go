package routerule

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	superglooV1 "github.com/solo-io/supergloo/pkg2/api/v1"
)

func EnsureHeaderManipulation(irOpts *options.InputHeaderManipulation, opts *options.Options) error {
	staging := &superglooV1.HeaderManipulation{
		RemoveResponseHeaders: []string{},
		AppendResponseHeaders: make(map[string]string),
		RemoveRequestHeaders:  []string{},
		AppendRequestHeaders:  make(map[string]string),
	}

	static := opts.Top.Static || opts.Top.File != ""

	// Response
	if err := ensureCsv("Please specify headers to remove from the response", irOpts.RemoveResponseHeaders, &staging.RemoveResponseHeaders, static, true); err != nil {
		return err
	}
	if err := ensureKVCsv("Please specify headers to append to the response", irOpts.AppendResponseHeaders, &staging.AppendResponseHeaders, static, true); err != nil {
		return err
	}

	// Request
	if err := ensureCsv("Please specify headers to remove from the request", irOpts.RemoveRequestHeaders, &staging.RemoveRequestHeaders, static, true); err != nil {
		return err
	}
	if err := ensureKVCsv("Please specify headers to append to the request", irOpts.AppendRequestHeaders, &staging.AppendRequestHeaders, static, true); err != nil {
		return err
	}

	opts.MeshTool.RoutingRule.HeaderManipulaition = staging
	return nil
}
