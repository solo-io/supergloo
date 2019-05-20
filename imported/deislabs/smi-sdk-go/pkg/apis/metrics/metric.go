package metrics

import (
	"fmt"

	apiresource "k8s.io/apimachinery/pkg/api/resource"
)

// AvailableMetrics are a list of the possible metrics that can be returned
var AvailableMetrics = []*Metric{
	{
		Name:  "p99_response_latency",
		Unit:  MilliSeconds,
		Value: apiresource.NewQuantity(0, apiresource.DecimalSI),
	},
	{
		Name:  "p90_response_latency",
		Unit:  MilliSeconds,
		Value: apiresource.NewQuantity(0, apiresource.DecimalSI),
	},
	{
		Name:  "p50_response_latency",
		Unit:  MilliSeconds,
		Value: apiresource.NewQuantity(0, apiresource.DecimalSI),
	},
	{
		Name:  "success_count",
		Value: apiresource.NewQuantity(0, apiresource.DecimalSI),
	},
	{
		Name:  "failure_count",
		Value: apiresource.NewQuantity(0, apiresource.DecimalSI),
	},
}

// Unit is associated with metrics and defines what unit the metric is using
type Unit string

// MilliSeconds is a time unit
const MilliSeconds Unit = "ms"

// Metric describes a name and value for specific metrics
type Metric struct {
	Name string `json:"name"`

	Unit  Unit                  `json:"unit,omitempty"`
	Value *apiresource.Quantity `json:"value"`
}

// Set will set the value correctly
func (m *Metric) Set(val float64) {
	m.Value = apiresource.NewMilliQuantity(
		int64(val*1000), apiresource.DecimalSI)
}

// String returns a formatted string representation of this struct
func (m *Metric) String() string {
	return fmt.Sprintf("%#v", m)
}
