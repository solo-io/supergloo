package cliconstants

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	RootCommand = cobra.Command{
		Use:   "meshctl",
		Short: "CLI for Service Mesh Hub",
	}

	ClusterCommand = cobra.Command{
		Use:   "cluster",
		Short: "Register and perform operations on clusters",
	}

	ClusterRegisterCommand = cobra.Command{
		Use: "register",
		Short: "Register a new cluster by creating a service account token in that cluster through which to authorize " +
			"Service Mesh Hub",
		Long: "In order to specify the remote cluster against which to perform this operation, one or both of the" +
			" --remote-kubeconfig or --remote-context flags must be set. The former selects the kubeconfig file, and " +
			"the latter selects which context should be used from that kubeconfig file",
	}

	DemoCommand = cobra.Command{
		Use:   "demo",
		Short: "Command line utilities for running/interacting with Service Mesh Hub demos",
	}

	DemoInitCommand = cobra.Command{
		Use:   "init",
		Short: "Bootstrap a new local Service Mesh Hub demo setup",
		Long: "Running the Service Mesh Hub demo setup locally requires 4 tools to be installed, and accessible via the " +
			"PATH. meshctl, kubectl, docker, and kind. This command will bootstrap 2 clusters, one of which will run " +
			"the Service Mesh Hub management-plane as well as Istio, and the other will just run Istio.",
	}

	DemoCleanupCommand = cobra.Command{
		Use:   "cleanup",
		Short: "Delete the local Service Mesh Hub demo setup",
		Long: "This will delete all kind clusters running locally, so make sure to only run this script if the only " +
			"kind clusters running are those created by mesctl demo init.",
	}

	GetCommand = struct {
		Root        cobra.Command
		Mesh        cobra.Command
		VirtualMesh cobra.Command
		VMCSR       cobra.Command
		Workload    cobra.Command
		Service     cobra.Command
		Cluster     cobra.Command
	}{
		Root: cobra.Command{
			Use:     "get",
			Aliases: []string{"g"},
			Short:   "Examine Service Mesh Hub resources",
		},
		Mesh: cobra.Command{
			Use:     "meshes",
			Aliases: []string{"m", "mesh"},
			Short:   "Examine discovered meshes",
		},
		Workload: cobra.Command{
			Use:     "workloads",
			Aliases: []string{"w", "workload"},
			Short:   "Examine discovered mesh workloads",
		},
		Service: cobra.Command{
			Use:     "services",
			Aliases: []string{"s", "service"},
			Short:   "Examine discovered mesh services",
		},
		Cluster: cobra.Command{
			Use:     "clusters",
			Aliases: []string{"k", "c", "kubernetescluster", "cluster"},
			Short:   "Examine registered kubernetes clusters",
		},
		VirtualMesh: cobra.Command{
			Use:     "virtualmeshes",
			Aliases: []string{"vm", "vms", "virtualmesh"},
			Short:   "Examine configured virtual meshes",
		},
		VMCSR: cobra.Command{
			Use:     "virtualmeshcertificatesigningrequests",
			Aliases: []string{"vmcsr", "vmcsrs", "virtualmeshcertificatesigningrequest"},
			Short:   "Examine configured virtual mesh ceriticate signing request",
		},
	}

	VersionCommand = cobra.Command{
		Use:   "version",
		Short: "Display the version of meshctl and Service Mesh Hub server components",
	}

	InstallCommand = cobra.Command{
		Use:   "install",
		Short: "Install Service Mesh Hub",
	}

	UpgradeCommand = cobra.Command{
		Use:   "upgrade",
		Short: "In-place upgrade of the meshctl binary",
	}

	UninstallCommand = cobra.Command{
		Use:   "uninstall",
		Short: "Completely uninstall Service Mesh Hub and remove associated CRDs",
	}

	CheckCommand = cobra.Command{
		Use:   "check",
		Short: "Check the status of a Service Mesh Hub installation",
	}

	DescribeCommand = func(validResourceTypes []string) cobra.Command {
		return cobra.Command{
			Use: fmt.Sprintf("describe (%s) resource_name", strings.Join(validResourceTypes, "|")),
			Short: "Describe policies affecting your Kubernetes services (kube-native services) or workloads (e.g., kube-native deployments). " +
				"Format the `resource_name` arg as kube-name.kube-namespace.registered-cluster-name",
		}
	}

	MeshCommand = cobra.Command{
		Use:   "mesh",
		Short: "Manage service meshes",
	}

	MeshInstallCommand = func(validMeshes []string) cobra.Command {
		return cobra.Command{
			Use:     fmt.Sprintf("install (%s)", strings.Join(validMeshes, "|")),
			Aliases: []string{"i"},
			Short:   "Install meshes using meshctl",
		}
	}

	CreateCommand = cobra.Command{
		Use:   "create",
		Aliases: []string{"c"},
		Short: "Create a Service Mesh Hub custom resource",
		Long:  "Utility for creating Service Mesh Hub custom resources",
	}

	CreateVirtualMeshCommand = cobra.Command{
		Use:   "virtualmesh",
		Aliases: []string{"vm"},
		Short: "Create a VirtualMesh resource",
	}

	CreateTrafficPolicyCommand = cobra.Command{
		Use:   "trafficpolicy",
		Aliases: []string{"tp"},
		Short: "Create a TrafficPolicy resource",
	}

	CreateAccessControlPolicyCommand = cobra.Command{
		Use:   "accesscontrolpolicy",
		Aliases: []string{"acp"},
		Short: "Create an AccessControlPolicy resource",
	}
)
