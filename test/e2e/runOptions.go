package e2e

const (
	// RunE2E if set allows the E2E tests to be run, otherwise they are skipped
	RunE2E = "RUN_E2E"
	// UseExisting if set to "1" means the the current tests should not create a new setup, but rather use an existing one
	UseExisting = "USE_EXISTING"
	// NoCleanup indicates to the test suite that the environment should not be brought down after the tests end
	NoCleanup = "NO_CLEANUP"
	// LimitedTrust if set means that the current test suite should run in limited trust test mode
	LimitedTrust = "LIMITED_TRUST"
)
