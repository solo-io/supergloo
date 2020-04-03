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

	ExploreCommand = func(validResourceTypes []string) cobra.Command {
		return cobra.Command{
			Use: fmt.Sprintf("explore (%s) resource_name", strings.Join(validResourceTypes, "|")),
			Short: "Explore policies affecting your Kubernetes services (kube-native services) or workloads (e.g., kube-native deployments). " +
				"Format the `resource_name` arg as kube-name.kube-namespace.registered-cluster-name",
		}
	}
)
