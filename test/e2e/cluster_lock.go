package e2e

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type clusterLock struct {
	started chan struct{}
}

func (c *clusterLock) NeedLeaderElection() bool {
	return true
}
func (c *clusterLock) Start(stop <-chan struct{}) error {
	close(c.started)
	<-stop
	return nil
}

// Blocks on acquiring a cluster lock shared among all ongoing e2e tests for the cluster referenced by the config.
// If ClusterLockTimeout duration is exceeded, the test fails.
func WaitForClusterLock(
	ctx context.Context,
	config *rest.Config,
	timeout time.Duration,
) {
	mgr, err := manager.New(config, manager.Options{
		LeaderElection:          true,
		LeaderElectionNamespace: "default",
		LeaderElectionID:        "clusterlock",
		// disable metrics
		HealthProbeBindAddress: "0",
		MetricsBindAddress:     "0",
	})
	Expect(err).ToNot(HaveOccurred())
	errc := make(chan error)
	go func() {
		errc <- mgr.Start(ctx.Done())
	}()

	lock := &clusterLock{started: make(chan struct{})}
	err = mgr.Add(lock)
	Expect(err).NotTo(HaveOccurred())
	fmt.Fprintln(GinkgoWriter, "Waiting to acquire cluster lock...")
	select {
	// This ctx will never be cancelled manually, instead the test will simply end, allowing the leader lock to be acquired by another test.
	case <-ctx.Done():
		fmt.Fprintln(GinkgoWriter, "Releasing cluster lock")
	case <-time.After(timeout):
		Fail("Timed out waiting for cluster lock")
	case <-lock.started:
		fmt.Fprintln(GinkgoWriter, "Acquired cluster lock")
	}
}
