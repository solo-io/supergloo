package appmesh

import (
	"context"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/go-utils/errors"
)

type Resources struct {
	VirtualNodes, VirtualServices []string
	VirtualRouters                map[string][]string
}

func ListAllForMesh(ctx context.Context, client Client, meshName string) (*Resources, error) {

	type virtualRouter struct {
		name       string
		routeNames []string
	}

	var (
		requestWg           sync.WaitGroup
		virtualNodesChan    = make(chan []string, 10)
		virtualServicesChan = make(chan []string, 10)
		virtualRouterChan   = make(chan virtualRouter, 10)
		errorChan           = make(chan error, 10)
		collectWg           sync.WaitGroup
		vnNames,
		vsNames []string
		virtualRouters = make(map[string][]string)
		err            *multierror.Error
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
		virtualNodesChan <- vNodes
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
		virtualServicesChan <- vServices
	}()

	// Virtual Routers
	requestWg.Add(1)
	go func() {
		defer requestWg.Done()

		// Get all virtual routers
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
					virtualRouterChan <- virtualRouter{name: vrName, routeNames: routes}
				}
			}(vr)
		}
		routeWg.Wait()
	}()

	// Close channels when all workers have finished
	go func() {
		requestWg.Wait()
		close(virtualNodesChan)
		close(virtualServicesChan)
		close(virtualRouterChan)
		close(errorChan)
	}()

	// Collect the results
	// Each `range` statement will block until the correspondent channel is closed
	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for vn := range virtualNodesChan {
			vnNames = vn
		}
	}()

	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for vs := range virtualServicesChan {
			vsNames = vs
		}
	}()

	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for vr := range virtualRouterChan {
			virtualRouters[vr.name] = vr.routeNames
		}
	}()

	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for e := range errorChan {
			err = multierror.Append(err, e)
		}
	}()

	// Wait on resource collection to finish
	collectWg.Wait()

	if mergedError := err.ErrorOrNil(); mergedError != nil {
		return nil, errors.Wrapf(mergedError, "failed to list all App Mesh resources for mesh %s", meshName)
	}

	return &Resources{
		VirtualNodes:    vnNames,
		VirtualServices: vsNames,
		VirtualRouters:  virtualRouters,
	}, nil
}
