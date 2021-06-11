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

	fmt.Printf("Running `meschtl debug snapshot`\n")
	var b bytes.Buffer
	snapshotFile := dir + "/meshctl_debug_snapshot.tgz"
	utils.RunShell(fmt.Sprintf("meshctl debug snapshot -c %s --zip %s", opts.meshctlConfigPath, snapshotFile), os.Stdout)

	fmt.Printf("Running `meschtl check`\n")
	b.Reset()
	meshctlCheckFile := dir + "/meshctl_check.txt"
	utils.RunShell(fmt.Sprintf(fmt.Sprintf("meshctl check -c %s", opts.meshctlConfigPath)), io.Writer(&b))
	err = ioutil.WriteFile(meshctlCheckFile, b.Bytes(), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Running `meschtl debug logs`\n")
	config, err := utils.ParseMeshctlConfig(opts.meshctlConfigPath)
	if err != nil {
		return err
	}

	err = collectLogs(ctx, opts, dir, config)
	if err != nil {
		return err
	}

	err = utils.Zip(opts.fs, dir, opts.outputFile)
	if err != nil {
		return err
	}

	return nil
}
