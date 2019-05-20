package helpers

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/solo-io/supergloo/cli/pkg/helpers/mocks"
)

//go:generate mockgen -destination=./mocks/kubectl.go -source kubectl.go -package mocks

type Kubectl interface {
	ApplyManifest(manifest string) error
}

type kubectl struct{}

var kubectlInst Kubectl

func NewKubectl() Kubectl {
	// For testing
	if kubectlInst != nil {
		return kubectlInst
	}
	return &kubectl{}
}

func SetKubectlMock(mockKubectl *mocks.MockKubectl) {
	kubectlInst = mockKubectl
}

func (k *kubectl) ApplyManifest(manifest string) error {
	return kubectlApply(bytes.NewBufferString(manifest), "apply", "-f", "-")
}

func kubectlApply(stdin io.Reader, args ...string) error {
	kubectl := exec.Command("kubectl", args...)
	if stdin != nil {
		kubectl.Stdin = stdin
	}
	kubectl.Stdout = os.Stdout
	kubectl.Stderr = os.Stderr
	return kubectl.Run()
}
