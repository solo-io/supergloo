package version

import (
	"github.com/solo-io/mesh-projects/pkg/project"
)

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

	BaseImageRepoName    = "quay.io/solo-io/mc-base-image"
	BaseImageRepoVersion = "0.0.2"
)

var GoBinarySummary = []*project.GoBinaryOutline{{
	BinaryNameBase:       MeshBridgeAppName,
	ImageName:            "mc-mesh-bridge",
	BinaryDir:            "services/mesh-bridge/cmd",
	DockerOutputFilepath: "services/mesh-bridge/cmd/Dockerfile",
}, {
	BinaryNameBase:       MeshDiscoveryAppName,
	ImageName:            "mc-mesh-discovery",
	BinaryDir:            "services/mesh-discovery/cmd",
	DockerOutputFilepath: "services/mesh-discovery/cmd/Dockerfile",
}, {
	BinaryNameBase:       MeshConfigAppName,
	ImageName:            "mc-mesh-config",
	BinaryDir:            "services/mesh-config/cmd",
	DockerOutputFilepath: "services/mesh-config/cmd/Dockerfile",
}}
