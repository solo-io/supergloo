package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/solo-io/solo-kit/test/helpers"

	. "github.com/onsi/ginkgo"
)

func TestE2E(t *testing.T) {
	if os.Getenv("RUN_E2E") == "" {
		return
	}
	helpers.RegisterCommonFailHandlers()
	helpers.RegisterPreFailHandler(GetEnv().DumpState)
	helpers.SetupLog()
	RunSpecs(t, "E2e Suite")
}

var eksNamespace string

var _ = BeforeSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	/* env := */ StartEnvOnce(ctx)
	// TODO: deploy test helper?
	eksNamespace = setupAppmeshEksEnvironment()
})

var _ = AfterSuite(func() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	ClearEnv(ctx)
	cleanupAppmeshEksEnvironment(eksNamespace)
})
