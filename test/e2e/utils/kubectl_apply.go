package utils

import (
	"bytes"
	"context"
	"fmt"
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
	return KubectlCtx(nil, stdin, args...)
}

func KubectlPortForward(ctx context.Context, namespace, deployment string, port int) error {
	return KubectlCtx(ctx, nil, "port-forward", "-n", namespace, "deployment/"+deployment, fmt.Sprintf("%v", port))
}

func KubectlCtx(ctx context.Context, stdin io.Reader, args ...string) error {
	kubectl := exec.Command("kubectl", args...)
	if ctx != nil {
		kubectl = exec.CommandContext(ctx, "kubectl", args...)
	}
	if stdin != nil {
		kubectl.Stdin = stdin
	}
	kubectl.Stdout = os.Stdout
	kubectl.Stderr = os.Stderr
	return kubectl.Start()
}
