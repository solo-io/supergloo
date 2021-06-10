package report

import (
	"context"
	"io"

	"github.com/solo-io/gloo-mesh/pkg/common/defaults"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/utils"
	"github.com/solo-io/go-utils/tarutils"
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
	flags.StringVarP(&o.meshctlConfigPath, "meshctl-config-file", "c", "",
		"path to the meshctl config file. defaults to `$HOME/.gloo-mesh/meshctl-config.yaml`")
	flags.StringVarP(&o.outputFile, "file", "f", "bug-report.tgz", "name of the output tgz file")
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

	err = collectLogs(ctx, opts, dir, config)
	if err != nil {
		return err
	}

	err = zip(opts.fs, dir, opts.outputFile)
	if err != nil {
		return err
	}

	return nil
}

func zip(fs afero.Fs, dir string, file string) error {
	tarball, err := fs.Create(file)
	if err != nil {
		return err
	}
	if err := tarutils.Tar(dir, fs, tarball); err != nil {
		return err
	}
	_, err = tarball.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	return nil
}
