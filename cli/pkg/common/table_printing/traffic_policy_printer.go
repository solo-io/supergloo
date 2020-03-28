package table_printing

import (
	"fmt"
	"io"
	"strings"

	types2 "github.com/gogo/protobuf/types"
	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/cli/pkg/common/table_printing/internal"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
)

func NewTrafficPolicyPrinter(tableBuilder TableBuilder) TrafficPolicyPrinter {
	return &trafficPolicyPrinter{
		tableBuilder: tableBuilder,
	}
}

type trafficPolicyPrinter struct {
	tableBuilder TableBuilder
}

func (t *trafficPolicyPrinter) Print(out io.Writer, printMode PrintMode, trafficPolicies []*networking_v1alpha1.TrafficPolicy) error {
	table := t.tableBuilder(out)

	// will always have these two - will add more as they become relevant
	preFilteredHeaderRow := []string{
		"Name",
		"Source",
		"Destination",
		"Request Matchers",
		"Traffic Shift",
		"Fault Injection",
		"Timeout",
		"Retries",
		"CORS",
		"Mirror",
		"Header Manipulation",
	}

	var preFilteredRows [][]string
	for _, trafficPolicy := range trafficPolicies {
		newRow := []string{trafficPolicy.GetName()}
		tpSpec := trafficPolicy.Spec

		if tpSpec.GetSourceSelector() != nil && printMode == ServicePrintMode {
			newRow = append(newRow, internal.SelectorToCell(tpSpec.SourceSelector))
		} else {
			newRow = append(newRow, "")
		}

		if tpSpec.GetDestinationSelector() != nil && printMode == WorkloadPrintMode {
			newRow = append(newRow, internal.SelectorToCell(tpSpec.DestinationSelector))
		} else {
			newRow = append(newRow, "")
		}

		if len(tpSpec.GetHttpRequestMatchers()) > 0 {
			matchers, err := t.matchersToCell(tpSpec.GetHttpRequestMatchers())
			if err != nil {
				return err
			}
			newRow = append(newRow, matchers)
		} else {
			newRow = append(newRow, "")
		}

		if tpSpec.GetTrafficShift() != nil {
			newRow = append(newRow, t.trafficShiftToCell(tpSpec.TrafficShift))
		} else {
			newRow = append(newRow, "")
		}

		if tpSpec.GetFaultInjection() != nil {
			faultInjection, err := t.faultInjectionToCell(tpSpec.FaultInjection)
			if err != nil {
				return err
			}
			newRow = append(newRow, faultInjection)
		} else {
			newRow = append(newRow, "")
		}

		if tpSpec.GetRequestTimeout() != nil {
			newRow = append(newRow, t.requestTimeoutToCell(tpSpec.RequestTimeout))
		} else {
			newRow = append(newRow, "")
		}

		if tpSpec.GetRetries() != nil {
			newRow = append(newRow, t.retriesToCell(tpSpec.Retries))
		} else {
			newRow = append(newRow, "")
		}

		if tpSpec.GetCorsPolicy() != nil {
			newRow = append(newRow, t.corsToCell(tpSpec.CorsPolicy))
		} else {
			newRow = append(newRow, "")
		}

		if tpSpec.GetMirror() != nil {
			newRow = append(newRow, t.mirrorToCell(tpSpec.Mirror))
		} else {
			newRow = append(newRow, "")
		}

		if tpSpec.GetHeaderManipulation() != nil {
			newRow = append(newRow, t.headerManipulationToCell(tpSpec.HeaderManipulation))
		} else {
			newRow = append(newRow, "")
		}

		preFilteredRows = append(preFilteredRows, newRow)
	}

	filteredHeaders, filteredRows := internal.FilterEmptyColumns(preFilteredHeaderRow, preFilteredRows)

	table.SetHeader(filteredHeaders)
	table.AppendBulk(filteredRows)
	table.Render()
	return nil
}

func (t *trafficPolicyPrinter) headerManipulationToCell(headerManipulation *types.HeaderManipulation) string {
	str := ""
	if len(headerManipulation.RemoveResponseHeaders) > 0 {
		str += fmt.Sprintf("Remove response headers:\n%s\n\n", strings.Join(headerManipulation.RemoveResponseHeaders, "\n"))
	}

	if len(headerManipulation.AppendResponseHeaders) > 0 {
		var headers []string
		for h, v := range headerManipulation.AppendResponseHeaders {
			headers = append(headers, "%s: %s", h, v)
		}

		str += fmt.Sprintf("Append response headers:\n%s\n\n", strings.Join(headers, "\n"))
	}

	if len(headerManipulation.RemoveRequestHeaders) > 0 {
		str += fmt.Sprintf("Remove request headers:\n%s\n\n", strings.Join(headerManipulation.RemoveRequestHeaders, "\n"))
	}

	if len(headerManipulation.AppendRequestHeaders) > 0 {
		var headers []string
		for h, v := range headerManipulation.AppendRequestHeaders {
			headers = append(headers, "%s: %s", h, v)
		}

		str += fmt.Sprintf("Append request headers:\n%s\n\n", strings.Join(headers, "\n"))
	}

	return str
}

func (t *trafficPolicyPrinter) mirrorToCell(mirror *types.Mirror) string {
	return fmt.Sprintf("%.2f%% to:\nName: %s\nNamespace: %s\nCluster: %s",
		mirror.Percentage,
		mirror.Destination.GetName(),
		mirror.Destination.GetNamespace(),
		mirror.Destination.GetCluster(),
	)
}

func (t *trafficPolicyPrinter) corsToCell(corsPolicy *types.CorsPolicy) string {
	var corsString string
	if len(corsPolicy.GetAllowOrigins()) > 0 {
		origins := []string{}
		for _, origin := range corsPolicy.GetAllowOrigins() {
			if origin.GetExact() != "" {
				origins = append(origins, origin.GetExact())
			}
			if origin.GetPrefix() != "" {
				origins = append(origins, origin.GetPrefix()+"*")
			}
			if origin.GetRegex() != "" {
				origins = append(origins, origin.GetRegex()+" (Regex)")
			}
		}

		corsString += fmt.Sprintf("Allowed origins:\n%s\n\n", strings.Join(origins, "\n"))
	}

	if len(corsPolicy.AllowMethods) > 0 {
		corsString += fmt.Sprintf("Allowed methods:\n%s\n\n", strings.Join(corsPolicy.AllowMethods, "\n"))
	}

	if len(corsPolicy.AllowHeaders) > 0 {
		corsString += fmt.Sprintf("Allowed headers:\n%s\n\n", strings.Join(corsPolicy.AllowHeaders, "\n"))
	}

	if len(corsPolicy.ExposeHeaders) > 0 {
		corsString += fmt.Sprintf("Expose headers:\n%s\n\n", strings.Join(corsPolicy.ExposeHeaders, "\n"))
	}

	if corsPolicy.GetMaxAge() != nil {
		corsString += fmt.Sprintf("Max age: %ds %dns\n\n", corsPolicy.GetMaxAge().GetSeconds(), corsPolicy.GetMaxAge().GetNanos())
	}

	if corsPolicy.GetAllowCredentials() != nil {
		corsString += fmt.Sprintf("Allow credentials: %t\n\n", corsPolicy.GetAllowCredentials().GetValue())
	}

	return corsString
}

func (t *trafficPolicyPrinter) retriesToCell(retries *types.RetryPolicy) string {
	str := fmt.Sprintf("%d attempts", retries.Attempts)
	if retries.GetPerTryTimeout() != nil {
		str += fmt.Sprintf(" with per-try timeout: %ds %dns", retries.PerTryTimeout.Seconds, retries.PerTryTimeout.Nanos)
	}

	return str
}

func (t *trafficPolicyPrinter) requestTimeoutToCell(requestTimeout *types2.Duration) string {
	return fmt.Sprintf("%ds %dns", requestTimeout.Seconds, requestTimeout.Nanos)
}

func (t *trafficPolicyPrinter) faultInjectionToCell(faultInjection *types.FaultInjection) (string, error) {
	var injectionType string
	switch faultInjection.GetFaultInjectionType().(type) {
	case *types.FaultInjection_Delay_:
		delay := faultInjection.GetDelay()
		if delay.GetFixedDelay() != nil {
			injectionType = fmt.Sprintf("Delay (Fixed: %ds %dns)", delay.GetFixedDelay().Seconds, delay.GetFixedDelay().Nanos)
		} else if delay.GetExponentialDelay() != nil {
			injectionType = fmt.Sprintf("Delay (Exponential: %ds %dns)", delay.GetFixedDelay().Seconds, delay.GetFixedDelay().Nanos)
		} else {
			injectionType = "Delay"
		}
	case *types.FaultInjection_Abort_:
		abort := faultInjection.GetAbort()
		if abort.GetHttpStatus() != 0 {
			injectionType = fmt.Sprintf("Abort with HTTP %d", abort.GetHttpStatus())
		} else {
			injectionType = "Abort"
		}
	default:
		return "", eris.Errorf("Unhandled fault injection type: %v", faultInjection)
	}

	return fmt.Sprintf("%s\nOn %.2f%% requests", injectionType, faultInjection.GetPercentage()), nil
}

func (t *trafficPolicyPrinter) matchersToCell(matchers []*types.HttpMatcher) (string, error) {
	var matcherStrings []string
	for _, matcher := range matchers {
		var headerStrings []string
		for _, headerMatch := range matcher.Headers {
			headerStrings = append(headerStrings, fmt.Sprintf(
				"%s: %s\n(Regex: %t, Invert: %t)",
				headerMatch.Name,
				headerMatch.Value,
				headerMatch.Regex,
				headerMatch.InvertMatch,
			))
		}

		var pathSpecifier string
		if matcher.GetPathSpecifier() != nil {
			switch matcher.GetPathSpecifier().(type) {
			case *types.HttpMatcher_Prefix:
				pathSpecifier = matcher.GetPrefix() + "*"
			case *types.HttpMatcher_Exact:
				pathSpecifier = matcher.GetExact()
			case *types.HttpMatcher_Regex:
				pathSpecifier = matcher.GetExact()
			default:
				return "", eris.Errorf("Unhandled matcher path specifier type: %+v", matcher)
			}
		}

		var queryParams []string
		for _, queryParam := range matcher.GetQueryParameters() {
			value := queryParam.Value
			if queryParam.Regex {
				value += "(Regex)"
			}
			queryParams = append(queryParams, fmt.Sprintf(
				"%s=%s",
				queryParam.Name,
				value,
			))
		}

		matcherStrings = append(matcherStrings, fmt.Sprintf(
			"%s %s\n\nQuery Params:\n\n%s\n\nHeaders:\n\n%s\n",
			matcher.GetMethod().GetMethod().String(),
			pathSpecifier,
			strings.Join(queryParams, "\n"),
			strings.Join(headerStrings, "\n"),
		))
	}

	return strings.Join(matcherStrings, "\n--------------\n"), nil
}

func (t *trafficPolicyPrinter) trafficShiftToCell(trafficShift *types.MultiDestination) string {
	var destinations []string
	for _, destination := range trafficShift.Destinations {
		var subsets []string
		for k, v := range destination.Subset {
			subsets = append(subsets, fmt.Sprintf("%s: %s", k, v))
		}
		var subsetsStringRepr string
		if len(subsets) > 0 {
			subsetsStringRepr = fmt.Sprintf("Subsets: %s\n", strings.Join(subsets, ", "))
		}

		var weightStringRepr string
		if destination.Weight > 0 && len(trafficShift.Destinations) > 0 {
			weightStringRepr = fmt.Sprintf("Weight: %d\n", destination.Weight)
		}
		destinations = append(destinations, fmt.Sprintf("Destination:\nName: %s\nNamespace: %s\nCluster: %s\n%s%s\n",
			destination.GetDestination().GetName(),
			destination.GetDestination().GetNamespace(),
			destination.GetDestination().GetCluster(),
			subsetsStringRepr,
			weightStringRepr,
		))
	}

	return strings.Join(destinations, "\n")
}
