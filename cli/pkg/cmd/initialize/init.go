package initialize

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/solo-io/supergloo/cli/pkg/helpers"
	v1 "k8s.io/api/core/v1"
	kubeerrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/cli/pkg/flagutils"
	"github.com/solo-io/supergloo/cli/pkg/options"
	"github.com/solo-io/supergloo/pkg/install/utils/helm"
	"github.com/solo-io/supergloo/pkg/version"
	"github.com/spf13/cobra"
)

const (
	sgChartUriTemplate = "https://storage.googleapis.com/supergloo-helm/charts/supergloo-%s.tgz"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "install SuperGloo to a Kubernetes cluster",
		Long: `Installs SuperGloo using default values based on the official helm chart located in install/helm/supergloo

The basic SuperGloo installation is composed of single-instance deployments for the supergloo-controller and discovery pods. 
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := installSuperGloo(opts); err != nil {
				return errors.Wrapf(err, "installing gloo in gateway mode")
			}
			return nil
		},
	}
	flagutils.AddInitFlags(cmd.PersistentFlags(), &opts.Init)
	return cmd
}

func installSuperGloo(opts *options.Options) error {
	version, err := getReleaseVersion(opts)
	if err != nil {
		return errors.Wrapf(err, "getting release version")
	}

	// Get location of Gloo helm chart
	chartUri := fmt.Sprintf(sgChartUriTemplate, version)
	if helmChartOverride := opts.Init.HelmChartOverride; helmChartOverride != "" {
		chartUri = helmChartOverride
	}

	values, err := readValues(opts.Init.HelmValues)
	if err != nil {
		return errors.Wrapf(err, "reading custom values")
	}

	kube := helpers.MustKubeClient()
	if _, err := kube.CoreV1().Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: opts.Init.InstallNamespace},
	}); err != nil && !kubeerrs.IsAlreadyExists(err) {
		return errors.Wrapf(err, "creating namespace")
	}

	manifests, err := helm.RenderManifests(opts.Ctx,
		chartUri,
		values,
		"supergloo",
		opts.Init.InstallNamespace,
		"",
		true,
	)
	if err != nil {
		return errors.Wrapf(err, "rendering manifests")
	}

	manifest := manifests.CombinedString()

	if opts.Init.DryRun {
		fmt.Printf("%s\n", manifest)
		return nil
	}

	fmt.Printf("installing supergloo version %v\nusing chart uri %v\n", version, chartUri)

	if err := kubectlApply(manifest); err != nil {
		return errors.Wrapf(err, "executing kubectl failed")
	}

	fmt.Printf("install successful!\n")
	return nil
}

func getReleaseVersion(opts *options.Options) (string, error) {
	if !version.IsReleaseVersion() {
		if opts.Init.ReleaseVersion == "" {
			return "", errors.Errorf("you must provide a " +
				"release version containing the manifest when " +
				"running an unreleased version of glooctl.")
		}
		return opts.Init.ReleaseVersion, nil
	}
	return version.Version, nil
}

func readValues(path string) (string, error) {
	if path == "" {
		return "", nil
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func kubectlApply(manifest string) error {
	return kubectl(bytes.NewBufferString(manifest), "apply", "-f", "-")
}

func kubectl(stdin io.Reader, args ...string) error {
	kubectl := exec.Command("kubectl", args...)
	if stdin != nil {
		kubectl.Stdin = stdin
	}
	kubectl.Stdout = os.Stdout
	kubectl.Stderr = os.Stderr
	return kubectl.Run()
}
