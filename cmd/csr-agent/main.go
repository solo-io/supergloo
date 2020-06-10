package main

import (
	"context"

	csr_agent "github.com/solo-io/service-mesh-hub/pkg/csr-agent"
)

func main() {
	ctx := context.Background()
	csr_agent.Run(ctx)
}
