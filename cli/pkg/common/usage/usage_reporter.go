package usage

import (
	"runtime"
	"time"

	usageapi "github.com/solo-io/reporting-client/pkg/api/v1"
	usageclient "github.com/solo-io/reporting-client/pkg/client"
	"github.com/solo-io/reporting-client/pkg/signature"
	"github.com/solo-io/service-mesh-hub/pkg/container-runtime/version"
)

const (
	// reporting interval doesn't matter- we just want to send an initial report when the CLI starts to run
	UsageReportingInterval = time.Hour * 24

	// this is the url for a grpc service
	// note that grpc name resolution is a little different than for a normal HTTP/1.1 service
	// https://github.com/grpc/grpc/blob/master/doc/naming.md
	reportingServiceUrl = "reporting.corp.solo.io:443"
)

var (
	meshctlProduct = &usageapi.Product{
		Product: "meshctl",
		Version: version.Version,
		Arch:    runtime.GOARCH,
		Os:      runtime.GOOS,
	}
)

//go:generate mockgen -destination ./mocks/mock_usage_reporter.go -package usage_mocks github.com/solo-io/reporting-client/pkg/client Client

func DefaultUsageReporterProvider() usageclient.Client {
	return usageclient.NewUsageClient(
		reportingServiceUrl,
		&meshctlUsageReporter{},
		meshctlProduct,
		&signature.FileBackedSignatureManager{},
	)
}

type meshctlUsageReporter struct{}

var _ usageclient.UsagePayloadReader = &meshctlUsageReporter{}

func (m *meshctlUsageReporter) GetPayload() (map[string]string, error) {
	// TODO: decide what data is interesting to us
	return map[string]string{}, nil
}
