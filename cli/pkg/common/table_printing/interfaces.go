package table_printing

import (
	"io"

	"github.com/olekukonko/tablewriter"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
)

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

// produce a new table configured with the layout options that you want
type TableBuilder func(out io.Writer) *tablewriter.Table

// tables may contain different information depending on if you are interested in its data
// from the point of view of a service or a workload
type PrintMode string

const (
	ServicePrintMode  PrintMode = "service"
	WorkloadPrintMode PrintMode = "workload"
)

type TrafficPolicyPrinter interface {
	Print(out io.Writer, printMode PrintMode, trafficPolicies []*networking_v1alpha1.TrafficPolicy) error
}

type AccessControlPolicyPrinter interface {
	Print(out io.Writer, printMode PrintMode, accessControlPolicies []*networking_v1alpha1.AccessControlPolicy) error
}
