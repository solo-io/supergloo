package cliconstants

import "github.com/spf13/cobra"

var (
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

	VersionCommand = cobra.Command{
		Use:   "version",
		Short: "Display the version of meshctl and Service Mesh Hub server components",
	}

	InstallCommand = cobra.Command{
		Use:   "install",
		Short: "Install Service Mesh Hub on different platforms",
	}

	UpgradeCommand = cobra.Command{
		Use:   "upgrade",
		Short: "In-place upgrade of meshctl",
	}
)
