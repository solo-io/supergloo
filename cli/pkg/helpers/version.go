package helpers

import (
	"context"
	"net/http"
	"strings"

	"github.com/pkg/errors"

	"github.com/google/go-github/github"
)

func GetLatestVersion(repo string) (string, error) {
	client := GetPublicGithubClient()
	release, _, err := client.Repositories.GetLatestRelease(context.TODO(), "solo-io", repo)
	if err != nil {
		return "", errors.Wrapf(err, "unable to get latest version for %s", repo)
	}
	return release.GetTagName()[1:], nil
}

func IsValidVersion(repo, version string) (string, error) {
	version = "v" + strings.TrimPrefix(version, "v")
	client := GetPublicGithubClient()
	release, _, err := client.Repositories.GetReleaseByTag(context.TODO(), "solo-io", repo, version)
	if err != nil {
		return "", errors.Wrapf(err, "%s is not a valid version for %s", version, repo)
	}
	return release.GetTagName()[1:], nil
}

func GetPublicGithubClient() *github.Client {
	return github.NewClient(http.DefaultClient)
}
