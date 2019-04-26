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
	defaultTag := "latest"
	if IsReleaseVersion() {
		defaultTag = Version
	}
	return defaultTag
}
