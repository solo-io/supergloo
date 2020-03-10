package version

import (
	"github.com/solo-io/mesh-projects/cli/pkg/common"
)

var (
	UndefinedVersion = "undefined"

	// Will be set by the linker during build. Does not include "v" prefix.
	Version string
)

func init() {
	/*
		We were using the linker incorrectly previously. From the docs: https://golang.org/cmd/link/

		-X importpath.name=value
			Set the value of the string variable in importpath named name to value.
			This is only effective if the variable is declared in the source code either uninitialized
			or initialized to a constant string expression.....
	*/
	if Version == "" {
		Version = UndefinedVersion
	}
}

func IsReleaseVersion() bool {
	return Version != UndefinedVersion && common.ValidReleaseSemver(Version)
}
