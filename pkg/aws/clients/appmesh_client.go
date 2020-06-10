package clients

import (
	"context"

	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
	"github.com/solo-io/go-utils/contextutils"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/aws/credentials"
	matcher2 "github.com/solo-io/service-mesh-hub/pkg/aws/matcher"
	"k8s.io/apimachinery/pkg/util/sets"
)

type AppmeshClientGetter func(mesh *smh_discovery.Mesh) (AppmeshClient, error)

func AppmeshClientGetterProvider(
	matcher matcher2.AppmeshMatcher,
	awsCredentialsGetter credentials.AwsCredentialsGetter,
	appmeshRawClientFactory AppmeshRawClientFactory,
) AppmeshClientGetter {
	return func(mesh *smh_discovery.Mesh) (AppmeshClient, error) {
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
	matcher matcher2.AppmeshMatcher
	client  appmeshiface.AppMeshAPI
}

func NewAppmeshClient(
	client appmeshiface.AppMeshAPI,
	matcher matcher2.AppmeshMatcher,
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
	existingVirtualRouterNames, err := a.listAllVirtualRouterNames(ctx, meshName)
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
	existingVirtualNodeNames, err := a.listVirtualNodeNames(ctx, meshName)
	if err != nil {
		return err
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
	existingVirtualServiceNames, err := a.listVirtualServiceNames(ctx, meshName)
	if err != nil {
		return err
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
		existingRouteNames, err := a.listRouteNames(ctx, meshName, aws2.String(vrName))
		if err != nil {
			return err
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

func (a *appmeshClient) listAllVirtualRouterNames(ctx context.Context, meshName *string) (sets.String, error) {
	vrNames := sets.NewString()
	req := &appmesh.ListVirtualRoutersInput{
		MeshName: meshName,
	}
	err := a.client.ListVirtualRoutersPagesWithContext(ctx, req, func(page *appmesh.ListVirtualRoutersOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}
		for _, vsRef := range page.VirtualRouters {
			vrNames.Insert(aws2.StringValue(vsRef.VirtualRouterName))
		}
		return !isLast
	})
	if err != nil {
		return nil, err
	}
	return vrNames, nil
}

func (a *appmeshClient) listRouteNames(
	ctx context.Context,
	meshName *string,
	virtualRouterName *string,
) (sets.String, error) {
	existingRouteNames := sets.NewString()
	req := &appmesh.ListRoutesInput{
		MeshName:          meshName,
		VirtualRouterName: virtualRouterName,
	}
	err := a.client.ListRoutesPagesWithContext(ctx, req, func(page *appmesh.ListRoutesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}
		for _, vsRef := range page.Routes {
			existingRouteNames.Insert(aws2.StringValue(vsRef.RouteName))
		}
		return !isLast
	})
	if err != nil {
		return nil, err
	}
	return existingRouteNames, nil
}

func (a *appmeshClient) listVirtualNodeNames(
	ctx context.Context,
	meshName *string,
) (sets.String, error) {
	virtualNodeNames := sets.NewString()
	req := &appmesh.ListVirtualNodesInput{
		MeshName: meshName,
	}
	err := a.client.ListVirtualNodesPagesWithContext(ctx, req, func(page *appmesh.ListVirtualNodesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}
		for _, vsRef := range page.VirtualNodes {
			virtualNodeNames.Insert(aws2.StringValue(vsRef.VirtualNodeName))
		}
		return !isLast
	})
	if err != nil {
		return nil, err
	}
	return virtualNodeNames, nil
}

func (a *appmeshClient) listVirtualServiceNames(
	ctx context.Context,
	meshName *string,
) (sets.String, error) {
	virtualServiceNames := sets.NewString()
	req := &appmesh.ListVirtualServicesInput{
		MeshName: meshName,
	}
	err := a.client.ListVirtualServicesPagesWithContext(ctx, req, func(page *appmesh.ListVirtualServicesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}
		for _, vsRef := range page.VirtualServices {
			virtualServiceNames.Insert(aws2.StringValue(vsRef.VirtualServiceName))
		}
		return !isLast
	})
	if err != nil {
		return nil, err
	}
	return virtualServiceNames, nil
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
