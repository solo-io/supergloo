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
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	Filename        = "/tmp/smh-snapshots.tgz"
	filePermissions = 0644
	kubePort        = "9091"

	// filters for snapshots
	networking = "networking"
	discovery  = "discovery"
	input      = "input"
	output     = "output"
)

type DebugSnapshotOpts struct {
	json bool
	file string
	zip  bool
}

func AddDebugSnapshotFlags(flags *pflag.FlagSet, opts *DebugSnapshotOpts) {
	flags.BoolVar(&opts.json, "json", false, "display the entire json snapshot (best used when piping the output into another command like jq)")
	flags.StringVarP(&opts.file, "file", "f", "", "file to be read or written to")
	flags.BoolVar(&opts.zip, "zip", false, "zip file output")
}

func Command(ctx context.Context) *cobra.Command {
	opts := &DebugSnapshotOpts{}
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Input and Output snapshots for the discovery and networking pod",
		Long: "The output can be piped into a command like jq. For example: \n" +
			"meshctl debug snapshot discovery input | jq 'to_entries | .[] | {kind: (.key), value: .value[]?} | {kind, name: .value.metadata?.name?, namespace: .value.metadata?.namespace?, cluster: .value.metadata?.clusterName?}'",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx, opts, []string{discovery, networking}, []string{input, output})
		},
	}
	cmd.AddCommand(
		Networking(ctx, opts),
		Discovery(ctx, opts),
	)
	AddDebugSnapshotFlags(cmd.PersistentFlags(), opts)
	return cmd
}

func Networking(ctx context.Context, opts *DebugSnapshotOpts) *cobra.Command {
	pods := []string{networking}
	cmd := &cobra.Command{
		Use:   "networking",
		Short: "for the networking pod only",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx, opts, pods, []string{input, output})
		},
	}
	cmd.AddCommand(
		Input(ctx, opts, pods),
		Output(ctx, opts, pods),
	)
	return cmd
}

func Discovery(ctx context.Context, opts *DebugSnapshotOpts) *cobra.Command {
	pods := []string{discovery}
	cmd := &cobra.Command{
		Use:   "discovery",
		Short: "for the discovery pod only",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx, opts, pods, []string{input, output})
		},
	}
	cmd.AddCommand(
		Input(ctx, opts, pods),
		Output(ctx, opts, pods),
	)
	return cmd
}

func Input(ctx context.Context, opts *DebugSnapshotOpts, pods []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "input",
		Short: "input snapshot only",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx, opts, pods, []string{input})
		},
	}
	return cmd
}

func Output(ctx context.Context, opts *DebugSnapshotOpts, pods []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "output snapshot only",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx, opts, pods, []string{output})
		},
	}
	return cmd
}

func debugSnapshot(ctx context.Context, opts *DebugSnapshotOpts, pods, types []string) error {
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
			snapshot := getSnapshot(ctx, localPort, podName, snapshotType)
			fileName := fmt.Sprintf("%s-%s-snapshot.json", podName, snapshotType)
			var snapshotStr string
			if opts.json {
				snapshotStr = snapshot
			} else {
				err = ioutil.WriteFile(fileName, []byte(snapshot), filePermissions)
				if err != nil {
					return err
				}
				fmt.Printf("%s %s snapshot:\n", podName, snapshotType)
				pipedCmd := "cat " + fileName + " | jq 'to_entries | .[] | {kind: (.key), value: .value[]?} | {kind, name: .value.metadata?.name?, namespace: .value.metadata?.namespace?, cluster: .value.metadata?.clusterName?}'"
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
			if opts.zip {
				err = storageClient.Save(dir, &debugutils.StorageObject{
					Resource: strings.NewReader(snapshotStr),
					Name:     fileName,
				})
				if err != nil {
					return err
				}
			} else if opts.file != "" {
				header := fmt.Sprintf("%s %s snapshot:\n", podName, snapshotType)
				_, err = f.WriteString(header + snapshotStr + "\n")
				if err != nil {
					fmt.Println(err.Error())
					return err
				}
			}
		}
	}
	if opts.zip {
		if opts.file == "" {
			opts.file = Filename
		}
		err = zip(fs, dir, opts.file)
	}
	return nil
}

func getSnapshot(ctx context.Context, localPort, podName, snapshotType string) string {
	snapshot, portFwdCmd, err := cliutils.PortForwardGet(ctx,
		"service-mesh-hub", "deploy/"+podName,
		localPort, kubePort, false, "/snapshots/"+snapshotType)
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
