package initpluginmanager

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/hashicorp/go-version"
	"github.com/rotisserie/eris"
	"github.com/sirupsen/logrus"
	pkgversion "github.com/solo-io/gloo-mesh/pkg/common/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "init-plugin-manager",
		Short: "Install the Gloo Mesh Enterprise CLI plugin manager",
		// TODO(ryantking): Add link to plugin docs after written
		RunE: func(*cobra.Command, []string) error {
			home, err := opts.getHome()
			if err != nil {
				return err
			}
			if err := checkExisting(home, opts.force); err != nil {
				return err
			}
			binary, err := downloadTempBinary(ctx, home)
			if err != nil {
				return err
			}
			const defaultIndexURL = "https://github.com/solo-io/meshctl-plugin-index.git"
			if out, err := binary.run("index", "add", "default", defaultIndexURL); err != nil {
				fmt.Println(out)
				return err
			}
			if out, err := binary.run("install", "plugin"); err != nil {
				fmt.Println(out)
				return err
			}
			homeStr := opts.home
			if homeStr == "" {
				homeStr = "$HOME/.gloo-mesh"
			}
			fmt.Printf(`The meshctl plugin manager was successfully installed 🎉
Add the meshctl plugins to your path with:
  export PATH=%s/bin:$PATH
Now run:
  meshctl plugin --help     # see the commands available to you
Please see visit the Gloo Mesh website for more info:  https://www.solo.io/products/gloo-mesh/
`, homeStr)
			return nil
		},
	}

	opts.addToFlags(cmd.Flags())
	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	home  string
	force bool
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	flags.StringVar(&o.home, "gm-home", "", "Gloo Mesh home directory (default: $HOME/.gloo-mesh)")
	flags.BoolVarP(&o.force, "force", "f", false, "Delete any existing plugin data if found and reinitialize")
}

func (o options) getHome() (string, error) {
	if o.home != "" {
		return o.home, nil
	}
	userHome, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userHome, ".gloo-mesh"), nil
}

type pluginBinary struct {
	path string
	home string
}

func checkExisting(home string, force bool) error {
	pluginDirs := []string{"index", "receipts", "store"}
	dirty := false
	for _, dir := range pluginDirs {
		if _, err := os.Stat(filepath.Join(home, dir)); err == nil {
			dirty = true
			break
		} else if !os.IsNotExist(err) {
			return err
		}
	}
	if !dirty {
		return nil
	}
	if !force {
		return eris.Errorf("found existing plugin manager files in %s, rerun with -f to delete and reinstall", home)
	}
	for _, dir := range pluginDirs {
		os.RemoveAll(filepath.Join(home, dir))
	}
	binFiles, err := ioutil.ReadDir(filepath.Join(home, "bin"))
	if err != nil {
		return err
	}
	for _, file := range binFiles {
		if file.Name() != "meshctl" {
			os.Remove(filepath.Join(home, "bin", file.Name()))
		}
	}
	return nil
}

func downloadTempBinary(ctx context.Context, home string) (*pluginBinary, error) {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	binPath := filepath.Join(tempDir, "plugin")
	if runtime.GOARCH != "amd64" {
		return nil, eris.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
	if runtime.GOOS != "darwin" && runtime.GOOS != "linux" {
		return nil, eris.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
	bin, err := findLatestCompatibleBinary(ctx)
	if err != nil {
		return nil, err
	}
	defer bin.Close()
	f, err := os.Create(binPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := io.Copy(f, bin); err != nil {
		return nil, err
	}
	if err := f.Chmod(0755); err != nil {
		return nil, err
	}

	return &pluginBinary{path: binPath, home: home}, nil
}

func (binary pluginBinary) run(args ...string) (string, error) {
	cmd := exec.Command(binary.path, args...)
	cmd.Env = append(cmd.Env, "MESHCTL_HOME="+binary.home)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Attempts to find the latest plugin manager binary compatible with the current version of meshctl
func findLatestCompatibleBinary(ctx context.Context) (io.ReadCloser, error) {
	v, err := version.NewVersion(pkgversion.Version)
	if err != nil {
		return nil, eris.Wrap(err, "unable to parse version")
	}
	major, minor, patch := v.Segments()[0], v.Segments()[1], v.Segments()[2]
	if v.Prerelease() != "" {
		minor -= 1
		patch = 20
	}
	for ; patch >= 0; patch-- {
		tryVersion := fmt.Sprintf("v%d.%d.%d", major, minor, patch)
		body, err := getBinary(ctx, tryVersion)
		if err != nil {
			return nil, err
		}
		if body != nil {
			return body, nil
		}
	}

	return nil, eris.Errorf("no compatible version found for meshctl %s:", pkgversion.Version)
}

// Attempts to download a plugin manager binary with the given version
// Returns the following based on the response from google cloud storage:
//   200: Return the body and a nil error
//   404: Return a nil body and nil error
//   Other: Return a nil body and an error with the unexpected status code
func getBinary(ctx context.Context, version string) (io.ReadCloser, error) {
	const urlFmt = "https://storage.googleapis.com/gloo-mesh-enterprise/meshctl-plugins/plugin/%s/meshctl-plugin-%s-%s"
	url := fmt.Sprintf(urlFmt, version, runtime.GOOS, runtime.GOARCH)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	} else if res.StatusCode == http.StatusNotFound {
		defer res.Body.Close()
		io.Copy(ioutil.Discard, res.Body)
		return nil, nil
	} else if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		if b, err := ioutil.ReadAll(res.Body); err == nil {
			logrus.Debug(string(b))
		} else {
			logrus.Debugf("unable to read Google Cloud response body: %s", err.Error())
		}

		return nil, eris.Errorf("could not download plugin manager binary: %d %s", res.StatusCode, res.Status)
	}

	return res.Body, nil
}
