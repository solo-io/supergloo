package status

import (
	"context"

	healthcheck_types "github.com/solo-io/mesh-projects/cli/pkg/tree/check/healthcheck/types"
)

type StatusClientFactory func(healthCheckClients healthcheck_types.Clients) StatusClient

func StatusClientFactoryProvider() StatusClientFactory {
	return NewStatusClient
}

func NewStatusClient(healthCheckClients healthcheck_types.Clients) StatusClient {
	return &statusClient{
		healthCheckClients: healthCheckClients,
	}
}

type statusClient struct {
	healthCheckClients healthcheck_types.Clients
}

func (s *statusClient) Check(ctx context.Context, installNamespace string, healthCheckSuite healthcheck_types.HealthCheckSuite) *StatusReport {
	statusReport := &StatusReport{
		Results: map[healthcheck_types.Category][]*HealthCheckResult{},
		Success: true,
	}

	// we don't want to report a ton of errors back to the user and overwhelm them
	// report on a single failure at a time
	foundErrorInSuite := false

	healthCheckSuite.ForEachCategoryInWeightOrder(func(category healthcheck_types.Category) {
		if foundErrorInSuite {
			// do nothing- we should report on the error we already found
			return
		}

		results := statusReport.Results[category]

		for _, check := range healthCheckSuite[category] {
			runFailure, checkApplies := check.Run(ctx, installNamespace, s.healthCheckClients)
			if !checkApplies {
				continue
			}

			foundErrorInSuite = runFailure != nil

			result := &HealthCheckResult{
				Description: check.GetDescription(),
				Success:     !foundErrorInSuite,
			}
			if runFailure != nil {
				statusReport.Success = false

				result.Message = runFailure.ErrorMessage
				result.Hint = runFailure.Hint

				if docsUrl := runFailure.DocsLink; docsUrl != nil {
					result.DocsLink = docsUrl.String()
				}
			}

			results = append(results, result)
			statusReport.Results[category] = results

			if foundErrorInSuite {
				break
			}
		}
	})

	return statusReport
}
