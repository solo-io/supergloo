package helpers

import (
	"context"
	"net/http"
	"os"

	"golang.org/x/oauth2"

	"github.com/google/go-github/github"
)

func GetLatestVersion(ctx context.Context) (string, error) {
	client := GetPublicGithubClient(ctx)
	release, _, err := client.Repositories.GetLatestRelease(ctx, "solo-io", "supergloo")
	if err != nil {
		return "", err
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
