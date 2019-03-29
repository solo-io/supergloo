package surveyutils

import (
	"context"
	"strconv"
	"time"

	"github.com/solo-io/gloo/pkg/cliutil"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/cli/pkg/options"
)

func SurveyFaultInjectionPercent(ctx context.Context, in *options.CreateRoutingRule) error {
	var percent uint32
	if err := cliutil.GetUint32Input("percentage of requests to inject (0-100)", &percent); err != nil {
		return errors.Wrapf(err, "error getting percentage value")
	}

	floatPercent := float64(percent)
	if floatPercent > 100 {
		return errors.Errorf("invalid value %v, percentage value must be between 0-100", floatPercent)
	}

	in.RoutingRuleSpec.FaultInjection.Percent = floatPercent
	return nil
}

const (
	delay = "delay"
	abort = "abort"

	fixed = "fixed"

	http = "http"
)

var faultInjectionOptions = []string{delay, abort}
var delayOptions = []string{fixed}
var abortOptions = []string{http}

func surveyFaultInjectionAbort(ctx context.Context, in *options.CreateRoutingRule) error {
	var abortType string
	if err := cliutil.ChooseFromList("select type of abort rule", &abortType, abortOptions); err != nil {
		return errors.Wrapf(err, "error selecting abort type")
	}
	switch abortType {
	case http:
		if err := SurveyFaultInjectionAbortHttp(ctx, in); err != nil {
			return err
		}
	default:
		return errors.Errorf("could not determine type of delay fault injection rule %s", abortType)
	}

	return nil
}

func SurveyFaultInjectionAbortHttp(ctx context.Context, in *options.CreateRoutingRule) error {
	var httpAbort string
	if err := cliutil.GetStringInput("enter status code to abort request with (valid http status code)", &httpAbort); err != nil {
		return errors.Wrapf(err, "unable to read abort value")
	}

	code, err := strconv.Atoi(httpAbort)
	if err != nil {
		return errors.Wrapf(err, "value was not a valid integer")
	}
	in.RoutingRuleSpec.FaultInjection.Abort.Http.HttpStatus = int32(code)
	return nil
}

func surveyFaultInjectionDelay(ctx context.Context, in *options.CreateRoutingRule) error {
	var delayChoice string
	if err := cliutil.ChooseFromList("select type of delay rule", &delayChoice, delayOptions); err != nil {
		return errors.Wrapf(err, "error selecting delay type")
	}

	switch delayChoice {
	case fixed:
		if err := SurveyFaultInjectionDelayFixed(ctx, in); err != nil {
			return err
		}
	default:
		return errors.Errorf("could not determine type of abort fault injection rule %s", delayChoice)
	}
	return nil
}

func SurveyFaultInjectionDelayFixed(ctx context.Context, in *options.CreateRoutingRule) error {
	var fixedDelay string
	if err := cliutil.GetStringInput("enter fixed delay duration", &fixedDelay); err != nil {
		return errors.Wrapf(err, "unable to read duration value")
	}
	dur, err := time.ParseDuration(fixedDelay)
	if err != nil {
		return errors.Wrapf(err, "invalid %v value passed for duration", fixedDelay)
	}
	in.RoutingRuleSpec.FaultInjection.Delay.Fixed = dur
	return nil
}

func SurveyFaultInjectionSpec(ctx context.Context, in *options.CreateRoutingRule) error {
	var choice string
	if err := cliutil.ChooseFromList("select type of fault injection rule", &choice, faultInjectionOptions); err != nil {
		return errors.Wrapf(err, "error selecting fault injection type")
	}

	switch choice {
	case delay:
		if err := surveyFaultInjectionDelay(ctx, in); err != nil {
			return err
		}
	case abort:
		if err := surveyFaultInjectionAbort(ctx, in); err != nil {
			return err
		}
	default:
		return errors.Errorf("could not determine type of fault injection rule %s", choice)
	}

	if err := SurveyFaultInjectionPercent(ctx, in); err != nil {
		return err
	}

	return nil
}
