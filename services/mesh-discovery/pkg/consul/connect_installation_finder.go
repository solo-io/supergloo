package consul

import (
	"regexp"
	"strings"

	consulconfig "github.com/hashicorp/consul/agent/config"
	"github.com/hashicorp/hcl"
	globalcommon "github.com/solo-io/mesh-projects/services/common"
	kubev1 "k8s.io/api/core/v1"
)

var (
	// a consul's invocation can include a line like:
	// -hcl="connect { enabled = true }"
	// hcl is HashiCorp configuration Language
	// https://github.com/hashicorp/hcl
	hclRegex = regexp.MustCompile("-hcl=\"([^\"]*)\"")
)

//go:generate mockgen -destination ./mocks/mock_connect_deployment_finder.go github.com/solo-io/mesh-projects/services/mesh-discovery/pkg/consul ConsulConnectInstallationFinder

// check whether a container from the cluster represents an instance of Consul Connect
type ConsulConnectInstallationFinder interface {
	// returns true if:
	//  * the container is running an image named "consul"
	//  * the container is running with `-server`
	//  * the container is running with Consul Connect enabled through an `-hcl` flag
	IsConsulConnect(kubev1.Container) (isConsulConnect bool, err error)
}

func NewConsulConnectInstallationFinder(imageNameParser globalcommon.ImageNameParser) ConsulConnectInstallationFinder {
	return &consulConnectInstallationFinder{imageNameParser}
}

type consulConnectInstallationFinder struct {
	imageNameParser globalcommon.ImageNameParser
}

func (c *consulConnectInstallationFinder) IsConsulConnect(container kubev1.Container) (isConsulConnect bool, err error) {
	parsedImage, err := c.imageNameParser.Parse(container.Image)
	if err != nil {
		return false, InvalidImageFormatError(err, container.Image)
	}

	// if the image appears to be a consul image, and
	// the container is starting up with a "-server" arg,
	// then declare that we've found consul
	if parsedImage.Path != normalizedConsulImagePath {
		return false, nil
	}

	cmd := strings.Join(container.Command, " ")

	isServerMode := strings.Contains(cmd, consulServerArg)
	if !isServerMode {
		return false, nil
	}

	hclMatches := hclRegex.FindStringSubmatch(cmd)
	if len(hclMatches) < 2 {
		return false, nil
	}

	config := &consulconfig.Config{}
	err = hcl.Decode(config, hclMatches[1])
	if err != nil {
		return false, HclParseError(err, hclMatches[1])
	}

	return config.Connect.Enabled != nil && *config.Connect.Enabled, nil
}
