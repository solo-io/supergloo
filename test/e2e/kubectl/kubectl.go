package kubectl

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

	"github.com/solo-io/go-utils/testutils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Curl(ctx context.Context, kubecontext, fromns, frompod string, args ...string) string {
	// note, we use sudo so that the curl is not from the istio-proxy user. we dont really need root.
	args = append([]string{"alpha", "debug", "--quiet",
		"--image=curlimages/curl@sha256:aa45e9d93122a3cfdf8d7de272e2798ea63733eeee6d06bd2ee4f2f8c4027d7c",
		"-i", frompod, "-n", fromns, "--", "curl", "--connect-timeout", "1", "--max-time", "5"}, args...)
	return execute(ctx, kubecontext, args...)
}

func DeployBookInfo(ctx context.Context, kubeContext, ns string) string {
	args := []string{"--namespace", ns, "apply", "-f", "../../ci/bookinfo.yaml"}
	return execute(ctx, kubeContext, args...)
}

func CreateNamespace(ctx context.Context, kubeContext, ns string) string {
	args := []string{"create", "namespace", ns}
	return execute(ctx, kubeContext, args...)
}

func DeleteNamespace(ctx context.Context, kubeContext, ns string) string {
	args := []string{"delete", "namespace", ns}
	return execute(ctx, kubeContext, args...)
}

func LabelNamespace(ctx context.Context, kubeContext, ns, label string) string {
	args := []string{"kubectl", "label", "namespace", ns, label}
	return execute(ctx, kubeContext, args...)
}

func execute(ctx context.Context, kubeContext string, args ...string) string {
	args = append([]string{"--context", kubeContext}, args...)
	fmt.Fprintf(GinkgoWriter, "Executing: kubectl %v", args)
	readerChan, done, err := testutils.KubectlOutChan(&bytes.Buffer{}, args...)
	Expect(err).NotTo(HaveOccurred())
	defer close(done)
	select {
	case <-ctx.Done():
		return ""
	case reader := <-readerChan:
		data, err := ioutil.ReadAll(reader)
		Expect(err).NotTo(HaveOccurred())
		return string(data)
	}
}
