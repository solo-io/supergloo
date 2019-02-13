package routerule

import (
	types "github.com/gogo/protobuf/types"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	v1alpha3 "github.com/solo-io/supergloo/pkg2/api/external/istio/networking/v1alpha3"
)

func EnsureCors(irOpts *options.InputCors, opts *options.Options) error {
	cOpts := &(opts.Create.InputRoutingRule).Cors
	// initialize the field
	target := &v1alpha3.CorsPolicy{}

	static := opts.Top.Static || opts.Top.File != ""

	if err := ensureCsv("Please specify the allowed origins (comma-separated list)", cOpts.AllowOrigin, &target.AllowOrigin, static, true); err != nil {
		return err
	}
	if err := ensureCsv("Please specify the allowed methods (comma-separated list)", cOpts.AllowMethods, &target.AllowMethods, static, true); err != nil {
		return err
	}
	if err := ensureCsv("Please specify the allowed headers (comma-separated list)", cOpts.AllowHeaders, &target.AllowHeaders, static, true); err != nil {
		return err
	}
	if err := ensureCsv("Please specify the exposed headers (comma-separated list)", cOpts.ExposeHeaders, &target.ExposeHeaders, static, true); err != nil {
		return err
	}
	target.MaxAge = &types.Duration{}
	if err := EnsureDuration("Please specify the max age", &cOpts.MaxAge, target.MaxAge, opts); err != nil {
		return err
	}

	opts.MeshTool.RoutingRule.CorsPolicy = target
	return nil
}
