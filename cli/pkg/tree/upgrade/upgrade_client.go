package upgrade

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/google/wire"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/cli/pkg/common"
	"github.com/solo-io/service-mesh-hub/cli/pkg/options"
	upgrade_assets "github.com/solo-io/service-mesh-hub/cli/pkg/tree/upgrade/assets"
)

var (
	GetReleaseForRepoError = func(err error, tag string) error {
		return eris.Wrapf(err, "failed to get release '%s' from %s/%s repository",
			tag, upgrade_assets.OrgName, upgrade_assets.RepoName)
	}
	CouldNotFindAssetForReleaseError = func(tag, assetName string) error {
		return eris.Errorf("failed to find asset %s in release %s", assetName, tag)
	}
	DownloadingAssetError = func(err error, assetName string) error {
		return eris.Wrapf(err, "error occurred attempting to download %s", assetName)
	}

	// wire set to package up upgrade client set dependencies
	UpgraderClientSet = wire.NewSet(
		upgrade_assets.DefaultGithubAssetClient,
		upgrade_assets.NewAssetHelper,
	)
)

type UpgradeOpts struct {
	DownloadPath string
	ReleaseTag   string
}

func Upgrade(ctx context.Context, opts *options.Options, out io.Writer, clientFactory common.ClientsFactory) error {
	clients, err := clientFactory(opts)
	if err != nil {
		return err
	}
	meshctlBinaryName := upgrade_assets.MeshctlBinaryName()
	release, err := clients.ReleaseAssetHelper.GetReleaseWithAsset(ctx, opts.Upgrade.ReleaseTag, meshctlBinaryName)
	if err != nil {
		return GetReleaseForRepoError(err, opts.Upgrade.ReleaseTag)
	}

	fmt.Fprintf(out, "Downloading %s from release tag %s\n", meshctlBinaryName, release.GetTagName())

	asset := upgrade_assets.TryGetAssetWithName(release, meshctlBinaryName)
	if asset == nil {
		return CouldNotFindAssetForReleaseError(meshctlBinaryName, release.GetTagName())
	}

	if err := clients.ReleaseAssetHelper.DownloadAsset(
		asset.GetBrowserDownloadURL(),
		opts.Upgrade.DownloadPath); err != nil {
		return DownloadingAssetError(err, meshctlBinaryName)
	}

	downloadPath := opts.Upgrade.DownloadPath
	if downloadPath == "" {
		downloadPath, err = os.Executable()
		if err != nil {
			return eris.Wrapf(err, "failed to get currently executing binary path")
		}
	}

	fmt.Fprintf(out, "Successfully downloaded and installed meshctl version %s to %s\n", release.GetTagName(), downloadPath)
	return nil
}
