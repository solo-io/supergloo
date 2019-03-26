package utils

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/solo-io/go-utils/errors"
)

func IstioInject(istioNamespace, input string) (string, error) {
	cmd := exec.Command("istioctl", "kube-inject", "-i", istioNamespace, "-f", "-")
	cmd.Stdin = bytes.NewBuffer([]byte(input))
	output := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Run()
	if err != nil {
		return "", errors.Wrapf(err, "kube inject failed: %v", output.String())
	}
	return output.String(), nil
}

func KubectlApply(namespace, yamlStr string) error {
	return Kubectl(bytes.NewBuffer([]byte(yamlStr)), "apply", "-n", namespace, "-f", "-")
}

func KubectlDelete(namespace, yamlStr string) error {
	return Kubectl(bytes.NewBuffer([]byte(yamlStr)), "delete", "-n", namespace, "-f", "-")
}

func Kubectl(stdin io.Reader, args ...string) error {
	kubectl := exec.Command("kubectl", args...)
	if stdin != nil {
		kubectl.Stdin = stdin
	}
	kubectl.Stdout = os.Stdout
	kubectl.Stderr = os.Stderr
	return kubectl.Run()
}
