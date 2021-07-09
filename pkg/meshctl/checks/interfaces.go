package checks

import (
	"context"
	"net/url"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Stage string

var (
	PreInstall  Stage = "pre-install"
	PostInstall       = "post-install"
	PreUpgrade        = "pre-upgrade"
	PostUpgrade       = "post-upgrade"
	Test              = "test"
)

type Component string

var (
	Server Component = "server"
	Agent            = "agent"
)

type Environment struct {
	MetricsPort uint32
	Namespace   string
	InCluster   bool
}

type CheckContext interface {
	Environment() Environment
	Client() client.Client
	AccessMgmtServerAdminPort(ctx context.Context, op func(ctx context.Context, addr string) (error, string)) (error, string)
}

type Check interface {
	// description of what is being checked
	GetDescription() string

	// Execute the check, pass in the namespace that Gloo Mesh is installed in
	Run(ctx context.Context, checkCtx CheckContext) *Failure
}

type Category struct {
	Name   string
	Checks []Check
}

type Failure struct {
	// user-facing error message describing failed check
	Errors []error

	// an optional suggestion for a next action for the user to take for resolving a failed check
	Hints []Hint
}

func (f *Failure) AddError(err ...error) *Failure {
	f.Errors = append(f.Errors, err...)
	return f
}

func (f *Failure) OrNil() *Failure {
	if len(f.Errors) == 0 {
		return nil
	}
	return f
}

func (f *Failure) AddHint(h string, d *url.URL) *Failure {
	if h != "" || d != nil {
		f.Hints = append(f.Hints, Hint{Hint: h, DocsLink: d})
	}
	return f
}

type Hint struct {
	// an optional suggestion for a next action for the user to take for resolving a failed check
	Hint string
	// optionally provide a link to a docs page that a user should consult to resolve the error
	DocsLink *url.URL
}
