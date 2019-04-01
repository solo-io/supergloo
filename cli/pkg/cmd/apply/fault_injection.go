package apply

import (
	"net/http"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/cli/pkg/surveyutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func faultInjectionConvertSpecFunc(opts options.RoutingRuleSpec) (*v1.RoutingRuleSpec, error) {
	if opts.FaultInjection.Percent == 0 || opts.FaultInjection.Percent > 100 {
		return nil, errors.Errorf("invalid value %v: percentage must be (0-100)", opts.FaultInjection.Percent)
	}
	spec := &v1.RoutingRuleSpec{}
	faultInjection := &v1.FaultInjection{
		Percentage: opts.FaultInjection.Percent,
	}
	if opts.FaultInjection.Abort.Http.HttpStatus != 0 {
		code := opts.FaultInjection.Abort.Http.HttpStatus
		if http.StatusText(int(code)) == "" {
			return nil, errors.Errorf("invalid value %v: must be valid http status code", code)
		}
		faultInjection.FaultInjectionType = &v1.FaultInjection_Abort_{
			Abort: &v1.FaultInjection_Abort{
				ErrorType: &v1.FaultInjection_Abort_HttpStatus{
					HttpStatus: code,
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
				Duration:  opts.FaultInjection.Delay.Fixed,
				DelayType: v1.FaultInjection_Delay_FIXED,
			},
		}
		spec.RuleType = &v1.RoutingRuleSpec_FaultInjection{
			FaultInjection: faultInjection,
		}
		return spec, nil
	}

	return nil, errors.Errorf("no fault injection type specified")
}

var autoInteractive = func(opts *options.Options) {
	opts.Interactive = true
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
	mutateOpts:      autoInteractive,
}

var faultInjectionTypes = []routingRuleSpecCommand{
	{
		use:             "delay",
		alias:           "d",
		short:           "apply a delay type fault injection rule",
		addFlagsFunc:    flagutils.AddFaultInjectionDelayFlags,
		specSurveyFunc:  surveyutils.SurveyFaultInjectionDelay,
		convertSpecFunc: faultInjectionConvertSpecFunc,
		mutateOpts:      autoInteractive,
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
		use:             "abort",
		alias:           "a",
		short:           "apply an abort type fault injection rule",
		addFlagsFunc:    flagutils.AddFaultInjectionAbortFlags,
		specSurveyFunc:  surveyutils.SurveyFaultInjectionAbort,
		convertSpecFunc: faultInjectionConvertSpecFunc,
		mutateOpts:      autoInteractive,
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
