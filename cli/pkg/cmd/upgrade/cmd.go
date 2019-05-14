package upgrade

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"

	"github.com/solo-io/supergloo/cli/pkg/options"

	"github.com/inconshreveable/go-update"

	"github.com/google/go-github/github"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/spf13/cobra"
)

func Cmd(opts *options.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upgrade",
		Aliases: []string{"ug"},
		Short:   "upgrade your supergloo cli binary to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return upgradeSuperglooCli(opts.Ctx, opts.Upgrade)
		},
	}

	cmd.PersistentFlags().StringVar(&opts.Upgrade.ReleaseTag, "release", "latest", "Which supergloo release "+
		"to download. Specify a git tag corresponding to the desired version of supergloo.")
	cmd.PersistentFlags().StringVar(&opts.Upgrade.DownloadPath, "path", "", "Desired path for your "+
		"upgraded supergloo cli binary. Defaults to the location of your currently executing binary.")
	return cmd
}

func upgradeSuperglooCli(ctx context.Context, upgrade options.Upgrade) error {
	superglooCliBinaryName := fmt.Sprintf("supergloo-cli-%v-amd64", runtime.GOOS)
	release, err := getReleaseWithAsset(ctx, upgrade.ReleaseTag, superglooCliBinaryName)
	if err != nil {
		return errors.Wrapf(err, "getting release '%v' from solo-io/supergloo repository", upgrade.ReleaseTag)
	}

	fmt.Printf("downloading %v from release tag %v\n", superglooCliBinaryName, release.GetTagName())

	asset := tryGetAssetWithName(release, superglooCliBinaryName)
	if asset == nil {
		return errors.Errorf("could not find asset %v in release %v", superglooCliBinaryName, release.GetTagName())
	}

	if err := downloadAsset(asset.GetBrowserDownloadURL(), upgrade.DownloadPath); err != nil {
		return errors.Wrapf(err, "downloading asset %v", superglooCliBinaryName)
	}

	downloadPath := upgrade.DownloadPath
	if downloadPath == "" {
		downloadPath, err = os.Executable()
		if err != nil {
			return errors.Wrapf(err, "getting currently executing binary path")
		}
	}

	fmt.Printf("successfully downloaded and installed supergloo-cli version %v to %v\n", release.GetTagName(), downloadPath)
	return nil
}

func getReleaseWithAsset(ctx context.Context, tag string, expectedAssetName string) (*github.RepositoryRelease, error) {
	g := github.NewClient(nil)
	if tag == "latest" {
		// don't use latest tag, because that might not have the assets yet if the release build is running.
		listOpts := github.ListOptions{PerPage: 10}
		releases, _, err := g.Repositories.ListReleases(ctx, "solo-io", "supergloo", &listOpts)
		if err != nil {
			return nil, errors.Wrapf(err, "error listing releases")
		}
		for _, release := range releases {
			if tryGetAssetWithName(release, expectedAssetName) != nil {
				return release, nil
			}
		}
		return nil, errors.Errorf("couldn't find any recent release with the desired asset")
	}
	release, _, err := g.Repositories.GetReleaseByTag(ctx, "solo-io", "gloo", tag)
	return release, err
}

func tryGetAssetWithName(release *github.RepositoryRelease, expectedAssetName string) *github.ReleaseAsset {
	for _, asset := range release.Assets {
		if asset.GetName() == expectedAssetName {
			return &asset
		}
	}
	return nil
}

func downloadAsset(downloadUrl string, destFile string) error {
	res, err := http.Get(downloadUrl)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if err := update.Apply(res.Body, update.Options{
		TargetPath: destFile,
	}); err != nil {
		return err
	}
	return nil
}
