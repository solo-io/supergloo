package version

var (
	UndefinedVersion = "undefined"

	// default version set if running without setting TAGGED_VERSION in env
	DevVersion = "dev"

	// This will be set by the linker during build
	Version = UndefinedVersion
)

func IsReleaseVersion() bool {
	return Version != UndefinedVersion && Version != DevVersion
}

func GetWebhookImageTag() string {
	// TODO(marco): temporarily default to a manually tagged image to make tests pass. Each build will soon build and push its own images.
	defaultTag := "temporary-for-tests"
	if IsReleaseVersion() {
		defaultTag = Version
	}
	return defaultTag
}
