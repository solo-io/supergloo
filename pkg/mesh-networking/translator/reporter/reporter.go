package reporter

import v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"

// the reporter reports status errors on user configuration objects
type Reporter interface {
	// report an error on a traffic policy
	ReportTrafficPolicy(trafficPolicy *v1.ObjectRef, err error)

	// report an error on an access policy
	ReportAccessPolicy(accessPolicy *v1.ObjectRef, err error)
}
