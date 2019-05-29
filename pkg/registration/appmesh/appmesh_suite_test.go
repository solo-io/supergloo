package appmesh

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	testNamespace            = "supergloo-system"
	sidecarInjectorImageName = "quay.io/solo-io/sidecar-injector:0.1.2"
)

var T *testing.T

func TestAppmesh(t *testing.T) {
	RegisterFailHandler(Fail)
	T = t
	RunSpecs(t, "AWS App Mesh Registration Suite")
}
