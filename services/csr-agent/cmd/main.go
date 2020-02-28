package main

import (
	"context"

	csr_agent "github.com/solo-io/mesh-projects/services/csr-agent"
)

func main() {
	ctx := context.Background()
	csr_agent.Run(ctx)
}
