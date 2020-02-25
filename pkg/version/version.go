package version

import (
	"github.com/solo-io/mesh-projects/cli/pkg/common"
)

var (
	UndefinedVersion = "undefined"

	// Will be set by the linker during build. Does not include "v" prefix.
	Version = UndefinedVersion
)

func IsReleaseVersion() bool {
	return Version != UndefinedVersion && common.ValidReleaseSemver(Version)
}
