package version

import "context"

//go:generate mockgen -source ./interfaces.go -destination ./mocks/mock_interfaces.go

type DeployedVersionFinder interface {
	// find the open source version without a leading "v" prefix
	OpenSourceVersion(ctx context.Context, installNamespace string) (string, error)
}
