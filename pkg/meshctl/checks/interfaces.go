package checks

import (
	"context"
	"net/url"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Check interface {
	// description of what is being checked
	GetDescription() string

	// Execute the check, pass in the namespace that Gloo Mesh is installed in
	Run(ctx context.Context, client client.Client, installNamespace string) *Failure
}

type Category struct {
	Name   string
	Checks []Check
}

type Failure struct {
	// user-facing error message describing failed check
	Errors []error

	// optionally provide a link to a docs page that a user should consult to resolve the error
	DocsLink *url.URL

	// an optional suggestion for a next action for the user to take for resolving a failed check
	Hint string
}
