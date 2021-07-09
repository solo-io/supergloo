package checks

import (
	"context"
	"fmt"
	"strings"
)

var (
	checkMarkChar = "✅"
	redXChar      = "❌"
)

func constructChecks() []Category {
	managementPlane := Category{
		Name: "Gloo Mesh",
		Checks: []Check{
			NewDeploymentsCheck(),
			NewEnterpriseRegistrationCheck(),
		},
	}

	configuration := Category{
		Name: "Management Configuration",
		Checks: []Check{
			NewNetworkingCrdCheck(),
		},
	}

	return []Category{
		managementPlane,
		configuration,
	}
}

func RunChecks(ctx context.Context, checkCtx CheckContext) error {
	for _, category := range constructChecks() {
		fmt.Println(category.Name)
		fmt.Printf(strings.Repeat("-", len(category.Name)+3) + "\n")
		for _, check := range category.Checks {
			failure := check.Run(ctx, checkCtx)
			printResult(failure, check.GetDescription())
		}
		fmt.Println()
	}
	return nil
}

func printResult(failure *Failure, description string) {
	if failure != nil {
		fmt.Printf("%s %s\n", redXChar, description)
		for _, err := range failure.Errors {
			fmt.Printf("  - %s\n", err.Error())
		}
		if len(failure.Hints) > 0 {
			fmt.Printf("Hints:\n")
			for _, hint := range failure.Hints {
				fmt.Printf("  %s", hint.Hint)
				if hint.DocsLink != nil {
					fmt.Printf(" (For more info, see: %s)", hint.DocsLink.String())
				}
				fmt.Println()
			}
		}
	} else {
		fmt.Printf("%s %s\n", checkMarkChar, description)
	}
}
