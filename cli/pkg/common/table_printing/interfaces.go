package table_printing

import (
	"io"

	"github.com/google/wire"
	"github.com/olekukonko/tablewriter"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	security_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
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
	Print(out io.Writer, printMode PrintMode, trafficPolicies []*networking_v1alpha1.TrafficPolicy) error
}

type AccessControlPolicyPrinter interface {
	Print(out io.Writer, printMode PrintMode, accessControlPolicies []*networking_v1alpha1.AccessControlPolicy) error
}

type VirtualMeshPrinter interface {
	Print(out io.Writer, meshServices []*networking_v1alpha1.VirtualMesh) error
}

type VirtualMeshCSRPrinter interface {
	Print(out io.Writer, meshServices []*security_v1alpha1.VirtualMeshCertificateSigningRequest) error
}

type MeshPrinter interface {
	Print(out io.Writer, meshes []*discovery_v1alpha1.Mesh) error
}

type MeshWorkloadPrinter interface {
	Print(out io.Writer, meshWorkloads []*discovery_v1alpha1.MeshWorkload) error
}

type MeshServicePrinter interface {
	Print(out io.Writer, meshServices []*discovery_v1alpha1.MeshService) error
}

type KubernetesClusterPrinter interface {
	Print(out io.Writer, meshes []*discovery_v1alpha1.KubernetesCluster) error
}
