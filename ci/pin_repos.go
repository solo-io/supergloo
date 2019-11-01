package main

import (
	"github.com/solo-io/go-utils/log"
	version "github.com/solo-io/go-utils/versionutils"
	md_version "github.com/solo-io/mesh-projects/pkg/version"
)

func main() {
	tomlTree, err := version.ParseToml()
	fatalCheck(err, "parsing error")

	soloKitVersion, err := version.GetVersion(md_version.SoloKitPkg, tomlTree)
	fatalCheck(err, "getting solo-kit version")
	fatalCheck(version.PinGitVersion("../solo-kit", soloKitVersion), "consider git fetching in solo-kit repo")

	glooVersion, err := version.GetVersion(md_version.GlooPkg, tomlTree)
	fatalCheck(err, "getting gloo version")
	fatalCheck(version.PinGitVersion("../gloo", glooVersion), "consider git fetching in gloo repo")
}

func fatalCheck(err error, msg string) {
	if err != nil {
		log.Fatalf("Error (%v) unable to pin repos!: %v", msg, err)
	}
}
