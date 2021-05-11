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
	"gopkg.in/yaml.v2"
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
	binURL, err := getBinaryURL(ctx)
	if err != nil {
		return nil, err
	}
	fmt.Printf("bin url: %s\n", binURL)
	binData, err := get(ctx, binURL)
	if err != nil {
		return nil, err
	}
	defer binData.Close()
	f, err := os.Create(binPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	if _, err := io.Copy(f, binData); err != nil {
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

type manifest struct {
	Versions []struct {
		Tag       string `json:"tag"`
		Platforms []struct {
			OS   string `json:"os"`
			Arch string `json:"arch"`
			URI  string `json:"uri"`
		} `json:"platforms"`
	} `json:"versions"`
}

func getBinaryURL(ctx context.Context) (string, error) {
	const manifestURL = "https://raw.githubusercontent.com/solo-io/meshctl-plugin-index/main/plugins/plugin.yaml"
	body, err := get(ctx, manifestURL)
	if err != nil {
		return "", err
	}
	defer func() {
		io.Copy(ioutil.Discard, body)
		body.Close()
	}()
	var mfst manifest
	if err := yaml.NewDecoder(body).Decode(&mfst); err != nil {
		return "", err
	}

	v, err := version.NewVersion(pkgversion.Version)
	if err != nil {
		return "", err
	}
	major, minor := v.Segments()[0], v.Segments()[1]
	for _, release := range mfst.Versions {
		v, err := version.NewVersion(release.Tag)
		if err != nil {
			logrus.Debugf("invalid semver: %s", release.Tag)
			continue
		}
		if major == v.Segments()[0] && minor == v.Segments()[1] {
			for _, platform := range release.Platforms {
				if platform.OS == runtime.GOOS && platform.Arch == runtime.GOARCH {
					return platform.URI, nil
				}
			}

			return "", eris.New("no compatible plugin manager binary found")
		}
	}

	return "", eris.New("no compatible plugin manager version found")
}

func get(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	} else if res.StatusCode != http.StatusOK {
		defer res.Body.Close()
		if b, err := ioutil.ReadAll(res.Body); err == nil {
			logrus.Debug(string(b))
		} else {
			logrus.Debugf("unable to read response body: %s", err.Error())
		}

		return nil, eris.Errorf("unexpected HTTP response: %d %s", res.StatusCode, res.Status)
	}

	return res.Body, nil
}
