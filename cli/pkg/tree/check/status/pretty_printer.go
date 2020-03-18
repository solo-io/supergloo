package status

import (
	"fmt"
	"io"
	"sort"
	"strings"

	healthcheck_types "github.com/solo-io/mesh-projects/cli/pkg/tree/check/healthcheck/types"
)

var (
	checkMarkEmoji = "\u2705"
	redXEmoji      = "\u274C"
)

type PrettyPrinter StatusPrinter

func NewPrettyPrinter() PrettyPrinter {
	return &prettyPrinter{}
}

type prettyPrinter struct{}

func (p *prettyPrinter) Print(out io.Writer, statusReport *StatusReport) {
	var sortedCategories healthcheck_types.SortableCategories
	for category, _ := range statusReport.Results {
		sortedCategories = append(sortedCategories, category)
	}

	// needed to produce deterministic output so that we can test against it
	sort.Sort(sortedCategories)

	for _, category := range sortedCategories {
		results := statusReport.Results[category]

		allPassed := true
		for _, result := range results {
			allPassed = allPassed && result.Success
		}

		if allPassed {
			fmt.Fprintf(out, successMessage(category.Name+"\n"))
		} else {
			fmt.Fprintf(out, failedMessage(category.Name+"\n"))
		}

		// + 3 is length of the unicode character plus a space before the actual category name
		fmt.Fprintf(out, strings.Repeat("-", len(category.Name)+3)+"\n")

		for _, result := range results {
			printer := successMessage
			if !result.Success {
				printer = failedMessage
			}

			fmt.Fprintf(out, printer(result.Description)+"\n")
			if !result.Success {
				if result.Message != "" {
					fmt.Fprintf(out, "\tMessage: %s\n", result.Message)
				}
				if result.Hint != "" {
					fmt.Fprintf(out, "\tHint: %s\n", result.Hint)
				}
				if result.DocsLink != "" {
					fmt.Fprintf(out, "\tDocs: %s\n", result.DocsLink)
				}
			}
		}

		fmt.Fprintf(out, "\n\n")
	}
	if statusReport.Success {
		fmt.Fprintf(out, successMessage("Service Mesh Hub check found no errors\n"))
	} else {
		fmt.Fprintf(out, failedMessage("Service Mesh Hub check found errors\n"))
	}
}

func successMessage(message string) string {
	return fmt.Sprintf("%s %s", checkMarkEmoji, message)
}

func failedMessage(message string) string {
	return fmt.Sprintf("%s %s", redXEmoji, message)
}
