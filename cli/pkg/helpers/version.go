package helpers

import (
	"context"
	"net/http"

	"github.com/google/go-github/github"
)

func GetLatestVersion() (string, error) {
	client := GetPublicGithubClient()
	release, _, err := client.Repositories.GetLatestRelease(context.TODO(), "solo-io", "supergloo")
	if err != nil {
		return "", err
	}
	return release.GetTagName()[1:], nil
}

func GetPublicGithubClient() *github.Client {
	return github.NewClient(http.DefaultClient)
}
