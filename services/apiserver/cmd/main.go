package main

import (
	"context"

	apiserver "github.com/solo-io/service-mesh-hub/services/apiserver/pkg"
)

func main() {
	ctx := context.Background()
	apiserver.Run(ctx)
}
