package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

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
func LinkerdInject(input string) (string, error) {
	cmd := exec.Command("linkerd", "inject", "-")
	cmd.Stdin = bytes.NewBuffer([]byte(input))
	output := &bytes.Buffer{}
	cmd.Stdout = output
	cmd.Stderr = output
	err := cmd.Run()
	if err != nil {
		return "", errors.Wrapf(err, "linkerd inject failed: %v", output.String())
	}
	return output.String(), nil
}

func KubectlApply(namespace, yamlStr string) error {
	return Kubectl(bytes.NewBuffer([]byte(yamlStr)), "apply", "-n", namespace, "-f", "-")
}

func KubectlDelete(namespace, yamlStr string) error {
	return Kubectl(bytes.NewBuffer([]byte(yamlStr)), "delete", "-n", namespace, "--ignore-not-found=true", "-f", "-")
}

func Kubectl(stdin io.Reader, args ...string) error {
	return KubectlCtx(nil, stdin, args...)
}

func KubectlPortForward(ctx context.Context, namespace, deployment string, port int) error {
	log.Printf("starting port forward on %v.%v:%v", namespace, deployment, port)
	return KubectlCtx(ctx, nil, "port-forward", "-n", namespace, "deployment/"+deployment, fmt.Sprintf("%v", port))
}

func KubectlCtx(ctx context.Context, stdin io.Reader, args ...string) error {
	kubectl := exec.Command("kubectl", args...)
	if stdin != nil {
		kubectl.Stdin = stdin
	}
	kubectl.Stdout = os.Stdout
	kubectl.Stderr = os.Stderr
	if ctx != nil {
		go func() {
			defer ginkgo.GinkgoRecover()
			err := kubectl.Run()
			select {
			case <-ctx.Done():
				return
			default:
				gomega.Expect(err).NotTo(gomega.HaveOccurred())
			}
		}()
		go func() {
			<-ctx.Done()
			if kubectl.Process != nil {
				kubectl.Process.Kill()
			}
		}()
		return nil
	}
	return kubectl.Run()
}
