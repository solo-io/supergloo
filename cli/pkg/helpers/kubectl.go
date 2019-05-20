package helpers

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

func KubectlApply(manifest string) error {
	return kubectl(bytes.NewBufferString(manifest), "apply", "-f", "-")
}

func kubectl(stdin io.Reader, args ...string) error {
	kubectl := exec.Command("kubectl", args...)
	if stdin != nil {
		kubectl.Stdin = stdin
	}
	kubectl.Stdout = os.Stdout
	kubectl.Stderr = os.Stderr
	return kubectl.Run()
}
