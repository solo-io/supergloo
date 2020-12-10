package utils

import (
	"istio.io/istio/pkg/config/host"
)

// Return the intersection of two lists of hostnames.
func CommonHostnames(userHostnames, translatedHostnames []string) host.Names {
	user := host.NewNames(userHostnames)
	translated := host.NewNames(translatedHostnames)
	return user.Intersection(translated)
}
