package status

import (
	"context"
	"io"

	healthcheck_types "github.com/solo-io/service-mesh-hub/cli/pkg/tree/check/healthcheck/types"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type StatusClient interface {
	Check(ctx context.Context, installNamespace string, healthCheckSuite healthcheck_types.HealthCheckSuite) *StatusReport
}

type StatusPrinter interface {
	Print(out io.Writer, report *StatusReport)
}

type StatusReport struct {
	Results map[healthcheck_types.Category][]*HealthCheckResult

	// represents overall success of the status report
	Success bool
}

// used for displaying to the user - note, can get rendered out to JSON
type HealthCheckResult struct {
	// description is the check being performed
	Description string

	// did this particular health check succeed
	Success bool

	// the following three fields should only be set if Success is false
	Message  string
	DocsLink string
	Hint     string
}
