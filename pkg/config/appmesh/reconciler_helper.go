package appmesh

import (
	"context"
	"sync"

	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/go-utils/errors"

	translator "github.com/solo-io/supergloo/pkg/translator/appmesh"
)

func newHelper(ctx context.Context, client Client, snapshot *translator.ResourceSnapshot) *reconcileHelper {
	return &reconcileHelper{
		ctx:      ctx,
		client:   client,
		snapshot: snapshot,
	}
}

type reconcileHelper struct {
	ctx      context.Context
	client   Client
	snapshot *translator.ResourceSnapshot
}

func (h *reconcileHelper) createAll() error {
	if _, err := h.client.CreateMesh(h.ctx, h.snapshot.MeshName); err != nil {
		return err
	}

	// Step 1:
	// Create Virtual Nodes and Virtual Routers in parallel, as these do not depend on other resources.
	// Virtual nodes reference Virtual Services, but the referential integrity is not enforced by the API.
	errorChan1 := make(chan error, 10)
	wg1 := sync.WaitGroup{}
	wg1.Add(len(h.snapshot.VirtualNodes) + len(h.snapshot.VirtualRouters))
	for _, vn := range h.snapshot.VirtualNodes {
		go func(data appmesh.VirtualNodeData) {
			defer wg1.Done()
			if _, err := h.client.CreateVirtualNode(h.ctx, &data); err != nil {
				errorChan1 <- err
			}
		}(*vn)
	}
	for _, vr := range h.snapshot.VirtualRouters {
		go func(data appmesh.VirtualRouterData) {
			defer wg1.Done()
			if _, err := h.client.CreateVirtualRouter(h.ctx, &data); err != nil {
				errorChan1 <- err
			}
		}(*vr)
	}

	// Close in separate go routine so we can start consuming the errors, the channel might fill up and we wait forever
	go func() {
		wg1.Wait()
		close(errorChan1)
	}()

	// Blocks until channel is closed
	if err := drainErrorChannel(errorChan1); err != nil {
		return errors.Wrapf(err, "failed to create resources for mesh %s", h.snapshot.MeshName)
	}

	// Step 2:
	// Create Routes and Virtual Services in parallel
	errorChan2 := make(chan error)
	wg2 := sync.WaitGroup{}
	wg2.Add(len(h.snapshot.Routes) + len(h.snapshot.VirtualServices))
	for _, route := range h.snapshot.Routes {
		go func(data appmesh.RouteData) {
			defer wg2.Done()
			if _, err := h.client.CreateRoute(h.ctx, &data); err != nil {
				errorChan2 <- err
			}
		}(*route)
	}
	for _, vs := range h.snapshot.VirtualServices {
		go func(data appmesh.VirtualServiceData) {
			defer wg2.Done()
			if _, err := h.client.CreateVirtualService(h.ctx, &data); err != nil {
				errorChan2 <- err
			}
		}(*vs)
	}

	go func() {
		wg2.Wait()
		close(errorChan2)
	}()

	if err := drainErrorChannel(errorChan2); err != nil {
		return errors.Wrapf(err, "failed to create resources for mesh %s", h.snapshot.MeshName)
	}

	return nil
}

func (h *reconcileHelper) reconcile() error {

	// List the existing resources of each type
	allResources, err := ListAllForMesh(h.ctx, h.client, h.snapshot.MeshName)
	if err != nil {
		return err
	}

	markedResources, err := h.writeAndMarkResources(allResources)
	if err != nil {
		return err
	}

	// For each of the existing resources, if it has no correspondent resource in the snapshot, then delete it
	if err := h.deleteUnmarkedResources(markedResources, allResources); err != nil {
		return err
	}

	return nil
}

// For each of the resources in the snapshot:
//   - if it does not exist, create it
//   - if it does already exist, overwrite the existing one
// Return information about which existing resources are present in the snapshot
func (h *reconcileHelper) writeAndMarkResources(resources *Resources) (*markedResources, error) {
	snap := h.snapshot

	existing := buildMarkedResourcesMap(resources)

	// 1. Virtual Nodes & Virtual Routers
	errorChan1 := make(chan error, 50)
	wg1 := sync.WaitGroup{}
	wg1.Add(len(snap.VirtualNodes) + len(snap.VirtualRouters))
	for name, vn := range snap.VirtualNodes {
		go func(name string, data appmesh.VirtualNodeData) {
			defer wg1.Done()

			if _, exists := existing.VirtualNodes.Load(name); exists {
				if _, err := h.client.UpdateVirtualNode(h.ctx, &data); err != nil {
					errorChan1 <- err
				}

				// Mark existing resource
				existing.VirtualNodes.Store(name, true)
			} else {
				if _, err := h.client.CreateVirtualNode(h.ctx, &data); err != nil {
					errorChan1 <- err
				}
			}
		}(name, *vn)
	}
	for name, vr := range snap.VirtualRouters {
		go func(name string, data appmesh.VirtualRouterData) {
			defer wg1.Done()
			if _, exists := existing.VirtualRouters.Load(name); exists {
				if _, err := h.client.UpdateVirtualRouter(h.ctx, &data); err != nil {
					errorChan1 <- err
				}

				// Mark existing resource
				existing.VirtualRouters.Store(name, true)
			} else {
				if _, err := h.client.CreateVirtualRouter(h.ctx, &data); err != nil {
					errorChan1 <- err
				}
			}
		}(name, *vr)
	}
	// Close in separate go routine so we can start consuming the errors, the channel might fill up and we wait forever
	go func() {
		wg1.Wait()
		close(errorChan1)
	}()

	if err := drainErrorChannel(errorChan1); err != nil {
		return nil, errors.Wrapf(err, "failed to write resources for mesh %s", snap.MeshName)
	}

	// 2. Routes (reference a VR) & Virtual Services (can reference either a VN or a VR)
	errorChan2 := make(chan error)
	wg2 := sync.WaitGroup{}
	wg2.Add(len(snap.Routes) + len(snap.VirtualServices))
	for name, route := range snap.Routes {
		go func(name string, data appmesh.RouteData) {
			defer wg2.Done()
			if _, exists := existing.Routes.Load(name); exists {
				if _, err := h.client.UpdateRoute(h.ctx, &data); err != nil {
					errorChan2 <- err
				}

				// Mark existing resource
				existing.Routes.Store(name, true)
			} else {
				if _, err := h.client.CreateRoute(h.ctx, &data); err != nil {
					errorChan2 <- err
				}
			}
		}(name, *route)
	}
	for name, vs := range snap.VirtualServices {
		go func(name string, data appmesh.VirtualServiceData) {
			defer wg2.Done()

			if _, exists := existing.VirtualServices.Load(name); exists {
				if _, err := h.client.UpdateVirtualService(h.ctx, &data); err != nil {
					errorChan2 <- err
				}

				// Mark existing resource
				existing.VirtualServices.Store(name, true)
			} else {
				if _, err := h.client.CreateVirtualService(h.ctx, &data); err != nil {
					errorChan2 <- err
				}
			}
		}(name, *vs)
	}
	go func() {
		wg2.Wait()
		close(errorChan2)
	}()

	if err := drainErrorChannel(errorChan2); err != nil {
		return nil, errors.Wrapf(err, "failed to write resources for mesh %s", snap.MeshName)
	}

	return existing, nil
}

func (h *reconcileHelper) deleteUnmarkedResources(marked *markedResources, all *Resources) error {

	// 1. Virtual Services
	errorChan := make(chan error, 10)
	vsWg := sync.WaitGroup{}
	marked.VirtualServices.Range(func(key, value interface{}) bool {
		if name, required := key.(string), value.(bool); !required {
			vsWg.Add(1)
			go func() {
				defer vsWg.Done()
				if err := h.client.DeleteVirtualService(h.ctx, h.snapshot.MeshName, name); err != nil {
					errorChan <- err
				}
			}()
		}
		return true
	})

	go func() {
		vsWg.Wait()
		close(errorChan)
	}()

	if err := drainErrorChannel(errorChan); err != nil {
		return err
	}

	// 2. Virtual Routers + Routes
	errorChan = make(chan error, 10)
	vrWg := sync.WaitGroup{}
	marked.VirtualRouters.Range(func(key, value interface{}) bool {
		vrName, required := key.(string), value.(bool)

		vrWg.Add(1)
		go func(vrName string, required bool) {
			defer vrWg.Done()

			// Delete the routes linked to this router
			for _, routeName := range all.VirtualRouters[vrName] {
				// If it is marked as required, the route has been assigned to a different VR, so don't delete it
				if required, _ := marked.Routes.Load(routeName); !required.(bool) {
					vrWg.Add(1)
					go func(rName, vrName string) {
						defer vrWg.Done()
						if err := h.client.DeleteRoute(h.ctx, h.snapshot.MeshName, vrName, rName); err != nil {
							errorChan <- err
						}
					}(routeName, vrName)
				}
			}

			if !required {
				if err := h.client.DeleteVirtualRouter(h.ctx, h.snapshot.MeshName, vrName); err != nil {
					errorChan <- err
				}
			}

		}(vrName, required)

		return true
	})
	go func() {
		vrWg.Wait()
		close(errorChan)
	}()
	if err := drainErrorChannel(errorChan); err != nil {
		return err
	}

	// 3. Virtual Nodes
	errorChan = make(chan error, 10)
	vnWg := sync.WaitGroup{}
	marked.VirtualNodes.Range(func(key, value interface{}) bool {
		if name, required := key.(string), value.(bool); !required {
			vnWg.Add(1)
			go func() {
				defer vnWg.Done()
				if err := h.client.DeleteVirtualNode(h.ctx, h.snapshot.MeshName, name); err != nil {
					errorChan <- err
				}
			}()
		}
		return true
	})
	go func() {
		vnWg.Wait()
		close(errorChan)
	}()
	if err := drainErrorChannel(errorChan); err != nil {
		return err
	}

	return nil
}

// Maps the names of all the existing resources to a boolean flag, initially set to false.
// During reconciliation, the flags is set to true if a matching resource is present in the snapshot.
// At the end of the process, the resources with a false flag will be deleted.
type markedResources struct {
	VirtualNodes, VirtualServices, VirtualRouters, Routes sync.Map
}

func buildMarkedResourcesMap(r *Resources) *markedResources {
	result := markedResources{
		VirtualNodes:    sync.Map{},
		VirtualServices: sync.Map{},
		VirtualRouters:  sync.Map{},
		Routes:          sync.Map{},
	}

	for _, vn := range r.VirtualNodes {
		result.VirtualNodes.Store(vn, false)
	}

	for _, vs := range r.VirtualServices {
		result.VirtualServices.Store(vs, false)
	}

	for vr, routes := range r.VirtualRouters {
		result.VirtualRouters.Store(vr, false)
		for _, route := range routes {
			result.Routes.Store(route, false)
		}
	}
	return &result
}

// Blocks until channel is closed
func drainErrorChannel(errorChan chan error) error {
	// Collect errors
	var err *multierror.Error
	for e := range errorChan {
		err = multierror.Append(err, e)
	}
	return err.ErrorOrNil()
}
