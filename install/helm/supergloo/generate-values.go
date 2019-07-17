package main

import (
	"os"

	"github.com/solo-io/go-utils/log"
	"github.com/solo-io/supergloo/install/helm/supergloo/generate"
)

func main() {
	var version, imageTag, imageRepoPrefix string
	if len(os.Args) < 4 {
		panic("Must provide 3 arguments: version, imageTag, and imageRepoPrefix")
	} else {
		version, imageTag, imageRepoPrefix = os.Args[1], os.Args[2], os.Args[3]
	}
	log.Printf("Generating helm files. Version: %s, ImageTag: %s, ImageRepoPrefix: %s", version, imageTag, imageRepoPrefix)
	if err := generate.Run(version, imageTag, imageRepoPrefix); err != nil {
		panic(err)
	}
}
