package kubectl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var isAlpha = true

func init() {
	readerChan, _, err := testutils.KubectlOutChan(&bytes.Buffer{}, "version", "--client=true", "-o", "json")
	if err != nil {
		panic(err)
	}
	select {
	case <-time.After(time.Minute):
		panic("timeout")
	case reader := <-readerChan:
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			panic(err)
		}
		cv := struct {
			ClientVersion struct {
				Minor string `json:"minor"`
			} `json:"clientVersion"`
		}{}
		json.Unmarshal(data, &cv)

		minor, _ := strconv.Atoi(cv.ClientVersion.Minor)
		if minor >= 21 {
			isAlpha = false
		}
	}
}

func CurlWithEphemeralPod(ctx context.Context, kubecontext, fromns, frompod string, args ...string) string {

	createargs := []string{"debug", "--quiet",
		"--image=curlimages/curl@sha256:aa45e9d93122a3cfdf8d7de272e2798ea63733eeee6d06bd2ee4f2f8c4027d7c",
		"--container=curl-ephemeral", frompod, "-n", fromns}
	if isAlpha {
		createargs = append([]string{"alpha"}, createargs...)
	} else {
		createargs = append(createargs, "--image-pull-policy=IfNotPresent")
	}
	createargs = append(createargs, "--", "sleep", "10h")

	// Execute curl commands from the same pod each time to avoid creating a burdensome number of ephemeral pods.
	// create the curl pod; we do this every time and it will only work the first time, so ignore failures
	executeNoFail(ctx, kubecontext, createargs...)
	// note, we use sudo so that the curl is not from the istio-proxy user. we dont really need root.
	args = append([]string{"exec",
		"--container=curl-ephemeral", frompod, "-n", fromns, "--", "curl", "--connect-timeout", "1", "--max-time", "5"}, args...)
	return execute(ctx, kubecontext, args...)
}

func WaitForRollout(ctx context.Context, kubecontext, ns, deployment string) {
	args := []string{"-n", ns, "rollout", "status", "deployment", deployment}
	execute(ctx, kubecontext, args...)
}

func Curl(ctx context.Context, kubecontext, ns, fromDeployment, fromContainer, url string) string {
	args := []string{
		"-n", ns,
		"exec", fmt.Sprintf("deployment/%s", fromDeployment),
		"-c", fromContainer,
		"--", "curl", url,
	}
	return execute(ctx, kubecontext, args...)
}

func DeployBookInfo(ctx context.Context, kubeContext, ns string) {
	args := []string{"--namespace", ns, "apply", "-f", "../../ci/bookinfo.yaml"}
	out := execute(ctx, kubeContext, args...)
	fmt.Fprintln(GinkgoWriter, out)
}

func CreateNamespace(ctx context.Context, kubeContext, ns string) {
	args := []string{"create", "namespace", ns}
	out := execute(ctx, kubeContext, args...)
	fmt.Fprintln(GinkgoWriter, out)
}

func DeleteNamespace(ctx context.Context, kubeContext, ns string) {
	args := []string{"delete", "namespace", ns}
	out := execute(ctx, kubeContext, args...)
	fmt.Fprintln(GinkgoWriter, out)
}

func LabelNamespace(ctx context.Context, kubeContext, ns, label string) {
	args := []string{"label", "namespace", ns, label}
	out := execute(ctx, kubeContext, args...)
	fmt.Fprintln(GinkgoWriter, out)
}

func SetDeploymentEnvVars(
	ctx context.Context,
	kubeContext string,
	ns string,
	deploymentName string,
	containerName string,
	envVars map[string]string,
) {
	var envVarStrings []string
	for name, value := range envVars {
		envVarStrings = append(envVarStrings, fmt.Sprintf("%s=%s", name, value))
	}
	args := append([]string{"set", "env", "-n", ns, fmt.Sprintf("deployment/%s", deploymentName), "-c", containerName}, envVarStrings...)
	out := execute(ctx, kubeContext, args...)
	fmt.Fprintln(GinkgoWriter, out)
}

func DisableContainer(
	ctx context.Context,
	kubeContext string,
	ns string,
	deploymentName string,
	containerName string,
) {
	args := append([]string{
		"-n", ns,
		"patch", "deployment", deploymentName,
		"--patch",
		fmt.Sprintf("{\"spec\": {\"template\": {\"spec\": {\"containers\": [{\"name\": \"%s\",\"command\": [\"sleep\", \"20h\"]}]}}}}",
			containerName),
	})
	out := execute(ctx, kubeContext, args...)
	fmt.Fprintln(GinkgoWriter, out)
}

func EnableContainer(
	ctx context.Context,
	kubeContext string,
	ns string,
	deploymentName string,
) {
	args := append([]string{
		"-n", ns,
		"patch", "deployment", deploymentName,
		"--type", "json",
		"-p", "[{\"op\": \"remove\", \"path\": \"/spec/template/spec/containers/0/command\"}]",
	})
	out := execute(ctx, kubeContext, args...)
	fmt.Fprintln(GinkgoWriter, out)
}

func execute(ctx context.Context, kubeContext string, args ...string) string {
	data, err := executeNoFail(ctx, kubeContext, args...)
	Expect(err).NotTo(HaveOccurred())
	return data
}

func executeNoFail(ctx context.Context, kubeContext string, args ...string) (string, error) {
	args = append([]string{"--context", kubeContext}, args...)
	fmt.Fprintf(GinkgoWriter, "Executing: kubectl %v \n", args)
	readerChan, done, err := testutils.KubectlOutChan(&bytes.Buffer{}, args...)
	if err != nil {
		return "", err
	}
	defer close(done)
	select {
	case <-ctx.Done():
		return "", nil
	case reader := <-readerChan:
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			return "", err
		}
		fmt.Fprintf(GinkgoWriter, "<kubectl %v> output: %v\n", args, string(data))
		return string(data), nil
	}
}
