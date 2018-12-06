package printers

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/utils/cliutils"
	"github.com/solo-io/supergloo/cli/pkg/common"
	gloov1 "github.com/solo-io/supergloo/pkg/api/external/gloo/v1"
	"github.com/solo-io/supergloo/pkg/api/v1"
)

func RoutingRuleTable(list *v1.RoutingRuleList, output string, template string) error {
	err := cliutils.PrintList(output, template, list,
		func(data interface{}, w io.Writer) error {
			routingRuleTable(data.(*v1.RoutingRuleList), w)
			return nil
		},
		os.Stdout)
	return err
}

func routingRuleTable(list *v1.RoutingRuleList, w io.Writer) {
	table := tablewriter.NewWriter(w)
	headers := []string{"", "name", "target-mesh", "sources", "destintations", "matchers"}
	table.SetHeader(headers)

	table.SetBorder(false)

	for i, v := range *list {
		table.Append(transformRoutingRule(v, i+1))
	}

	table.Render()
}

func transformRoutingRule(routingRule *v1.RoutingRule, index int) []string {
	name := routingRule.Metadata.Name
	targetMesh := routingRule.TargetMesh.Name
	sources := getUpstreams(routingRule.Sources)
	destinations := getUpstreams(routingRule.Destinations)
	matchers := getMatchers(routingRule.RequestMatchers)
	row := []string{strconv.Itoa(index), name, targetMesh, sources, destinations, matchers}
	return row
}

func getUpstreams(refs []*core.ResourceRef) string {
	var b strings.Builder
	for i, ref := range refs {
		fmt.Fprintf(&b, "%s%s%s", ref.Namespace, common.NamespacedResourceSeparator, ref.Name)
		// Add separator, except for last entry
		if i != len(refs)-1 {
			b.WriteString(common.ListOptionSeparator)
		}
	}
	return b.String()
}

func getMatchers(matchers []*gloov1.Matcher) string {
	result := make([]string, len(matchers))
	for i, m := range matchers {

		clauses := make([]string, 0)
		if m.PathSpecifier != nil {
			switch specifier := m.PathSpecifier.(type) {
			case *gloov1.Matcher_Prefix:
				clauses = append(clauses, fmt.Sprintf("prefix=%s", specifier.Prefix))
			default:
				//TODO: ignore path specifiers that we currently don't support
			}
		}

		if m.Methods != nil {
			clauses = append(clauses, fmt.Sprintf("methods=%s", strings.Join(m.Methods, common.SubListOptionSeparator)))
		}

		result[i] = strings.Join(clauses, common.ListOptionSeparator)
	}
	return strings.Join(result, " && ")
}
