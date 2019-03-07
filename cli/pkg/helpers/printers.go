package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/gogo/protobuf/proto"

	"github.com/olekukonko/tablewriter"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/protoutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func PrintInstalls(list v1.InstallList, outputType string) {
	cliutils.PrintList(outputType, "", list,
		func(data interface{}, w io.Writer) error {
			tablePrintInstalls(data.(v1.InstallList), w)
			return nil
		}, os.Stdout)
}

func tablePrintInstalls(list v1.InstallList, w io.Writer) {

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Install", "type", "status", "details"})

	for _, install := range list {
		name := install.GetMetadata().Name
		s := install.Status.State.String()
		t := installType(install)

		details := installDetails(install)
		if len(details) == 0 {
			details = []string{""}
		}
		for i, line := range details {
			if i == 0 {
				table.Append([]string{name, t, s, line})
			} else {
				table.Append([]string{"", "", "", line})
			}
		}

	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}

func installType(in *v1.Install) string {
	switch in.InstallType.(type) {
	case *v1.Install_Istio_:
		return "Istio"
	}
	return "Unknown"
}

func installDetails(in *v1.Install) []string {
	var details []string
	add := func(s ...string) {
		details = append(details, s...)
	}
	add(
		fmt.Sprintf("enabled: %v", !in.Disabled),
	)
	switch t := in.InstallType.(type) {
	case *v1.Install_Istio_:
		add(
			fmt.Sprintf("version: %v", t.Istio.IstioVersion),
			fmt.Sprintf("namespace: %v", t.Istio.InstallationNamespace),
			fmt.Sprintf("mtls enabled: %v", t.Istio.EnableMtls),
			fmt.Sprintf("auto inject enabled: %v", t.Istio.EnableAutoInject),
		)
		if t.Istio.CustomRootCert != nil {
			add(
				fmt.Sprintf("mtls enabled: %v", t.Istio.CustomRootCert),
			)
		}
		add(
			fmt.Sprintf("grafana enabled: %v", t.Istio.InstallGrafana),
			fmt.Sprintf("prometheus enabled: %v", t.Istio.InstallPrometheus),
			fmt.Sprintf("jaeger enabled: %v", t.Istio.InstallJaeger),
		)
	}
	return details
}

func PrintRoutingRules(list v1.RoutingRuleList, outputType string) {
	cliutils.PrintList(outputType, "", list,
		func(data interface{}, w io.Writer) error {
			tablePrintRoutingRules(data.(v1.RoutingRuleList), w)
			return nil
		}, os.Stdout)
}

func tablePrintRoutingRules(list v1.RoutingRuleList, w io.Writer) {

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"RoutingRule", "type", "status", "sources", "destinations", "spec"})

	for _, routingRule := range list {
		name := routingRule.GetMetadata().Name
		s := routingRule.Status.State.String()
		t := routingRuleType(routingRule)

		details := routingRuleDetails(routingRule)
		if len(details) == 0 {
			details = []string{""}
		}
		for i, line := range details {
			if i == 0 {
				table.Append([]string{name, t, s, selector(routingRule.SourceSelector), selector(routingRule.DestinationSelector), line})
			} else {
				table.Append([]string{"", "", "", "", "", line})
			}
		}

	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}

func MustMarshal(v interface{}) string {
	jsn, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return string(jsn)
}

func MustMarshalProto(v proto.Message) string {
	jsn, err := protoutils.MarshalBytes(v)
	if err != nil {
		panic(err)
	}
	return string(jsn)
}

func selector(in *v1.PodSelector) string {
	if in == nil {
		return "all"
	}
	switch t := in.SelectorType.(type) {
	case *v1.PodSelector_LabelSelector_:
		return fmt.Sprintf("labels: %v", MustMarshal(t.LabelSelector.LabelsToMatch))
	case *v1.PodSelector_UpstreamSelector_:
		return fmt.Sprintf("upstreams: %v", MustMarshal(t.UpstreamSelector.Upstreams))
	case *v1.PodSelector_NamespaceSelector_:
		return fmt.Sprintf("namespaces: %v", MustMarshal(t.NamespaceSelector.Namespaces))
	}
	return "Unknown"
}

func routingRuleType(in *v1.RoutingRule) string {
	switch in.Spec.RuleType.(type) {
	case *v1.RoutingRuleSpec_TrafficShifting:
		return "TrafficShifting"
	}
	return "Unknown"
}

func routingRuleDetails(in *v1.RoutingRule) []string {
	var details []string
	add := func(s ...string) {
		details = append(details, s...)
	}
	switch t := in.Spec.RuleType.(type) {
	case *v1.RoutingRuleSpec_TrafficShifting:
		add(
			"traffic shifting: ",
		)
		for _, dest := range t.TrafficShifting.Destinations.Destinations {
			add(
				fmt.Sprintf("- %v", dest.Destination.Upstream),
				fmt.Sprintf("  weight: %v", dest.Weight),
			)
		}
	}
	return details
}
