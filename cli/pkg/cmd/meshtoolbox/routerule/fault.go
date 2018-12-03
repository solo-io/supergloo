package routerule

import (
	"fmt"
	"strconv"
	"strings"

	types "github.com/gogo/protobuf/types"
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"
	"github.com/solo-io/supergloo/cli/pkg/common"
	"github.com/solo-io/supergloo/cli/pkg/common/iutil"
	v1alpha3 "github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
)

const (
	HTTP              = "http"
	HTTP2             = "http2"
	GRPC              = "grpc"
	FIXED_DELAY       = "fixed delay"
	EXPONENTIAL_DELAY = "exponential delay"
)

var delayTypeOptions = []string{FIXED_DELAY, EXPONENTIAL_DELAY}
var errorTypeOptions = []string{HTTP, HTTP2, GRPC}

func EnsureFault(fo *options.InputFaultInjection, opts *options.Options) error {
	// initialize the output fields
	opts.MeshTool.RoutingRule.FaultInjection = &v1alpha3.HTTPFaultInjection{
		Delay: &v1alpha3.HTTPFaultInjection_Delay{},
		Abort: &v1alpha3.HTTPFaultInjection_Abort{},
	}
	// get a shorthand for the target
	t := opts.MeshTool.RoutingRule.FaultInjection

	// Delay
	if err := EnsurePercentage("Delay percentage (integer, 0-100)", &fo.DelayPercent, &(t.Delay).Percent, opts); err != nil {
		return err
	}
	if !opts.Top.Static {
		if err := iutil.ChooseFromList("Delay type", &fo.HttpDelayType, delayTypeOptions); err != nil {
			return err
		}
	}
	if !common.Contains(delayTypeOptions, fo.HttpDelayType) {
		return fmt.Errorf("Must specify a valid http delay type: %v", strings.Join(delayTypeOptions[:], ", "))
	}
	delayDuration := &types.Duration{}
	if err := EnsureDuration("Please specify the delay", &fo.HttpDelayValue,
		delayDuration,
		opts); err != nil {
		return err
	}
	if fo.HttpDelayType == FIXED_DELAY {
		t.Delay.HttpDelayType = &v1alpha3.HTTPFaultInjection_Delay_FixedDelay{
			FixedDelay: delayDuration,
		}
	} else {
		t.Delay.HttpDelayType = &v1alpha3.HTTPFaultInjection_Delay_ExponentialDelay{
			ExponentialDelay: delayDuration,
		}
	}

	// Abort
	if err := EnsurePercentage("Abort percentage (integer, 0-100)", &fo.AbortPercent, &(t.Abort).Percent, opts); err != nil {
		return err
	}

	if !opts.Top.Static {

		if err := iutil.ChooseFromList("Error type", &fo.ErrorType, errorTypeOptions); err != nil {
			return err
		}
		expectedErrorMessageType := "string"
		if fo.ErrorType == HTTP {
			expectedErrorMessageType = "int"
		}
		if err := iutil.GetStringInput(fmt.Sprintf("Error message (%v)", expectedErrorMessageType), &fo.ErrorMessage); err != nil {
			return err
		}
	}
	if !common.Contains(errorTypeOptions, fo.ErrorType) {
		return fmt.Errorf("Must specify a valid error type: %v", strings.Join(errorTypeOptions[:], ", "))
	}
	switch fo.ErrorType {
	case HTTP:
		eMessage, err := strconv.Atoi(fo.ErrorMessage)
		if err != nil {
			return err
		}
		t.Abort.ErrorType = &v1alpha3.HTTPFaultInjection_Abort_HttpStatus{
			HttpStatus: int32(eMessage),
		}
	case HTTP2:
		t.Abort.ErrorType = &v1alpha3.HTTPFaultInjection_Abort_Http2Error{
			Http2Error: fo.ErrorMessage,
		}
	case GRPC:
		t.Abort.ErrorType = &v1alpha3.HTTPFaultInjection_Abort_GrpcStatus{
			GrpcStatus: fo.ErrorMessage,
		}
	}

	return nil
}
