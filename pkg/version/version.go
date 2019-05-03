package version

var (
	UndefinedVersion = "undefined"

	// default version set if running without setting TAGGED_VERSION in env
	DevVersion = "dev"

	// These will be set by the linker during build
	Version         = UndefinedVersion
	ImageRepoPrefix = UndefinedVersion
)

func IsReleaseVersion() bool {
	return Version != UndefinedVersion && Version != DevVersion
}
