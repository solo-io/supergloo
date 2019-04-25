package appmesh

import (
	"context"

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

	for _, vn := range h.snapshot.VirtualNodes {
		if _, err := h.client.CreateVirtualNode(h.ctx, vn); err != nil {
			return err
		}
	}

	for _, vr := range h.snapshot.VirtualRouters {
		if _, err := h.client.CreateVirtualRouter(h.ctx, vr); err != nil {
			return err
		}
	}

	for _, route := range h.snapshot.Routes {
		if _, err := h.client.CreateRoute(h.ctx, route); err != nil {
			return err
		}
	}

	for _, vs := range h.snapshot.VirtualServices {
		if _, err := h.client.CreateVirtualService(h.ctx, vs); err != nil {
			return err
		}
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

	existing := buildMarkedResourcesMap(resources)

	// 1. Virtual Nodes
	for name, vn := range h.snapshot.VirtualNodes {
		if _, exists := existing.VirtualNodes[name]; exists {
			if _, err := h.client.UpdateVirtualNode(h.ctx, vn); err != nil {
				return nil, err
			}

			// Mark existing resource
			existing.VirtualNodes[name] = true
		} else {
			if _, err := h.client.CreateVirtualNode(h.ctx, vn); err != nil {
				return nil, err
			}
		}
	}

	// 2. Virtual Router
	for name, vr := range h.snapshot.VirtualRouters {
		if _, exists := existing.VirtualRouters[name]; exists {
			if _, err := h.client.UpdateVirtualRouter(h.ctx, vr); err != nil {
				return nil, err
			}

			// Mark existing resource
			existing.VirtualRouters[name] = true
		} else {
			if _, err := h.client.CreateVirtualRouter(h.ctx, vr); err != nil {
				return nil, err
			}
		}
	}

	// 3. Routes (reference a VR)
	for name, route := range h.snapshot.Routes {
		if _, exists := existing.Routes[name]; exists {
			if _, err := h.client.UpdateRoute(h.ctx, route); err != nil {
				return nil, err
			}

			// Mark existing resource
			existing.Routes[name] = true
		} else {
			if _, err := h.client.CreateRoute(h.ctx, route); err != nil {
				return nil, err
			}
		}
	}

	// 4. Virtual Services (can reference either a VN or a VR)
	for name, vs := range h.snapshot.VirtualServices {
		if _, exists := existing.VirtualServices[name]; exists {
			if _, err := h.client.UpdateVirtualService(h.ctx, vs); err != nil {
				return nil, err
			}

			// Mark existing resource
			existing.VirtualServices[name] = true
		} else {
			if _, err := h.client.CreateVirtualService(h.ctx, vs); err != nil {
				return nil, err
			}
		}
	}

	return existing, nil
}

func (h *reconcileHelper) deleteUnmarkedResources(marked *markedResources, all *Resources) error {

	// 1. Virtual Services
	for vsName, required := range marked.VirtualServices {
		if !required {
			if err := h.client.DeleteVirtualService(h.ctx, h.snapshot.MeshName, vsName); err != nil {
				return err
			}
		}
	}

	// 2. Virtual Routers + Routes
	for vrName, required := range marked.VirtualRouters {

		// Delete the routes linked to this router
		for _, routeName := range all.VirtualRouters[vrName] {
			// If it is marked as required, the route has have been assigned to a different VR, so don't delete it
			if required := marked.Routes[routeName]; !required {
				if err := h.client.DeleteRoute(h.ctx, h.snapshot.MeshName, vrName, routeName); err != nil {
					return err
				}
			}
		}
		if !required {
			if err := h.client.DeleteVirtualRouter(h.ctx, h.snapshot.MeshName, vrName); err != nil {
				return err
			}
		}
	}

	// 3. Virtual Nodes
	for vnName, required := range marked.VirtualNodes {
		if !required {
			if err := h.client.DeleteVirtualRouter(h.ctx, h.snapshot.MeshName, vnName); err != nil {
				return err
			}
		}
	}

	return nil
}

// Maps the names of all the existing resources to a boolean flag, initially set to false.
// During reconciliation, the flags is set to true if a matching resource is present in the snapshot.
// At the end of the process, the resources with a false flag will be deleted.
type markedResources struct {
	VirtualNodes, VirtualServices, VirtualRouters, Routes map[string]bool
}

func buildMarkedResourcesMap(r *Resources) *markedResources {
	result := markedResources{
		VirtualNodes:    make(map[string]bool),
		VirtualServices: make(map[string]bool),
		VirtualRouters:  make(map[string]bool),
		Routes:          make(map[string]bool),
	}

	for _, vn := range r.VirtualNodes {
		result.VirtualNodes[vn] = false
	}

	for _, vs := range r.VirtualServices {
		result.VirtualServices[vs] = false
	}

	for vr, routes := range r.VirtualRouters {
		result.VirtualRouters[vr] = false
		for _, route := range routes {
			result.Routes[route] = false
		}
	}
	return &result
}
