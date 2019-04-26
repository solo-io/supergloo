package appmesh

import (
	"context"
	"sync"

	"github.com/solo-io/go-utils/errors"
)

type Resources struct {
	VirtualNodes, VirtualServices []string
	VirtualRouters                map[string][]string
}

func ListAllForMesh(ctx context.Context, client Client, meshName string) (*Resources, error) {
	var (
		requestWg        sync.WaitGroup
		vnNames, vsNames []string
		vrMap            = sync.Map{}
		errorChan        = make(chan error, 10)
	)

	// Virtual Nodes
	requestWg.Add(1)
	go func() {
		defer requestWg.Done()

		vNodes, err := client.ListVirtualNodes(ctx, meshName)
		if err != nil {
			errorChan <- err
			return
		}
		vnNames = vNodes
	}()

	// Virtual Services
	requestWg.Add(1)
	go func() {
		defer requestWg.Done()

		vServices, err := client.ListVirtualServices(ctx, meshName)
		if err != nil {
			errorChan <- err
			return
		}
		vsNames = vServices
	}()

	// Virtual Routers
	requestWg.Add(1)
	go func() {
		defer requestWg.Done()

		vRouters, err := client.ListVirtualRouters(ctx, meshName)
		if err != nil {
			errorChan <- err
			return
		}

		// Get the routes for each virtual router
		var routeWg sync.WaitGroup
		for _, vr := range vRouters {
			routeWg.Add(1)

			go func(vrName string) {
				defer routeWg.Done()

				routes, err := client.ListRoutes(ctx, meshName, vrName)
				if err != nil {
					errorChan <- err
					return
				} else {
					vrMap.Store(vrName, routes)
				}
			}(vr)
		}
		routeWg.Wait()
	}()

	// Close in separate go routine so we can start consuming the errors, the channel might fill up and we wait forever
	go func() {
		requestWg.Wait()
		close(errorChan)
	}()

	// Collect errors
	if err := drainErrorChannel(errorChan); err != nil {
		return nil, errors.Wrapf(err, "failed to list all App Mesh resources for mesh %s", meshName)
	}

	// Transform VR map to regular map
	virtualRouters := make(map[string][]string)
	vrMap.Range(func(key, value interface{}) bool {
		vrName := key.(string)
		vrRoutes := value.([]string)
		virtualRouters[vrName] = vrRoutes

		return true
	})

	return &Resources{
		VirtualNodes:    vnNames,
		VirtualServices: vsNames,
		VirtualRouters:  virtualRouters,
	}, nil
}
