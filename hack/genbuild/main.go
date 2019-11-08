package main

import (
	"log"

	"github.com/solo-io/mesh-projects/pkg/project"

	"github.com/solo-io/mesh-projects/hack/genbuild/generator"
	"github.com/solo-io/mesh-projects/pkg/version"
)

func main() {

	err := run()
	if err != nil {
		log.Fatalf("unable to run %v", err)
	}

}

func run() error {
	genOpts := &generator.GenBuildOptions{
		GlobalOptions: &project.GlobalOptions{
			BaseImageRepo:    version.BaseImageRepoName,
			BaseImageVersion: version.BaseImageRepoVersion,
		},
		OutputDir:  "_output/gen",
		GoBinaries: version.GoBinarySummary,
		ServiceManifestOutlines: []*generator.ServiceManifestOutline{{
			AppGroup: "smMarketplace",
			AppName:  "meshConfig",
		}},
	}
	return generator.Generate(genOpts)
}
