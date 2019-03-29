package apply

import (
	"github.com/gogo/protobuf/types"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func faultInjectionConvertSpecFunc(opts options.RoutingRuleSpec) (*v1.RoutingRuleSpec, error) {
	if opts.FaultInjection.Percent == 0 {
		return nil, errors.Errorf("invalid value %v: percentage cannot be zero", opts.FaultInjection.Percent)
	}
	spec := &v1.RoutingRuleSpec{}
	faultInjection := &v1.FaultInjection{
		Percentage: opts.FaultInjection.Percent,
	}
	if opts.FaultInjection.Abort.Http.HttpStatus != 0 {
		faultInjection.FaultInjectionType = &v1.FaultInjection_Abort_{
			Abort: &v1.FaultInjection_Abort{
				ErrorType: &v1.FaultInjection_Abort_HttpStatus{
					HttpStatus: opts.FaultInjection.Abort.Http.HttpStatus,
				},
			},
		}
		spec.RuleType = &v1.RoutingRuleSpec_FaultInjection{
			FaultInjection: faultInjection,
		}
		return spec, nil
	}

	if opts.FaultInjection.Delay.Fixed != 0 {
		faultInjection.FaultInjectionType = &v1.FaultInjection_Delay_{
			Delay: &v1.FaultInjection_Delay{
				HttpDelayType: &v1.FaultInjection_Delay_FixedDelay{
					FixedDelay: &types.Duration{
						Seconds: int64(opts.FaultInjection.Delay.Fixed.Seconds()),
						Nanos:   int32(opts.FaultInjection.Delay.Fixed.Nanoseconds()),
					},
				},
			},
		}
		spec.RuleType = &v1.RoutingRuleSpec_FaultInjection{
			FaultInjection: faultInjection,
		}
		return spec, nil
	}

	return nil, errors.Errorf("no fault injection type specified")
}

var faultInjectionSpecCommand = routingRuleSpecCommand{
	use:             "faultinjection",
	alias:           "fi",
	short:           "apply a fault injection rule",
	long:            `Fault injection rules are used to inject faults into requests in order to test for tolerance.`,
	specSurveyFunc:  surveyutils.SurveyFaultInjectionSpec,
	convertSpecFunc: faultInjectionConvertSpecFunc,
	addFlagsFunc:    flagutils.AddFaultInjectionFlags,
	subCmds:         faultInjectionTypes,
	mutateOpts: func(opts *options.Options) {
		opts.Interactive = true
	},
}

var faultInjectionTypes = []routingRuleSpecCommand{
	{
		use:            "delay",
		alias:          "d",
		short:          "apply a delay type fault injection rule",
		addFlagsFunc:   flagutils.AddFaultInjectionDelayFlags,
		specSurveyFunc: surveyutils.SurveyFaultInjectionPercent,
		subCmds: []routingRuleSpecCommand{
			{
				use:             "fixed",
				alias:           "f",
				short:           "apply a fixed delay type fault injection rule",
				convertSpecFunc: faultInjectionConvertSpecFunc,
				specSurveyFunc:  surveyutils.SurveyFaultInjectionDelayFixed,
			},
		},
	},
	{
		use:            "abort",
		alias:          "a",
		short:          "apply an abort type fault injection rule",
		addFlagsFunc:   flagutils.AddFaultInjectionAbortFlags,
		specSurveyFunc: surveyutils.SurveyFaultInjectionPercent,
		subCmds: []routingRuleSpecCommand{
			{
				use:             "http",
				alias:           "http",
				short:           "apply an http abort type fault injection rule",
				convertSpecFunc: faultInjectionConvertSpecFunc,
				specSurveyFunc:  surveyutils.SurveyFaultInjectionAbortHttp,
			},
		},
	},
}

// func faultInjectionCmd(opts *options.Options) *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use: "faultinjection",
// 		Aliases: []string{"fi"},
// 		Short: "apply a fault injection rule",
// 		Long: `Fault injection rules are used to inject faults into requests in order to test for tolerance.`,
// 		PreRunE: func(cmd *cobra.Command, args []string) error {
// 			if err := surveyutils.SurveyFaultInjectionSpec(opts.Ctx, &opts.CreateRoutingRule); err != nil {
// 				return err
// 			}
// 			return nil
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			spec, err := faultInjectionConvertSpecFunc(opts.CreateRoutingRule.RoutingRuleSpec)
// 			if err != nil {
// 				return err
// 			}
// 			return applyRoutingRule(opts, spec)
// 		},
// 	}
//
// 	flagutils.AddFaultInjectionFlags(cmd.PersistentFlags(), &opts.CreateRoutingRule.RoutingRuleSpec)
//
// 	for _, v := range faultInjectionTypes {
// 		cmd.AddCommand(createFaultInjectionRule(v, opts))
// 	}
// 	return cmd
// }
//
// func createFaultInjectionRule(subCmd routingRuleSpecCommand, opts *options.Options) *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:     subCmd.use,
// 		Aliases: []string{subCmd.alias},
// 		Short:   subCmd.alias,
// 		Long:    subCmd.long,
// 	}
// 	if subCmd.addFlagsFunc != nil {
// 		subCmd.addFlagsFunc(cmd.PersistentFlags(), &opts.CreateRoutingRule.RoutingRuleSpec)
// 	}
// 	if subCmd.convertSpecFunc != nil {
// 		cmd.RunE = func (cmd *cobra.Command, args []string) error {
// 			spec, err := faultInjectionConvertSpecFunc(opts.CreateRoutingRule.RoutingRuleSpec)
// 			if err != nil {
// 				return err
// 			}
// 			return applyRoutingRule(opts, spec)
// 		}
// 	}
// 	for _, v := range subCmd.subCmds {
// 		cmd.AddCommand(createFaultInjectionRule(v, opts))
// 	}
// 	return cmd
// }
