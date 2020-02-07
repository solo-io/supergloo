package cli_util

import (
	"strings"

	"github.com/solo-io/go-utils/versionutils"
)

func ValidReleaseSemver(v string) bool {
	return versionutils.MatchesRegex(suffixVersion(v))
}

// prepend 'v' suffix to semver string if it doesn't exist
func suffixVersion(v string) string {
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return v
}
