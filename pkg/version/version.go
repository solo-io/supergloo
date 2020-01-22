package version

var (
	UndefinedVersion = "undefined"

	// default version set if running without setting TAGGED_VERSION in env
	DevVersion = "dev"

	// Will be set by the linker during build
	Version = UndefinedVersion
)

func IsReleaseVersion() bool {
	return Version != UndefinedVersion && Version != DevVersion
}

const (
	MeshBridgeAppName    = "mesh-bridge"
	MeshDiscoveryAppName = "mesh-discovery"
	MeshConfigAppName    = "mesh-config"
)
