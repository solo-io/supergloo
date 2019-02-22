package helpers

import (
	"fmt"
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/solo-io/go-utils/cliutils"
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
