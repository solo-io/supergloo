package internal

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Category struct {
	Name   string
	Checks []Check
}

type Check interface {
	// description of what is being checked
	GetDescription() string

	// Execute the check, pass in the namespace that Service Mesh Hub is installed in
	Run(ctx context.Context, client client.Client, installNamespace string) *Failure
}

type Failure struct {
	// user-facing error message describing failed check
	Errors []error

	// optionally provide a link to a docs page that a user should consult to resolve the error
	DocsLink *url.URL

	// an optional suggestion for a next action for the user to take for resolving a failed check
	Hint string
}

var (
	checkMarkChar = "\u2705"
	redXChar      = "\u274C"

	// TODO implement kube connectivity check

	managementPlane = Category{
		Name: "Service Mesh Hub",
		Checks: []Check{
			NewDeploymentsCheck(),
		},
	}

	configuration = Category{
		Name: "Management Configuration",
		Checks: []Check{
			NewNetworkingCrdCheck(),
		},
	}

	categories = []Category{
		managementPlane,
		configuration,
	}
)

func RunChecks(ctx context.Context, client client.Client, installNamespace string) error {
	for _, category := range categories {
		fmt.Println(category.Name)
		fmt.Printf(strings.Repeat("-", len(category.Name)+3) + "\n")
		for _, check := range category.Checks {
			failure := check.Run(ctx, client, installNamespace)
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
	} else {
		fmt.Printf("%s %s\n", checkMarkChar, description)
	}
}
