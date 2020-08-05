package internal

import (
	"context"
	"net/url"

	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/check/internal/components"
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
	ErrorMessage string

	// optionally provide a link to a docs page that a user should consult to resolve the error
	DocsLink *url.URL

	// an optional suggestion for a next action for the user to take for resolving a failed check
	Hint string
}

var (
	// TODO implement kube connectivity check

	managementPlane = Category{
		Name: "Service Mesh Hub Management Plane",
		Checks: []Check{
			components.NewComponentsCheck(),
		},
	}

	categories = []Category{
		managementPlane,
	}
)

func RunChecks(ctx context.Context, client client.Client, installNamespace string) error {
	for _, category := range categories {
		for _, check := range category.Checks {
			failure := check.Run(ctx, client, installNamespace)
		}
	}
	return nil
}

func print
