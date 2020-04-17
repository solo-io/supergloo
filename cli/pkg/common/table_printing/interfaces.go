package table_printing

import (
	"io"

	"github.com/google/wire"
	"github.com/olekukonko/tablewriter"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
)

var TablePrintingSet = wire.NewSet(
	DefaultTableBuilder,
	NewTrafficPolicyPrinter,
	NewAccessControlPolicyPrinter,
	NewMeshPrinter,
	NewMeshServicePrinter,
	NewMeshWorkloadPrinter,
	NewKubernetesClusterPrinter,
	NewVirtualMeshPrinter,
	NewVirtualMeshMCSRPrinter,
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
	Print(out io.Writer, printMode PrintMode, trafficPolicies []*zephyr_networking.TrafficPolicy) error
}

type AccessControlPolicyPrinter interface {
	Print(out io.Writer, printMode PrintMode, accessControlPolicies []*zephyr_networking.AccessControlPolicy) error
}

type VirtualMeshPrinter interface {
	Print(out io.Writer, meshServices []*zephyr_networking.VirtualMesh) error
}

type VirtualMeshCSRPrinter interface {
	Print(out io.Writer, meshServices []*zephyr_security.VirtualMeshCertificateSigningRequest) error
}

type MeshPrinter interface {
	Print(out io.Writer, meshes []*zephyr_discovery.Mesh) error
}

type MeshWorkloadPrinter interface {
	Print(out io.Writer, meshWorkloads []*zephyr_discovery.MeshWorkload) error
}

type MeshServicePrinter interface {
	Print(out io.Writer, meshServices []*zephyr_discovery.MeshService) error
}

type KubernetesClusterPrinter interface {
	Print(out io.Writer, meshes []*zephyr_discovery.KubernetesCluster) error
}
