package upgrade_assets

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/google/go-github/github"
	"github.com/inconshreveable/go-update"
	"github.com/rotisserie/eris"
)

const (
	RepoName = "mesh-projects"
	OrgName  = "solo-io"
)

func MeshctlBinaryName() string {
	switch runtime.GOOS {
	case `windows`:
		return fmt.Sprintf("meshctl-%s-amd64.exe", runtime.GOOS)
	default:
		return fmt.Sprintf("meshctl-%s-amd64", runtime.GOOS)
	}
}

//go:generate mockgen -destination ./mocks/helpers.go -source ./helpers.go
var (
	DefaultRequestTimeout = 5 * time.Second

	NoReleaseContainingAssetError = eris.New("couldn't find any release with the desired asset")
)

type GithubAssetClient interface {
	ListReleases(ctx context.Context, opt *github.ListOptions) ([]*github.RepositoryRelease, error)
	GetReleaseByTag(ctx context.Context, tag string) (*github.RepositoryRelease, error)
}

type githubAssetClient struct {
	org    string
	repo   string
	client *github.Client
}

func NewGithubAssetClient(client *github.Client, org, repo string) GithubAssetClient {
	return &githubAssetClient{org: org, repo: repo, client: client}
}

func DefaultGithubAssetClient() GithubAssetClient {
	return &githubAssetClient{
		org:  OrgName,
		repo: RepoName,
		client: github.NewClient(&http.Client{
			Timeout: DefaultRequestTimeout,
		}),
	}
}

func (a *githubAssetClient) ListReleases(ctx context.Context, opt *github.ListOptions) ([]*github.RepositoryRelease, error) {
	release, _, err := a.client.Repositories.ListReleases(ctx, a.org, a.repo, opt)
	return release, err
}

func (a *githubAssetClient) GetReleaseByTag(ctx context.Context, tag string) (*github.RepositoryRelease, error) {
	release, _, err := a.client.Repositories.GetReleaseByTag(ctx, a.org, a.repo, tag)
	return release, err
}

type AssetHelper interface {
	GetReleaseWithAsset(ctx context.Context, tag string, expectedAssetName string) (*github.RepositoryRelease, error)
	DownloadAsset(downloadUrl string, destFile string) error
}

type assetHelper struct {
	client     GithubAssetClient
	httpClient *http.Client
}

func NewAssetHelper(client GithubAssetClient) AssetHelper {
	return &assetHelper{
		client: client,
		httpClient: &http.Client{
			Timeout: DefaultRequestTimeout,
		},
	}
}

func (a *assetHelper) GetReleaseWithAsset(ctx context.Context, tag string, expectedAssetName string) (*github.RepositoryRelease, error) {
	if tag != "latest" {
		return a.client.GetReleaseByTag(ctx, tag)
	}
	// don't use latest tag, because that might not have the assets yet if the release build is running.
	listOpts := github.ListOptions{PerPage: 10}
	releases, err := a.client.ListReleases(ctx, &listOpts)
	if err != nil {
		return nil, err
	}
	for _, release := range releases {
		if TryGetAssetWithName(release, expectedAssetName) != nil {
			return release, nil
		}
	}
	return nil, NoReleaseContainingAssetError
}

func (a *assetHelper) DownloadAsset(downloadUrl string, destFile string) error {
	res, err := a.httpClient.Get(downloadUrl)
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

func TryGetAssetWithName(release *github.RepositoryRelease, expectedAssetName string) *github.ReleaseAsset {
	for _, asset := range release.Assets {
		if asset.GetName() == expectedAssetName {
			return &asset
		}
	}
	return nil
}
