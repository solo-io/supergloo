package helpers

import (
	"context"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"

	"github.com/pkg/errors"

	"github.com/google/go-github/github"
)

func GetLatestVersion(ctx context.Context, repo string) (string, error) {
	client := GetPublicGithubClient(ctx)
	release, _, err := client.Repositories.GetLatestRelease(ctx, "solo-io", repo)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get latest version for %s", repo)
	}
	return release.GetTagName()[1:], nil
}

func IsValidVersion(ctx context.Context, repo, version string) (string, error) {
	version = "v" + strings.TrimPrefix(version, "v")
	client := GetPublicGithubClient(ctx)
	release, _, err := client.Repositories.GetReleaseByTag(ctx, "solo-io", repo, version)
	if err != nil {
		return "", errors.Wrapf(err, "%s is not a valid version for %s", version, repo)
	}
	return release.GetTagName()[1:], nil
}

func GetPublicGithubClient(ctx context.Context) *github.Client {
	client := http.DefaultClient
	if githubToken := os.Getenv("GITHUB_TOKEN"); githubToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: githubToken},
		)
		client = oauth2.NewClient(ctx, ts)
	}
	return github.NewClient(client)
}
