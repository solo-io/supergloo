package docsgen

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/solo-io/go-utils/versionutils"
)

type VersionInfo struct {
	// LatestVersion is the latest stable version
	LatestVersion string `json:"latest_version"`

	// SupportedVersions is a list of all versions to build doc sites for.
	SupportedVersions []string `json:"supported_versions"`
}

// BuildVersionInfo uses git tags to build meta information about the versions.
func BuildVersionInfo(repo *git.Repository, minVersion *versionutils.Version) (*VersionInfo, error) {
	tagRefs, err := repo.Tags()
	if err != nil {
		return nil, err
	}

	latestVersion := versionutils.Zero()
	// latest patch version for each minor version
	minorVersions := make(map[string]versionutils.Version) // key is <major>.<minor>
	if err := tagRefs.ForEach(func(tagRef *plumbing.Reference) error {
		version, err := versionutils.ParseVersion(tagRef.Name().Short())
		if err != nil {
			return nil // skip invalid tags by ignoring the error
		}

		// filter out unsupported versions (less than the minimum supported version)
		if minVersion.MustIsGreaterThan(*version) {
			return nil
		}

		// filter out tags that do not have a changelog entry
		if _, err := os.Stat(filepath.Join("changelog", version.String())); os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return err
		}

		// find the latest stable release by ignoring labeled releases (e.g. beta releases)
		if version.Label == "" && version.MustIsGreaterThan(latestVersion) {
			latestVersion = *version
		}

		versionKey := fmt.Sprintf("%d.%d", version.Major, version.Minor)
		if maxSeenPatch, ok := minorVersions[versionKey]; !ok || version.MustIsGreaterThan(maxSeenPatch) {
			minorVersions[versionKey] = *version
		}

		return nil
	}); err != nil {
		return nil, err
	}

	supportedVersions := make([]versionutils.Version, 0, len(minorVersions))
	for _, version := range minorVersions {
		supportedVersions = append(supportedVersions, version)
	}
	sortVersionList(supportedVersions)

	return &VersionInfo{
		LatestVersion:     latestVersion.String(),
		SupportedVersions: append([]string{"main"}, stringifyVersionList(supportedVersions)...),
	}, nil
}

func sortVersionList(versions []versionutils.Version) {
	for i := 1; i < len(versions); i++ {
		for j := i; j > 0; j-- {
			if versions[j].MustIsGreaterThan(versions[j-1]) {
				versions[j-1], versions[j] = versions[j], versions[j-1]
			}
		}
	}
}

func stringifyVersionList(versions []versionutils.Version) []string {
	versionStrs := make([]string, len(versions))
	for i, version := range versions {
		versionStrs[i] = version.String()
	}

	return versionStrs
}
