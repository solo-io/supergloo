package snapshot

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/solo-io/go-utils/cliutils"
	"github.com/solo-io/go-utils/debugutils"
	"github.com/solo-io/go-utils/tarutils"
	"github.com/solo-io/service-mesh-hub/pkg/common/defaults"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	filePermissions = 0644

	// filters for snapshots
	networking = "networking"
	discovery  = "discovery"
	input      = "input"
	output     = "output"

	// jq query
	query = "to_entries | .[] | select(.key != \"clusters\") | select(.key != \"name\") | {kind: .key, list : [.value[]? | {name: .metadata.name, namespace: .metadata.namespace, cluster: .metadata.clusterName}]}"
)

type DebugSnapshotOpts struct {
	json    bool
	file    string
	zip     string
	verbose bool

	// hidden optional values
	metricsBindPort uint32
	namespace       string
}

func AddDebugSnapshotFlags(flags *pflag.FlagSet, opts *DebugSnapshotOpts) {
	flags.BoolVar(&opts.json, "json", false, "display the entire json snapshot. The output can be piped into a command like jq (https://stedolan.github.io/jq/tutorial/). For example:\n meshctl debug snapshot discovery input | jq '.'")
	flags.StringVarP(&opts.file, "file", "f", "", "file to write output to")
	flags.StringVar(&opts.zip, "zip", "", "zip file output")
	flags.BoolVar(&opts.verbose, "verbose", false, "enables verbose/debug logging")
	flags.Uint32Var(&opts.metricsBindPort, "port", defaults.MetricsPort, "metrics port")
	flags.StringVarP(&opts.namespace, "namespace", "n", defaults.GetPodNamespace(), "service-mesh-hub namespace")
}

func Command(ctx context.Context) *cobra.Command {
	opts := &DebugSnapshotOpts{}
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Input and Output snapshots for the discovery and networking pod. Requires jq to be installed if the --json flag is not being used.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx, opts, []string{discovery, networking}, []string{input, output})
		},
	}
	cmd.AddCommand(
		Networking(ctx, opts),
		Discovery(ctx, opts),
	)
	AddDebugSnapshotFlags(cmd.PersistentFlags(), opts)
	cmd.PersistentFlags().Lookup("namespace").Hidden = true
	cmd.PersistentFlags().Lookup("port").Hidden = true
	return cmd
}

func Networking(ctx context.Context, opts *DebugSnapshotOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "networking",
		Short: "Input and output snapshots for the networking pod",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx, opts, []string{networking}, []string{input, output})
		},
	}
	cmd.AddCommand(
		Input(ctx, opts, networking),
		Output(ctx, opts, networking),
	)
	return cmd
}

func Discovery(ctx context.Context, opts *DebugSnapshotOpts) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "discovery",
		Short: "Input and output snapshots for the discovery pod",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx, opts, []string{discovery}, []string{input, output})
		},
	}
	cmd.AddCommand(
		Input(ctx, opts, discovery),
		Output(ctx, opts, discovery),
	)
	return cmd
}

func Input(ctx context.Context, opts *DebugSnapshotOpts, pod string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "input",
		Short: fmt.Sprintf("Input snapshot for the %s pod", pod),
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx, opts, []string{pod}, []string{input})
		},
	}
	return cmd
}

func Output(ctx context.Context, opts *DebugSnapshotOpts, pod string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: fmt.Sprintf("Output snapshot for the %s pod", pod),
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx, opts, []string{pod}, []string{output})
		},
	}
	return cmd
}

func debugSnapshot(ctx context.Context, opts *DebugSnapshotOpts, pods, types []string) error {
	// Check prerequisite jq is installed
	_, err := exec.Command("which", "jq").Output()
	if err != nil {
		fmt.Printf("Error: Could not find jq! Please install it from https://stedolan.github.io/jq/download/\n")
		return err
	}

	freePort, err := cliutils.GetFreePort()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	localPort := strconv.Itoa(freePort)

	f, err := os.Create(opts.file)
	defer f.Close()

	fs := afero.NewOsFs()
	dir, err := afero.TempDir(fs, "", "")
	if err != nil {
		return err
	}
	defer fs.RemoveAll(dir)
	storageClient := debugutils.NewFileStorageClient(fs)
	for _, podName := range pods {
		for _, snapshotType := range types {
			snapshot := getSnapshot(ctx, opts, localPort, podName, snapshotType)
			fileName := fmt.Sprintf("%s-%s-snapshot.json", podName, snapshotType)
			var snapshotStr string
			if opts.json {
				snapshotStr = snapshot
			} else {
				err = ioutil.WriteFile(fileName, []byte(snapshot), filePermissions)
				if err != nil {
					return err
				}
				pipedCmd := "cat " + fileName + " | jq '" + query + "'"
				cmdOut, err := exec.Command("bash", "-c", pipedCmd).Output()
				if err != nil {
					return err
				}
				snapshotStr = string(cmdOut)
				err = os.Remove(fileName)
				if err != nil {
					return err
				}
			}
			if opts.file != "" {
				_, err = f.WriteString(snapshotStr)
				if err != nil {
					fmt.Println(err.Error())
					return err
				}
			} else if opts.zip != "" {
				err = storageClient.Save(dir, &debugutils.StorageObject{
					Resource: strings.NewReader(snapshotStr),
					Name:     fileName,
				})
				if err != nil {
					return err
				}
			} else {
				fmt.Print(snapshotStr)
			}
		}
	}
	if opts.zip != "" {
		err = zip(fs, dir, opts.zip)
	}
	return nil
}

func getSnapshot(ctx context.Context, opts *DebugSnapshotOpts, localPort, podName, snapshotType string) string {
	snapshot, portFwdCmd, err := cliutils.PortForwardGet(ctx, opts.namespace, "deploy/"+podName,
		localPort, strconv.Itoa(int(opts.metricsBindPort)), opts.verbose, "/snapshots/"+snapshotType)
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	if portFwdCmd.Process != nil {
		defer portFwdCmd.Process.Release()
		defer portFwdCmd.Process.Kill()
	}
	return snapshot
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
