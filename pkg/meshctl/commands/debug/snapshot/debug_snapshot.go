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
	zip  string
}

func AddDebugSnapshotFlags(flags *pflag.FlagSet, opts *DebugSnapshotOpts) {
	flags.BoolVar(&opts.json, "json", false, "display the entire json snapshot The output can be piped into a command like jq. For example:\n meshctl debug snapshot discovery input | jq '.'")
	flags.StringVarP(&opts.file, "file", "f", "", "file to write output to")
	flags.StringVar(&opts.zip, "zip", "", "zip file output")
}

func Command(ctx context.Context) *cobra.Command {
	opts := &DebugSnapshotOpts{}
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Input and Output snapshots for the discovery and networking pod",
		Long: "The output can be piped into a command like jq. For example:\n" +
			"meshctl debug snapshot discovery input | jq '.'",
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
		Short: "for the networking pod",
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
		Short: "for the discovery pod",
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
		Short: "input snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return debugSnapshot(ctx, opts, pods, []string{input})
		},
	}
	return cmd
}

func Output(ctx context.Context, opts *DebugSnapshotOpts, pods []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "output",
		Short: "output snapshot",
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
