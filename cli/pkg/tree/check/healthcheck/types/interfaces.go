package healthcheck_types

import (
	"context"
	"net/url"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	kubernetes_core "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/core"
	kubernetes_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/kubernetes/discovery"
)

type Category struct {
	Name   string
	Weight int // higher weight should take higher precedence- ie, categories with a higher weight should be run first
}

type HealthCheckSuite map[Category][]HealthCheck

type HealthCheck interface {
	// this should be a human-readable description of what we're looking for
	// should be nonempty
	GetDescription() string

	// if Run returns (_, false), then this check determined that itself was not applicable, and no result (success or failure) should be reported
	//   (example: the federation health checks should mark themselves as N/A if no federation has been configured yet)
	// if Run returns (runFailure, true) and runFailure is non-nil, then the result's message should be reported to the user
	// if Run returns (nil, true) then the check succeeded
	Run(ctx context.Context, installNamesapce string, clients Clients) (runFailure *RunFailure, checkApplies bool)
}

type RunFailure struct {
	// a human-readable summary of went wrong
	// should be non-empty
	ErrorMessage string

	// optionally provide a link to a docs page that a user should consult to resolve the error
	// can be nil
	DocsLink *url.URL

	// a suggestion for a next action for the user to take
	// can be empty
	Hint string
}

// clients that will be passed to every health check to use
type Clients struct {
	NamespaceClient     kubernetes_core.NamespaceClient
	ServerVersionClient kubernetes_discovery.ServerVersionClient
	PodClient           kubernetes_core.PodClient
	MeshServiceClient   zephyr_discovery.MeshServiceClient
}
