package table_printing

import (
	"io"

	"github.com/google/wire"
	"github.com/olekukonko/tablewriter"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
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
	Print(out io.Writer, printMode PrintMode, trafficPolicies []*smh_networking.TrafficPolicy) error
}

type AccessControlPolicyPrinter interface {
	Print(out io.Writer, printMode PrintMode, accessControlPolicies []*smh_networking.AccessControlPolicy) error
}

type VirtualMeshPrinter interface {
	Print(out io.Writer, virtualMeshes []*smh_networking.VirtualMesh) error
}

type VirtualMeshCSRPrinter interface {
	Print(out io.Writer, vmcsrs []*smh_security.VirtualMeshCertificateSigningRequest) error
}

type MeshPrinter interface {
	Print(out io.Writer, meshes []*smh_discovery.Mesh) error
}

type MeshWorkloadPrinter interface {
	Print(out io.Writer, meshWorkloads []*smh_discovery.MeshWorkload) error
}

type MeshServicePrinter interface {
	Print(out io.Writer, meshServices []*smh_discovery.MeshService) error
}

type KubernetesClusterPrinter interface {
	Print(out io.Writer, meshes []*smh_discovery.KubernetesCluster) error
}
