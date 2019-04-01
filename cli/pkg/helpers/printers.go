package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"

	"github.com/gogo/protobuf/proto"

	"github.com/olekukonko/tablewriter"
	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/protoutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func PrintInstalls(list v1.InstallList, outputType string) {
	_ = cliutils.PrintList(outputType, "", list,
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
	switch installType := in.InstallType.(type) {
	case *v1.Install_Mesh:
		switch installType.Mesh.MeshInstallType.(type) {
		case *v1.MeshInstall_IstioMesh:
			return "Istio Mesh"
		}
	case *v1.Install_Ingress:
		switch installType.Ingress.IngressInstallType.(type) {
		case *v1.MeshIngressInstall_Gloo:
			return "Gloo Ingress"
		}
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

	switch installType := in.InstallType.(type) {
	case *v1.Install_Mesh:
		switch meshType := installType.Mesh.MeshInstallType.(type) {
		case *v1.MeshInstall_IstioMesh:
			add(
				fmt.Sprintf("version: %v", meshType.IstioMesh.IstioVersion),
				fmt.Sprintf("namespace: %v", in.InstallationNamespace),
				fmt.Sprintf("mtls enabled: %v", meshType.IstioMesh.EnableMtls),
				fmt.Sprintf("auto inject enabled: %v", meshType.IstioMesh.EnableAutoInject),
			)
			if meshType.IstioMesh.CustomRootCert != nil {
				add(
					fmt.Sprintf("mtls enabled: %v", meshType.IstioMesh.CustomRootCert),
				)
			}
			add(
				fmt.Sprintf("grafana enabled: %v", meshType.IstioMesh.InstallGrafana),
				fmt.Sprintf("prometheus enabled: %v", meshType.IstioMesh.InstallPrometheus),
				fmt.Sprintf("jaeger enabled: %v", meshType.IstioMesh.InstallJaeger),
			)
		}
	case *v1.Install_Ingress:
		switch installType.Ingress.IngressInstallType.(type) {
		case *v1.MeshIngressInstall_Gloo:
		}
	}
	return details
}

func PrintRoutingRules(list v1.RoutingRuleList, outputType string) {
	_ = cliutils.PrintList(outputType, "", list,
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

func PrintSecurityRules(list v1.SecurityRuleList, outputType string) {
	_ = cliutils.PrintList(outputType, "", list,
		func(data interface{}, w io.Writer) error {
			tablePrintSecurityRules(data.(v1.SecurityRuleList), w)
			return nil
		}, os.Stdout)
}

func tablePrintSecurityRules(list v1.SecurityRuleList, w io.Writer) {

	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"SecurityRule", "status", "paths", "methods", "sources", "destinations"})

	for _, securityRule := range list {
		name := securityRule.GetMetadata().Name
		stat := securityRule.Status.State.String()
		paths := strings.Join(securityRule.AllowedPaths, ",")
		methods := strings.Join(securityRule.AllowedMethods, ",")
		table.Append([]string{name, stat, paths, methods, selector(securityRule.SourceSelector), selector(securityRule.DestinationSelector)})
	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}

func PrintTlsSecrets(list v1.TlsSecretList, outputType string) {
	_ = cliutils.PrintList(outputType, "", list,
		func(data interface{}, w io.Writer) error {
			tablePrintTlsSecrets(data.(v1.TlsSecretList), w)
			return nil
		}, os.Stdout)
}

func tablePrintTlsSecrets(list v1.TlsSecretList, w io.Writer) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"TlsSecret"})

	for _, tlsSecret := range list {
		name := tlsSecret.GetMetadata().Name
		table.Append([]string{name})
	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}

func PrintSecrets(list gloov1.SecretList, outputType string) {
	_ = cliutils.PrintList(outputType, "", list,
		func(data interface{}, w io.Writer) error {
			tablePrintSecrets(data.(gloov1.SecretList), w)
			return nil
		}, os.Stdout)
}

func tablePrintSecrets(list gloov1.SecretList, w io.Writer) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Secrets"})

	for _, secret := range list {
		name := secret.GetMetadata().Name
		table.Append([]string{name})
	}

	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}

func PrintMeshes(list v1.MeshList, outputType string) {
	_ = cliutils.PrintList(outputType, "", list,
		func(data interface{}, w io.Writer) error {
			tablePrintMeshes(data.(v1.MeshList), w)
			return nil
		}, os.Stdout)
}

func tablePrintMeshes(list v1.MeshList, w io.Writer) {
	table := tablewriter.NewWriter(w)
	table.SetHeader([]string{"Mesh", "Type", "mTLS"})

	for _, mesh := range list {
		name := mesh.GetMetadata().Name
		meshType := func() string {
			switch mesh.MeshType.(type) {
			case *v1.Mesh_Istio:
				return "Istio"
			}
			return "unknown"
		}()
		mtls := fmt.Sprintf("%v", mesh.MtlsConfig != nil && mesh.MtlsConfig.MtlsEnabled)

		table.Append([]string{name, meshType, mtls})
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
	case *v1.RoutingRuleSpec_FaultInjection:
		return "FaultInjection"
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
	case *v1.RoutingRuleSpec_FaultInjection:
		add(
			"fault injection: ",
		)
		switch faultType := t.FaultInjection.FaultInjectionType.(type) {
		case *v1.FaultInjection_Delay_:
			switch faultType.Delay.DelayType {
			case v1.FaultInjection_Delay_fixed:
				add(
					"- fixed delay",
					fmt.Sprintf("%v", faultType.Delay.Duration),
				)
			}
		case *v1.FaultInjection_Abort_:
			switch faultAbortType := faultType.Abort.ErrorType.(type) {
			case *v1.FaultInjection_Abort_HttpStatus:
				add(
					"- http status abort",
					fmt.Sprintf("%v", faultAbortType.HttpStatus),
				)
			}
		}

	}
	return details
}
