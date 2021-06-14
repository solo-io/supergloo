package report

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type DebugReportOpts struct {
	verbose bool

	outputFile        string
	meshctlConfigPath string
	namespace         string

	fs afero.Fs
}

func AddDebugReportFlags(flags *pflag.FlagSet, o *DebugReportOpts) {
	utils.AddMeshctlConfigFlags(&o.meshctlConfigPath, flags)
	flags.StringVarP(&o.outputFile, "file", "f", "meshctl-bug-report.tgz", "name of the output tgz file")
	flags.StringVarP(&o.namespace, "namespace", "n", defaults.GetPodNamespace(), "gloo-mesh namespace")
}

func Command(ctx context.Context, globalFlags *utils.GlobalFlags) *cobra.Command {
	opts := &DebugReportOpts{}
	cmd := &cobra.Command{
		Use:   "report",
		Short: "meshctl debug report selectively captures cluster information and logs into an archive to help diagnose problems.",
		Long: `
Running this command requires

1) istioctl to be installed and accessible via your PATH.
2) a meshctl-config-file to be passed in. You can configure this file by running 'meschtl cluster config'.
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.verbose = globalFlags.Verbose
			if opts.meshctlConfigPath == "" {
				var err error
				opts.meshctlConfigPath, err = utils.DefaultMeshctlConfigFilePath()
				if err != nil {
					return err
				}
			}
			return runDebugReportCommand(ctx, opts)
		},
	}
	AddDebugReportFlags(cmd.PersistentFlags(), opts)
	cmd.MarkFlagRequired("meshctl-config-file")
	return cmd
}

func runDebugReportCommand(ctx context.Context, opts *DebugReportOpts) error {
	opts.fs = afero.NewOsFs()
	dir, err := afero.TempDir(opts.fs, "", "")
	if err != nil {
		return err
	}
	defer opts.fs.RemoveAll(dir)

	config, err := utils.ParseMeshctlConfig(opts.meshctlConfigPath)
	if err != nil {
		return err
	}

	fmt.Printf("Running `meschtl version`\n")
	versionsDir := dir + "/version"
	err = opts.fs.MkdirAll(versionsDir, 0755)
	if err != nil {
		return err
	}
	for name, cluster := range config.Clusters {
		clusterVersionDir := versionsDir + "/" + name
		err = opts.fs.MkdirAll(clusterVersionDir, 0755)
		if err != nil {
			return err
		}
		meshctlVersionFile := clusterVersionDir + "/meshctl_version"
		var b bytes.Buffer
		utils.RunShell(fmt.Sprintf("meshctl version --kubeconfig %s --kubecontext %s",
			cluster.KubeConfig, cluster.KubeContext), io.Writer(&b))
		err = ioutil.WriteFile(meshctlVersionFile, b.Bytes(), 0644)
		if err != nil {
			panic(err)
		}
		b.Reset()
	}

	fmt.Printf("Running `meschtl check`\n")
	checkDir := dir + "/check"
	err = opts.fs.MkdirAll(checkDir, 0755)
	if err != nil {
		return err
	}
	for name, cluster := range config.Clusters {
		clusterCheckDir := checkDir + "/" + name
		err = opts.fs.MkdirAll(clusterCheckDir, 0755)
		if err != nil {
			return err
		}
		meshctlCheckFile := clusterCheckDir + "/meshctl_check"
		var b bytes.Buffer
		utils.RunShell(fmt.Sprintf("meshctl check --kubeconfig %s --kubecontext %s",
			cluster.KubeConfig, cluster.KubeContext), io.Writer(&b))
		err = ioutil.WriteFile(meshctlCheckFile, b.Bytes(), 0644)
		if err != nil {
			panic(err)
		}
		b.Reset()
	}

	fmt.Printf("Running `meschtl debug snapshot`\n")
	snapshotsDir := dir + "/snapshot"
	err = opts.fs.MkdirAll(snapshotsDir, 0755)
	if err != nil {
		return err
	}
	for name, cluster := range config.Clusters {
		clusterSnapshotDir := snapshotsDir + "/" + name
		err = opts.fs.MkdirAll(clusterSnapshotDir, 0755)
		if err != nil {
			return err
		}
		snapshotFile := fmt.Sprintf("%s/meshctl_debug_snapshot.tgz", clusterSnapshotDir)
		utils.RunShell(fmt.Sprintf("meshctl debug snapshot --kubeconfig %s --kubecontext %s --zip %s",
			cluster.KubeConfig, cluster.KubeContext, snapshotFile), os.Stdout)
	}

	fmt.Printf("Running `meschtl debug logs`\n")
	err = collectLogs(ctx, opts, dir, config)
	if err != nil {
		return err
	}

	// Istioctl bug-report
	fmt.Printf("Running `istioctl bug-report`\n")
	istioctlBugReportDir := dir + "/istio-bug-report"
	err = opts.fs.MkdirAll(istioctlBugReportDir, 0755)
	if err != nil {
		return err
	}
	for name, cluster := range config.Clusters {
		clusterIstioReportDir := istioctlBugReportDir + "/" + name
		err = opts.fs.MkdirAll(clusterIstioReportDir, 0755)
		if err != nil {
			return err
		}
		if !config.IsMgmtCluster(name) {
			istioctlBugReportCmd := fmt.Sprintf("istioctl bug-report --kubeconfig %s --context %s",
				cluster.KubeConfig, cluster.KubeContext)
			utils.RunShell(istioctlBugReportCmd, os.Stdout)
			utils.RunShell(fmt.Sprintf("mv bug-report.tgz %s", clusterIstioReportDir), os.Stdout)
		}
	}

	err = utils.Zip(opts.fs, dir, opts.outputFile)
	if err != nil {
		return err
	}

	return nil
}
