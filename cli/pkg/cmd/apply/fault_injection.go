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
