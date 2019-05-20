package helpers

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

//go:generate mockgen -destination=./mocks/kubectl.go -source kubectl.go -package mocks

type Kubectl interface {
	ApplyManifest(manifest string) error
}

type kubectl struct{}

func NewKubectl() Kubectl {
	return &kubectl{}
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
