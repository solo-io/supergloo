package appmesh

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/pkg/errors"
)

//go:generate mockgen -destination=./client_mock.go -source client.go -package appmesh

// Represents the App Mesh API
type Client interface {
	// Get operations
	GetMesh(ctx context.Context, meshName string) (*appmesh.MeshData, error)

	// List operations
	ListMeshes(ctx context.Context) ([]string, error)
	ListVirtualNodes(ctx context.Context, meshName string) ([]string, error)
	ListVirtualServices(ctx context.Context, meshName string) ([]string, error)
	ListVirtualRouters(ctx context.Context, meshName string) ([]string, error)
	ListRoutes(ctx context.Context, meshName, virtualRouterName string) ([]string, error)

	// Create operations
	CreateMesh(ctx context.Context, meshName string) (*appmesh.MeshData, error)
	CreateVirtualNode(ctx context.Context, virtualNode appmesh.VirtualNodeData) (*appmesh.VirtualNodeData, error)
	CreateVirtualService(ctx context.Context, virtualService appmesh.VirtualServiceData) (*appmesh.VirtualServiceData, error)
	CreateVirtualRouter(ctx context.Context, virtualRouter appmesh.VirtualRouterData) (*appmesh.VirtualRouterData, error)
	CreateRoute(ctx context.Context, route appmesh.RouteData) (*appmesh.RouteData, error)

	// Update operations
	UpdateVirtualNode(ctx context.Context, virtualNode appmesh.VirtualNodeData) (*appmesh.VirtualNodeData, error)
	UpdateVirtualService(ctx context.Context, virtualService appmesh.VirtualServiceData) (*appmesh.VirtualServiceData, error)
	UpdateVirtualRouter(ctx context.Context, virtualRouter appmesh.VirtualRouterData) (*appmesh.VirtualRouterData, error)
	UpdateRoute(ctx context.Context, route appmesh.RouteData) (*appmesh.RouteData, error)

	// Delete operations
	DeleteMesh(ctx context.Context, meshName string) error
	DeleteVirtualNode(ctx context.Context, meshName, virtualNodeName string) error
	DeleteVirtualService(ctx context.Context, meshName, virtualServiceName string) error
	DeleteVirtualRouter(ctx context.Context, meshName, virtualRouterName string) error
	DeleteRoute(ctx context.Context, meshName, virtualRouterName, routeName string) error
}

type client struct {
	api *appmesh.AppMesh
}

func (c *client) GetMesh(ctx context.Context, meshName string) (*appmesh.MeshData, error) {
	input := &appmesh.DescribeMeshInput{
		MeshName: aws.String(meshName),
	}
	if output, err := c.api.DescribeMeshWithContext(ctx, input); err != nil {
		if IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrapf(err, "failed to describe mesh %s", meshName)
	} else if output == nil {
		return nil, nil
	} else {
		return output.Mesh, nil
	}
}

func (c *client) ListMeshes(ctx context.Context) ([]string, error) {
	var meshRefs []*appmesh.MeshRef

	paginationFunc := func(output *appmesh.ListMeshesOutput, b bool) bool {
		meshRefs = append(meshRefs, output.Meshes...)
		return true
	}

	if err := c.api.ListMeshesPagesWithContext(ctx, &appmesh.ListMeshesInput{}, paginationFunc); err != nil {
		return nil, errors.Wrapf(err, "failed to list meshes")
	}

	var result []string
	for _, meshRef := range meshRefs {
		result = append(result, *meshRef.MeshName)
	}

	return result, nil
}

func (c *client) ListVirtualNodes(ctx context.Context, meshName string) ([]string, error) {
	var vnRefs []*appmesh.VirtualNodeRef

	input := &appmesh.ListVirtualNodesInput{
		MeshName: aws.String(meshName),
	}

	paginationFunc := func(output *appmesh.ListVirtualNodesOutput, b bool) bool {
		vnRefs = append(vnRefs, output.VirtualNodes...)
		return true
	}

	if err := c.api.ListVirtualNodesPagesWithContext(ctx, input, paginationFunc); err != nil {
		return nil, errors.Wrapf(err, "failed to list virtual nodes for mesh %s", meshName)
	}

	var result []string
	for _, vnRef := range vnRefs {
		result = append(result, *vnRef.VirtualNodeName)
	}

	return result, nil
}

func (c *client) ListVirtualServices(ctx context.Context, meshName string) ([]string, error) {
	var vsRefs []*appmesh.VirtualServiceRef

	input := &appmesh.ListVirtualServicesInput{
		MeshName: aws.String(meshName),
	}

	paginationFunc := func(output *appmesh.ListVirtualServicesOutput, b bool) bool {
		vsRefs = append(vsRefs, output.VirtualServices...)
		return true
	}

	if err := c.api.ListVirtualServicesPagesWithContext(ctx, input, paginationFunc); err != nil {
		return nil, errors.Wrapf(err, "failed to list virtual services for mesh %s", meshName)
	}

	var result []string
	for _, vsRef := range vsRefs {
		result = append(result, *vsRef.VirtualServiceName)
	}

	return result, nil
}

func (c *client) ListVirtualRouters(ctx context.Context, meshName string) ([]string, error) {
	var vrRefs []*appmesh.VirtualRouterRef

	input := &appmesh.ListVirtualRoutersInput{
		MeshName: aws.String(meshName),
	}

	paginationFunc := func(output *appmesh.ListVirtualRoutersOutput, b bool) bool {
		vrRefs = append(vrRefs, output.VirtualRouters...)
		return true
	}

	if err := c.api.ListVirtualRoutersPagesWithContext(ctx, input, paginationFunc); err != nil {
		return nil, errors.Wrapf(err, "failed to list virtual routers for mesh %s", meshName)
	}

	var result []string
	for _, vrRef := range vrRefs {
		result = append(result, *vrRef.VirtualRouterName)
	}

	return result, nil
}

func (c *client) ListRoutes(ctx context.Context, meshName, virtualRouterName string) ([]string, error) {
	var routeRefs []*appmesh.RouteRef

	input := &appmesh.ListRoutesInput{
		MeshName:          aws.String(meshName),
		VirtualRouterName: aws.String(virtualRouterName),
	}

	paginationFunc := func(output *appmesh.ListRoutesOutput, b bool) bool {
		routeRefs = append(routeRefs, output.Routes...)
		return true
	}

	if err := c.api.ListRoutesPagesWithContext(ctx, input, paginationFunc); err != nil {
		return nil, errors.Wrapf(err, "failed to list routes for mesh %s and virtual router %s", meshName, virtualRouterName)
	}

	var result []string
	for _, routeRef := range routeRefs {
		result = append(result, *routeRef.RouteName)
	}

	return result, nil
}

func (c *client) CreateMesh(ctx context.Context, meshName string) (*appmesh.MeshData, error) {
	input := &appmesh.CreateMeshInput{
		MeshName: aws.String(meshName),
	}

	if output, err := c.api.CreateMeshWithContext(ctx, input); err != nil {
		return nil, errors.Wrapf(err, "failed to create mesh %s", meshName)
	} else if output == nil || output.Mesh == nil {
		return nil, fmt.Errorf("unexpected empty output while creating mesh %s", meshName)
	} else {
		return output.Mesh, nil
	}
}

func (c *client) CreateVirtualNode(ctx context.Context, vn appmesh.VirtualNodeData) (*appmesh.VirtualNodeData, error) {
	input := &appmesh.CreateVirtualNodeInput{
		MeshName:        vn.MeshName,
		VirtualNodeName: vn.VirtualNodeName,
		Spec:            vn.Spec,
	}

	if output, err := c.api.CreateVirtualNodeWithContext(ctx, input); err != nil {
		return nil, errors.Wrapf(err, "failed to create virtual node %s in mesh %s", *vn.VirtualNodeName, *vn.MeshName)
	} else if output == nil || output.VirtualNode == nil {
		return nil, fmt.Errorf("unexpected empty output while creating virtual node %s in mesh %s", *vn.VirtualNodeName, *vn.MeshName)
	} else {
		return output.VirtualNode, nil
	}
}

func (c *client) CreateVirtualService(ctx context.Context, vs appmesh.VirtualServiceData) (*appmesh.VirtualServiceData, error) {
	input := &appmesh.CreateVirtualServiceInput{
		MeshName:           vs.MeshName,
		VirtualServiceName: vs.VirtualServiceName,
		Spec:               vs.Spec,
	}

	if output, err := c.api.CreateVirtualServiceWithContext(ctx, input); err != nil {
		return nil, errors.Wrapf(err, "failed to create virtual service %s in mesh %s", *vs.VirtualServiceName, *vs.MeshName)
	} else if output == nil || output.VirtualService == nil {
		return nil, fmt.Errorf("unexpected empty output while creating virtual service %s in mesh %s", *vs.VirtualServiceName, *vs.MeshName)
	} else {
		return output.VirtualService, nil
	}
}

func (c *client) CreateVirtualRouter(ctx context.Context, vr appmesh.VirtualRouterData) (*appmesh.VirtualRouterData, error) {
	input := &appmesh.CreateVirtualRouterInput{
		MeshName:          vr.MeshName,
		VirtualRouterName: vr.VirtualRouterName,
		Spec:              vr.Spec,
	}

	if output, err := c.api.CreateVirtualRouterWithContext(ctx, input); err != nil {
		return nil, errors.Wrapf(err, "failed to create virtual router %s in mesh %s", *vr.VirtualRouterName, *vr.MeshName)
	} else if output == nil || output.VirtualRouter == nil {
		return nil, fmt.Errorf("unexpected empty output while creating virtual router %s in mesh %s", *vr.VirtualRouterName, *vr.MeshName)
	} else {
		return output.VirtualRouter, nil
	}
}

func (c *client) CreateRoute(ctx context.Context, route appmesh.RouteData) (*appmesh.RouteData, error) {
	input := &appmesh.CreateRouteInput{
		MeshName:          route.MeshName,
		RouteName:         route.RouteName,
		VirtualRouterName: route.VirtualRouterName,
		Spec:              route.Spec,
	}

	if output, err := c.api.CreateRouteWithContext(ctx, input); err != nil {
		return nil, errors.Wrapf(err, "failed to create route %s for virtual router %s in mesh %s",
			*route.RouteName, *route.VirtualRouterName, *route.MeshName)
	} else if output == nil || output.Route == nil {
		return nil, fmt.Errorf("unexpected empty output while creating route %s for virtual router %s in mesh %s",
			*route.RouteName, *route.VirtualRouterName, *route.MeshName)
	} else {
		return output.Route, nil
	}
}

func (c *client) UpdateVirtualNode(ctx context.Context, vn appmesh.VirtualNodeData) (*appmesh.VirtualNodeData, error) {
	input := &appmesh.UpdateVirtualNodeInput{
		MeshName:        vn.MeshName,
		VirtualNodeName: vn.VirtualNodeName,
		Spec:            vn.Spec,
	}

	if output, err := c.api.UpdateVirtualNodeWithContext(ctx, input); err != nil {
		return nil, errors.Wrapf(err, "failed to update virtual node %s in mesh %s", *vn.VirtualNodeName, *vn.MeshName)
	} else if output == nil || output.VirtualNode == nil {
		return nil, fmt.Errorf("unexpected empty output while updating virtual node %s in mesh %s", *vn.VirtualNodeName, *vn.MeshName)
	} else {
		return output.VirtualNode, nil
	}
}

func (c *client) UpdateVirtualService(ctx context.Context, vs appmesh.VirtualServiceData) (*appmesh.VirtualServiceData, error) {
	input := &appmesh.UpdateVirtualServiceInput{
		MeshName:           vs.MeshName,
		VirtualServiceName: vs.VirtualServiceName,
		Spec:               vs.Spec,
	}

	if output, err := c.api.UpdateVirtualServiceWithContext(ctx, input); err != nil {
		return nil, errors.Wrapf(err, "failed to update virtual service %s in mesh %s", *vs.VirtualServiceName, *vs.MeshName)
	} else if output == nil || output.VirtualService == nil {
		return nil, fmt.Errorf("unexpected empty output while updating virtual service %s", *vs.VirtualServiceName)
	} else {
		return output.VirtualService, nil
	}
}

func (c *client) UpdateVirtualRouter(ctx context.Context, vr appmesh.VirtualRouterData) (*appmesh.VirtualRouterData, error) {
	input := &appmesh.UpdateVirtualRouterInput{
		MeshName:          vr.MeshName,
		VirtualRouterName: vr.VirtualRouterName,
		Spec:              vr.Spec,
	}

	if output, err := c.api.UpdateVirtualRouterWithContext(ctx, input); err != nil {
		return nil, errors.Wrapf(err, "failed to update virtual router %s in mesh %s", *vr.VirtualRouterName, *vr.MeshName)
	} else if output == nil || output.VirtualRouter == nil {
		return nil, fmt.Errorf("unexpected empty output while updating virtual router %s in mesh %s", *vr.VirtualRouterName, *vr.MeshName)
	} else {
		return output.VirtualRouter, nil
	}
}

func (c *client) UpdateRoute(ctx context.Context, route appmesh.RouteData) (*appmesh.RouteData, error) {
	input := &appmesh.UpdateRouteInput{
		MeshName:          route.MeshName,
		RouteName:         route.RouteName,
		VirtualRouterName: route.VirtualRouterName,
		Spec:              route.Spec,
	}

	if output, err := c.api.UpdateRouteWithContext(ctx, input); err != nil {
		return nil, errors.Wrapf(err, "failed to update route %s for virtual router %s in mesh %s",
			*route.RouteName, *route.VirtualRouterName, *route.MeshName)
	} else if output == nil || output.Route == nil {
		return nil, fmt.Errorf("unexpected empty output while updating route %s for virtual router %s in mesh %s",
			*route.RouteName, *route.VirtualRouterName, *route.MeshName)
	} else {
		return output.Route, nil
	}
}

func (c *client) DeleteMesh(ctx context.Context, meshName string) error {
	input := &appmesh.DeleteMeshInput{
		MeshName: aws.String(meshName),
	}
	if _, err := c.api.DeleteMeshWithContext(ctx, input); err != nil {
		return errors.Wrapf(err, "failed to delete mesh %s", meshName)
	}
	return nil
}

func (c *client) DeleteVirtualNode(ctx context.Context, meshName, virtualNodeName string) error {
	input := &appmesh.DeleteVirtualNodeInput{
		MeshName:        aws.String(meshName),
		VirtualNodeName: aws.String(virtualNodeName),
	}
	if _, err := c.api.DeleteVirtualNodeWithContext(ctx, input); err != nil {
		return errors.Wrapf(err, "failed to delete virtual node %s in mesh %s", virtualNodeName, meshName)
	}
	return nil
}

func (c *client) DeleteVirtualService(ctx context.Context, meshName, virtualServiceName string) error {
	input := &appmesh.DeleteVirtualServiceInput{
		MeshName:           aws.String(meshName),
		VirtualServiceName: aws.String(virtualServiceName),
	}
	if _, err := c.api.DeleteVirtualServiceWithContext(ctx, input); err != nil {
		return errors.Wrapf(err, "failed to delete virtual service %s in mesh %s", virtualServiceName, meshName)
	}
	return nil
}

func (c *client) DeleteVirtualRouter(ctx context.Context, meshName, virtualRouterName string) error {
	input := &appmesh.DeleteVirtualRouterInput{
		MeshName:          aws.String(meshName),
		VirtualRouterName: aws.String(virtualRouterName),
	}
	if _, err := c.api.DeleteVirtualRouterWithContext(ctx, input); err != nil {
		return errors.Wrapf(err, "failed to delete virtual router %s in mesh %s", virtualRouterName, meshName)
	}
	return nil
}

func (c *client) DeleteRoute(ctx context.Context, meshName, virtualRouterName, routeName string) error {
	input := &appmesh.DeleteRouteInput{
		MeshName:          aws.String(meshName),
		VirtualRouterName: aws.String(virtualRouterName),
		RouteName:         aws.String(routeName),
	}
	if _, err := c.api.DeleteRouteWithContext(ctx, input); err != nil {
		return errors.Wrapf(err, "failed to delete route %s for virtual router %s in mesh %s", routeName, virtualRouterName, meshName)
	}
	return nil
}

func IsNotFound(err error) bool {
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == appmesh.ErrCodeNotFoundException {
				return true
			}
		}
	}
	return false
}
