package e2e_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/solo-io/service-mesh-hub/test/e2e"
)

// to skip testing this package, run `make run-tests SKIP_PACKAGES=test/e2e
// to test only this package, run `make run-tests TEST_PKG=test/e2e
func TestE2e(t *testing.T) {
	RegisterFailHandler(func(message string, callerSkip ...int) {
		runShell("kubectl logs -n=service-mesh-hub -l app=service-mesh-hub")
		runShell("kubectl logs -n=service-mesh-hub -l app=service-mesh-hub --previous")
		runShell("kubectl get virtualservices.networking.istio.io -A -oyaml")
		runShell("kubectl get apiproducts -A -oyaml")
		runShell("kubectl get pod -A")
		runShell("kubectl logs -n istio-system $(kubectl get pod -n istio-system | grep istiod | awk '{print $1}')")
		runShell("kubectl logs -n istio-system $(kubectl get pod -n istio-system | grep istio-ingressgateway | awk '{print $1}')")
		runShell("istioctl proxy-config route $(kubectl get pod -n istio-system | grep istio-ingressgateway | awk '{print $1}').istio-system -ojson")
		Fail(message, callerSkip...)
	})
	RunSpecs(t, "E2e Suite")
}

func runShell(c string) {
	buf := &bytes.Buffer{}
	cmd := exec.Command("sh", "-c", c)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(GinkgoWriter, "[%v] command FAILED: %v", c, err)
		return
	}
	fmt.Fprintf(GinkgoWriter, "[%v] command result: \n%v", c, buf.String())
}

var _ = BeforeSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	/* env := */ StartEnvOnce(ctx)
})

var _ = AfterSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if os.Getenv("NO_CLEANUP") != "" {
		return
	}
	_ = ClearEnv(ctx)
})
