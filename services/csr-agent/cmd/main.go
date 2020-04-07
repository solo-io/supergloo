package main

import (
	"context"

	csr_agent "github.com/solo-io/service-mesh-hub/services/csr-agent"
)

func main() {
	ctx := context.Background()
	csr_agent.Run(ctx)
}
