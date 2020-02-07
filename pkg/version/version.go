package version

import (
	cli_util "github.com/solo-io/mesh-projects/cli/pkg/util"
)

var (
	UndefinedVersion = "undefined"

	// Will be set by the linker during build. Does not include "v" prefix.
	Version = UndefinedVersion
)

func IsReleaseVersion() bool {
	return Version != UndefinedVersion && cli_util.ValidReleaseSemver(Version)
}
