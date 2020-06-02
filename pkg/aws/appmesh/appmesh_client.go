package appmesh

import (
	"context"

	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
	"github.com/solo-io/go-utils/contextutils"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/aws"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	ListLimit = aws2.Int64(100)
)

type AppmeshClientGetter func(mesh *zephyr_discovery.Mesh) (AppmeshClient, error)

func AppmeshClientGetterProvider(
	matcher AppmeshMatcher,
	awsCredentialsGetter aws.AwsCredentialsGetter,
	appmeshRawClientFactory AppmeshRawClientFactory,
) AppmeshClientGetter {
	return func(mesh *zephyr_discovery.Mesh) (AppmeshClient, error) {
		creds, err := awsCredentialsGetter.Get(mesh.Spec.GetAwsAppMesh().GetAwsAccountId())
		if err != nil {
			return nil, err
		}
		rawAppmeshClient, err := appmeshRawClientFactory(creds, mesh.Spec.GetAwsAppMesh().GetRegion())
		if err != nil {
			return nil, err
		}
		return NewAppmeshClient(rawAppmeshClient, matcher), nil
	}
}

type appmeshClient struct {
	matcher AppmeshMatcher
	client  appmeshiface.AppMeshAPI
}

func NewAppmeshClient(
	client appmeshiface.AppMeshAPI,
	matcher AppmeshMatcher,
) AppmeshClient {
	return &appmeshClient{
		client:  client,
		matcher: matcher,
	}
}

func (a *appmeshClient) EnsureVirtualService(virtualService *appmesh.VirtualServiceData) error {
	resp, err := a.client.DescribeVirtualService(&appmesh.DescribeVirtualServiceInput{
		MeshName:           virtualService.MeshName,
		VirtualServiceName: virtualService.VirtualServiceName,
	})
	if err != nil {
		if isNotFound(err) {
			_, err = a.client.CreateVirtualService(&appmesh.CreateVirtualServiceInput{
				MeshName:           virtualService.MeshName,
				VirtualServiceName: virtualService.VirtualServiceName,
				Spec:               virtualService.Spec,
			})
			return err
		} else {
			return err
		}
	}
	if !a.matcher.AreVirtualServicesEqual(resp.VirtualService, virtualService) {
		_, err := a.client.UpdateVirtualService(&appmesh.UpdateVirtualServiceInput{
			MeshName:           virtualService.MeshName,
			VirtualServiceName: virtualService.VirtualServiceName,
			Spec:               virtualService.Spec,
		})
		return err
	}
	return nil
}

func (a *appmeshClient) EnsureVirtualRouter(virtualRouter *appmesh.VirtualRouterData) error {
	resp, err := a.client.DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
		MeshName:          virtualRouter.MeshName,
		VirtualRouterName: virtualRouter.VirtualRouterName,
	})
	if err != nil {
		if isNotFound(err) {
			_, err = a.client.CreateVirtualRouter(&appmesh.CreateVirtualRouterInput{
				MeshName:          virtualRouter.MeshName,
				VirtualRouterName: virtualRouter.VirtualRouterName,
				Spec:              virtualRouter.Spec,
			})
			return err
		} else {
			return err
		}
	}
	if !a.matcher.AreVirtualRoutersEqual(resp.VirtualRouter, virtualRouter) {
		_, err := a.client.UpdateVirtualRouter(&appmesh.UpdateVirtualRouterInput{
			MeshName:          virtualRouter.MeshName,
			VirtualRouterName: virtualRouter.VirtualRouterName,
			Spec:              virtualRouter.Spec,
		})
		return err
	}
	return nil
}

func (a *appmeshClient) EnsureRoute(route *appmesh.RouteData) error {
	resp, err := a.client.DescribeRoute(&appmesh.DescribeRouteInput{
		MeshName:          route.MeshName,
		VirtualRouterName: route.VirtualRouterName,
		RouteName:         route.RouteName,
	})
	if err != nil {
		if isNotFound(err) {
			_, err = a.client.CreateRoute(&appmesh.CreateRouteInput{
				MeshName:          route.MeshName,
				RouteName:         route.RouteName,
				VirtualRouterName: route.VirtualRouterName,
				Spec:              route.Spec,
			})
			return err
		} else {
			return err
		}
	}
	if !a.matcher.AreRoutesEqual(resp.Route, route) {
		_, err := a.client.UpdateRoute(&appmesh.UpdateRouteInput{
			MeshName:          route.MeshName,
			RouteName:         route.RouteName,
			VirtualRouterName: route.VirtualRouterName,
			Spec:              route.Spec,
		})
		return err
	}
	return nil
}

func (a *appmeshClient) EnsureVirtualNode(virtualNode *appmesh.VirtualNodeData) error {
	resp, err := a.client.DescribeVirtualNode(&appmesh.DescribeVirtualNodeInput{
		MeshName:        virtualNode.MeshName,
		VirtualNodeName: virtualNode.VirtualNodeName,
	})
	if err != nil {
		if isNotFound(err) {
			_, err = a.client.CreateVirtualNode(&appmesh.CreateVirtualNodeInput{
				MeshName:        virtualNode.MeshName,
				VirtualNodeName: virtualNode.VirtualNodeName,
				Spec:            virtualNode.Spec,
			})
			return err
		} else {
			return err
		}
	}
	if !a.matcher.AreVirtualNodesEqual(resp.VirtualNode, virtualNode) {
		_, err := a.client.UpdateVirtualNode(&appmesh.UpdateVirtualNodeInput{
			MeshName:        virtualNode.MeshName,
			VirtualNodeName: virtualNode.VirtualNodeName,
			Spec:            virtualNode.Spec,
		})
		return err
	}
	return nil
}

/*
	AWS only allows deleting VirtualRouters if it has no associated Routes and is not a provider for a VirtualService.
	AWS cannot create a Route for a non-existing VirtualRouter.

	The order of operations is as follows:
	1. Create VirtualRouters
	2. Create Routes
	3. Create VirtualServices
	4. Delete VirtualServices
	5. Delete Routes
	6. Delete VirtualRouters
*/
func (a *appmeshClient) ReconcileVirtualRoutersAndRoutesAndVirtualServices(
	ctx context.Context,
	meshName *string,
	virtualRouters []*appmesh.VirtualRouterData,
	routes []*appmesh.RouteData,
	virtualServices []*appmesh.VirtualServiceData,
) error {
	err := a.ensureVirtualRouters(ctx, meshName, virtualRouters)
	if err != nil {
		return err
	}
	err = a.ensureRoutes(ctx, meshName, routes)
	if err != nil {
		return err
	}
	err = a.reconcileVirtualServices(ctx, meshName, virtualServices)
	if err != nil {
		return err
	}
	existingVirtualRouterNames, err := a.listAllVirtualRouterNames(meshName)
	if err != nil {
		return err
	}
	err = a.deleteRoutes(ctx, meshName, routes, existingVirtualRouterNames)
	if err != nil {
		return err
	}
	return a.deleteVirtualRouters(ctx, meshName, virtualRouters, existingVirtualRouterNames)
}

func (a *appmeshClient) ReconcileVirtualNodes(
	ctx context.Context,
	meshName *string,
	virtualNodes []*appmesh.VirtualNodeData,
) error {
	logger := contextutils.LoggerFrom(ctx)
	existingVirtualNodeNames := sets.NewString()
	declaredVirtualNodeNames := sets.NewString()
	// For each declared VirtualNode, ensure it exists with an equivalent spec.
	for _, vn := range virtualNodes {
		if aws2.StringValue(vn.MeshName) != aws2.StringValue(meshName) {
			logger.Warnf("Skipping VirtualNode (Name: %s, MeshName: %s) that doesn't belong under the provided Mesh (%s).",
				aws2.StringValue(vn.VirtualNodeName),
				aws2.StringValue(vn.MeshName),
				aws2.StringValue(meshName),
			)
			continue
		}
		err := a.EnsureVirtualNode(vn)
		if err != nil {
			logger.Errorf("Error ensuring VirtualNode. %+v", err)
		}
		declaredVirtualNodeNames.Insert(aws2.StringValue(vn.VirtualNodeName))
	}
	req := &appmesh.ListVirtualNodesInput{
		Limit:    ListLimit,
		MeshName: meshName,
	}
	for {
		resp, err := a.client.ListVirtualNodes(req)
		if err != nil {
			return err
		}
		for _, vsRef := range resp.VirtualNodes {
			existingVirtualNodeNames.Insert(aws2.StringValue(vsRef.VirtualNodeName))
		}
		if resp.NextToken == nil {
			break
		}
		req.NextToken = resp.NextToken
	}
	// Delete any VirtualNodes not declared
	for vnName, _ := range existingVirtualNodeNames.Difference(declaredVirtualNodeNames) {
		_, err := a.client.DeleteVirtualNode(&appmesh.DeleteVirtualNodeInput{
			MeshName:        meshName,
			VirtualNodeName: aws2.String(vnName),
		})
		if err != nil {
			logger.Errorf("Error deleting VirtualNode. %+v", err)
		}
	}
	return nil
}

func (a *appmeshClient) reconcileVirtualServices(
	ctx context.Context,
	meshName *string,
	virtualServices []*appmesh.VirtualServiceData,
) error {
	logger := contextutils.LoggerFrom(ctx)
	existingVirtualServiceNames := sets.NewString()
	declaredVirtualServiceNames := sets.NewString()
	// For each declared VirtualService, ensure it exists with an equivalent spec.
	for _, vs := range virtualServices {
		if aws2.StringValue(vs.MeshName) != aws2.StringValue(meshName) {
			logger.Warnf("Skipping VirtualService (Name: %s, MeshName: %s) that doesn't belong under the provided Mesh (%s).",
				aws2.StringValue(vs.VirtualServiceName),
				aws2.StringValue(vs.MeshName),
				aws2.StringValue(meshName),
			)
			continue
		}
		err := a.EnsureVirtualService(vs)
		if err != nil {
			logger.Errorf("Error ensuring VirtualService. %+v", err)
			return err
		}
		declaredVirtualServiceNames.Insert(aws2.StringValue(vs.VirtualServiceName))
	}
	req := &appmesh.ListVirtualServicesInput{
		Limit:    ListLimit,
		MeshName: meshName,
	}
	for {
		resp, err := a.client.ListVirtualServices(req)
		if err != nil {
			return err
		}
		for _, vsRef := range resp.VirtualServices {
			existingVirtualServiceNames.Insert(aws2.StringValue(vsRef.VirtualServiceName))
		}
		if resp.NextToken == nil {
			break
		}
		req.NextToken = resp.NextToken
	}
	// Delete any VirtualServices not declared
	for vsName, _ := range existingVirtualServiceNames.Difference(declaredVirtualServiceNames) {
		_, err := a.client.DeleteVirtualService(&appmesh.DeleteVirtualServiceInput{
			MeshName:           meshName,
			VirtualServiceName: aws2.String(vsName),
		})
		if err != nil {
			logger.Errorf("Error deleting VirtualService. %+v", err)
			return err
		}
	}
	return nil
}

func (a *appmeshClient) ensureRoutes(
	ctx context.Context,
	meshName *string,
	routes []*appmesh.RouteData,
) error {
	logger := contextutils.LoggerFrom(ctx)
	declaredVirtualRouterToRoutes := map[string]sets.String{}
	// For each declared Route, ensure it exists with an equivalent spec.
	for _, route := range routes {
		if aws2.StringValue(route.MeshName) != aws2.StringValue(meshName) {
			continue
		}
		err := a.EnsureRoute(route)
		if err != nil {
			logger.Errorf("Error ensuring Route. %+v", err)
			return err
		}
		routes, ok := declaredVirtualRouterToRoutes[aws2.StringValue(route.VirtualRouterName)]
		if !ok {
			declaredVirtualRouterToRoutes[aws2.StringValue(route.VirtualRouterName)] = sets.NewString()
			routes = declaredVirtualRouterToRoutes[aws2.StringValue(route.VirtualRouterName)]
		}
		routes.Insert(aws2.StringValue(route.RouteName))
	}
	return nil
}

func (a *appmeshClient) deleteRoutes(
	ctx context.Context,
	meshName *string,
	declaredRoutes []*appmesh.RouteData,
	existingVirtualRouterNames sets.String,
) error {
	logger := contextutils.LoggerFrom(ctx)
	declaredVirtualRouterToRoutes := map[string]sets.String{}
	for _, route := range declaredRoutes {
		if aws2.StringValue(route.MeshName) != aws2.StringValue(meshName) {
			continue
		}
		routes, ok := declaredVirtualRouterToRoutes[aws2.StringValue(route.VirtualRouterName)]
		if !ok {
			declaredVirtualRouterToRoutes[aws2.StringValue(route.VirtualRouterName)] = sets.NewString()
			routes = declaredVirtualRouterToRoutes[aws2.StringValue(route.VirtualRouterName)]
		}
		routes.Insert(aws2.StringValue(route.RouteName))
	}
	// Delete any Routes not declared
	for _, vrName := range existingVirtualRouterNames.List() {
		declaredRouteNames, ok := declaredVirtualRouterToRoutes[vrName]
		if !ok {
			// If no routes declared for VirtualRouter, delete all of its Routes.
			declaredRouteNames = sets.NewString()
		}
		existingRouteNames := sets.NewString()
		req := &appmesh.ListRoutesInput{
			Limit:             ListLimit,
			MeshName:          meshName,
			VirtualRouterName: aws2.String(vrName),
		}
		for {
			resp, err := a.client.ListRoutes(req)
			if err != nil {
				return err
			}
			for _, vsRef := range resp.Routes {
				existingRouteNames.Insert(aws2.StringValue(vsRef.RouteName))
			}
			if resp.NextToken == nil {
				break
			}
			req.NextToken = resp.NextToken
		}
		for routeName, _ := range existingRouteNames.Difference(declaredRouteNames) {
			_, err := a.client.DeleteRoute(&appmesh.DeleteRouteInput{
				MeshName:          meshName,
				VirtualRouterName: aws2.String(vrName),
				RouteName:         aws2.String(routeName),
			})
			if err != nil {
				logger.Errorf("Error deleting Route. %+v", err)
			}
		}
	}
	return nil
}

func (a *appmeshClient) ensureVirtualRouters(
	ctx context.Context,
	meshName *string,
	virtualRouters []*appmesh.VirtualRouterData,
) error {
	logger := contextutils.LoggerFrom(ctx)
	declaredVirtualRouterNames := sets.NewString()
	// For each declared VirtualRouter, ensure it exists with an equivalent spec.
	for _, vr := range virtualRouters {
		if aws2.StringValue(vr.MeshName) != aws2.StringValue(meshName) {
			logger.Warnf("Skipping VirtualRouter (Name: %s, MeshName: %s) that doesn't belong under the provided Mesh (%s).",
				aws2.StringValue(vr.VirtualRouterName),
				aws2.StringValue(vr.MeshName),
				aws2.StringValue(meshName),
			)
			continue
		}
		err := a.EnsureVirtualRouter(vr)
		if err != nil {
			logger.Errorf("Error ensuring VirtualRouter. %+v", err)
			return err
		}
		declaredVirtualRouterNames.Insert(aws2.StringValue(vr.VirtualRouterName))
	}
	return nil
}

func (a *appmeshClient) deleteVirtualRouters(
	ctx context.Context,
	meshName *string,
	declaredVirtualRouters []*appmesh.VirtualRouterData,
	existingVirtualRouterNames sets.String,
) error {
	logger := contextutils.LoggerFrom(ctx)
	declaredVirtualRouterNames := sets.NewString()
	for _, vr := range declaredVirtualRouters {
		if aws2.StringValue(vr.MeshName) != aws2.StringValue(meshName) {
			continue
		}
		declaredVirtualRouterNames.Insert(aws2.StringValue(vr.VirtualRouterName))
	}
	// Delete any VirtualRouters not declared
	for vrName, _ := range existingVirtualRouterNames.Difference(declaredVirtualRouterNames) {
		_, err := a.client.DeleteVirtualRouter(&appmesh.DeleteVirtualRouterInput{
			MeshName:          meshName,
			VirtualRouterName: aws2.String(vrName),
		})
		if err != nil {
			logger.Errorf("Error deleting VirtualRouter. %+v", err)
		}
	}
	return nil
}

func (a *appmeshClient) listAllVirtualRouterNames(meshName *string) (sets.String, error) {
	vrNames := sets.NewString()
	req := &appmesh.ListVirtualRoutersInput{
		Limit:    ListLimit,
		MeshName: meshName,
	}
	for {
		resp, err := a.client.ListVirtualRouters(req)
		if err != nil {
			return nil, err
		}
		for _, vsRef := range resp.VirtualRouters {
			vrNames.Insert(aws2.StringValue(vsRef.VirtualRouterName))
		}
		if resp.NextToken == nil {
			break
		}
		req.NextToken = resp.NextToken
	}
	return vrNames, nil
}

func isNotFound(err error) bool {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case appmesh.ErrCodeNotFoundException:
			return true
		default:
			return false
		}
	}
	return false
}
