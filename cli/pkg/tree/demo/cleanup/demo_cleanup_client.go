package cleanup

import (
	"github.com/solo-io/service-mesh-hub/cli/pkg/common/exec"
)

func IstioMulticlusterCleanup(runner exec.Runner) error {
	return runner.Run("bash", "-c", istioMulticlusterCleanupScript)
}

const (
	istioMulticlusterCleanupScript = `
kind get clusters | grep -E  '(management-plane|remote-cluster)-[a-z0-9]+' | while read -r r; do kind delete cluster --name "$r"; done
exit 0
`
)
