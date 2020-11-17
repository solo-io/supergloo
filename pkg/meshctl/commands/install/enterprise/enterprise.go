package enterprise

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/commands/install/internal/flags"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/install/gloomesh"
	"github.com/solo-io/gloo-mesh/pkg/meshctl/registration"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

func Command(ctx context.Context) *cobra.Command {
	opts := &options{}
	cmd := &cobra.Command{
		Use:   "enterprise",
		Short: "Install Gloo Mesh enterprise",
		RunE: func(cmd *cobra.Command, args []string) error {
			return install(ctx, opts)
		},
	}
	opts.addToFlags(cmd.Flags())

	cmd.SilenceUsage = true
	return cmd
}

type options struct {
	flags.Options
	licenseKey string
	skipUI     bool
	skipRBAC   bool
}

func (o *options) addToFlags(flags *pflag.FlagSet) {
	o.AddToFlags(flags)
	flags.StringVar(&o.licenseKey, "license", "", "Gloo Mesh Enterprise license key")
	cobra.MarkFlagRequired(flags, "license")
	flags.BoolVar(&o.skipUI, "skip-ui", false, "Skip installation of the Gloo Mesh UI")
	flags.BoolVar(&o.skipRBAC, "skip-rbac", false, "Skip installation of the RBAC Webhook")
}

func install(ctx context.Context, opts *options) error {
	version, err := latestChartVersion()
	if err != nil {
		return err
	}
	opts.Version = version
	installer := opts.GetInstaller(gloomesh.GlooMeshEnterpriseChartUriTemplate)
	installer.Values["license.key"] = opts.licenseKey
	if opts.skipUI {
		installer.Values["gloo-mesh-ui.enabled"] = "false"
	}
	if opts.skipRBAC {
		installer.Values["rbac-webhook.enabled"] = "false"
	}
	if err := installer.InstallGlooMeshEnterprise(ctx); err != nil {
		return eris.Wrap(err, "installing gloo-mesh-enterprise")
	}
	if opts.Register && !opts.DryRun {
		registrantOpts := opts.GetRegistrationOptions()
		registrant, err := registration.NewRegistrant(&registrantOpts)
		if err != nil {
			return eris.Wrap(err, "initializing registrant")
		}
		if err := registrant.RegisterCluster(ctx); err != nil {
			return eris.Wrap(err, "registering management-plane cluster")
		}
	}

	return nil
}

func latestChartVersion() (string, error) {
	const chartIndexURI = "https://storage.googleapis.com/gloo-mesh-enterprise/gloo-mesh-enterprise/index.yaml"
	res, err := http.Get(chartIndexURI)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		io.Copy(ioutil.Discard, res.Body)
		return "", fmt.Errorf("invalid response from the Helm repository: %d %s", res.StatusCode, res.Status)
	}
	index := struct {
		Entries struct {
			GlooMesh []struct {
				Version string `yaml:"version"`
			} `yaml:"gloo-mesh"`
		} `yaml:"entries"`
	}{}
	if err := yaml.NewDecoder(res.Body).Decode(&index); err != nil {
		return "", err
	}
	if len(index.Entries.GlooMesh) == 0 {
		return "", eris.New("no Gloo Mesh Enterprise versions found")
	}

	return index.Entries.GlooMesh[0].Version, nil
}
