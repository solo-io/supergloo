package version

import "context"

type DeployedVersionFinder interface {
	// find the open source version without a leading "v" prefix
	OpenSourceVersion(ctx context.Context, installNamespace string) (string, error)
}
