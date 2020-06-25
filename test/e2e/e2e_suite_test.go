package e2e

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/solo-io/solo-kit/test/helpers"

	. "github.com/onsi/ginkgo"
)

func TestE2E(t *testing.T) {
	if os.Getenv("RUN_E2E") == "" {
		fmt.Println("skipping E2E tests")
		return
	}
	helpers.RegisterCommonFailHandlers()
	helpers.RegisterPreFailHandler(GetEnv().DumpState)
	helpers.SetupLog()
	RunSpecs(t, "E2e Suite")
}

var eksNamespace string

func RunEKS() bool {
	// allow disabling EKS tests explicitly to allow running istio tests locally
	return os.Getenv("RUN_EKS") != "0"
}

var _ = BeforeSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	/* env := */ StartEnvOnce(ctx)
	// TODO: deploy test helper?
	if RunEKS() {
		eksNamespace = setupAppmeshEksEnvironment()
	}
})

var _ = AfterSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if RunEKS() {
		cleanupAppmeshEksEnvironment(eksNamespace)
	}
	ClearEnv(ctx)
})
