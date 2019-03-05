package main

import (
	"github.com/solo-io/go-utils/docsutils"
)

func main() {
	spec := docsutils.DocsPRSpec{
		Owner:           "solo-io",
		Repo:            "supergloo",
		Product:         "supergloo",
		ChangelogPrefix: "supergloo",
		ApiPaths: []string{
			"docs/v1/github.com/solo-io/supergloo",
			"docs/v1/github.com/solo-io/solo-kit",
			"docs/v1/gogoproto",
			"docs/v1/google",
		},
		CliPrefix: "supergloo",
	}
	docsutils.PushDocsCli(&spec)
}
