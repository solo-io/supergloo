package appmesh

import (
	aws2 "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/aws/aws-sdk-go/service/appmesh/appmeshiface"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/aws"
)

type AppmeshClientFactory func(mesh *zephyr_discovery.Mesh) (AppmeshClient, error)

func AppmeshClientFactoryProvider(
	matcher AppmeshMatcher,
	awsCredentialsGetter aws.AwsCredentialsGetter,
	appmeshRawClientFactory AppmeshRawClientFactory,
) AppmeshClientFactory {
	return func(mesh *zephyr_discovery.Mesh) (AppmeshClient, error) {
		creds, err := awsCredentialsGetter.Get(mesh.Spec.GetAwsAppMesh().GetAwsAccountId())
		if err != nil {
			return nil, err
		}
		rawAppmeshClient, err := appmeshRawClientFactory(creds, mesh.Spec.GetAwsAppMesh().GetRegion())
		if err != nil {
			return nil, err
		}
		return NewAppmeshClient(rawAppmeshClient, matcher, awsCredentialsGetter, appmeshRawClientFactory), nil
	}
}

type appmeshClient struct {
	matcher                 AppmeshMatcher
	awsCredentialsGetter    aws.AwsCredentialsGetter
	appmeshRawClientFactory AppmeshRawClientFactory
	client                  appmeshiface.AppMeshAPI
}

func NewAppmeshClient(
	client appmeshiface.AppMeshAPI,
	matcher AppmeshMatcher,
	awsCredentialsGetter aws.AwsCredentialsGetter,
	appmeshRawClientFactory AppmeshRawClientFactory,
) AppmeshClient {
	return &appmeshClient{
		client:                  client,
		matcher:                 matcher,
		awsCredentialsGetter:    awsCredentialsGetter,
		appmeshRawClientFactory: appmeshRawClientFactory,
	}
}

func (a *appmeshClient) EnsureVirtualService(virtualService *appmesh.VirtualServiceData) error {
	resp, err := a.client.DescribeVirtualService(&appmesh.DescribeVirtualServiceInput{
		MeshName:           virtualService.MeshName,
		VirtualServiceName: virtualService.VirtualServiceName,
	})
	if err != nil {
		if !isNotFound(err) {
			return err
		} else if !a.matcher.AreVirtualServicesEqual(resp.VirtualService, virtualService) {
			_, err := a.client.UpdateVirtualService(&appmesh.UpdateVirtualServiceInput{
				MeshName:           virtualService.MeshName,
				VirtualServiceName: virtualService.VirtualServiceName,
				Spec:               virtualService.Spec,
			})
			return err
		} else {
			return nil
		}
	}
	_, err = a.client.CreateVirtualService(&appmesh.CreateVirtualServiceInput{
		MeshName:           virtualService.MeshName,
		VirtualServiceName: virtualService.VirtualServiceName,
		Spec:               virtualService.Spec,
	})
	return err
}

func (a *appmeshClient) EnsureVirtualRouter(virtualRouter *appmesh.VirtualRouterData) error {
	resp, err := a.client.DescribeVirtualRouter(&appmesh.DescribeVirtualRouterInput{
		MeshName:          virtualRouter.MeshName,
		VirtualRouterName: virtualRouter.VirtualRouterName,
	})
	if err != nil {
		if !isNotFound(err) {
			return err
		} else if !a.matcher.AreVirtualRoutersEqual(resp.VirtualRouter, virtualRouter) {
			_, err := a.client.UpdateVirtualRouter(&appmesh.UpdateVirtualRouterInput{
				MeshName:          virtualRouter.MeshName,
				VirtualRouterName: virtualRouter.VirtualRouterName,
				Spec:              virtualRouter.Spec,
			})
			return err
		} else {
			return nil
		}
	}
	_, err = a.client.CreateVirtualRouter(&appmesh.CreateVirtualRouterInput{
		MeshName:          virtualRouter.MeshName,
		VirtualRouterName: virtualRouter.VirtualRouterName,
		Spec:              virtualRouter.Spec,
	})
	return err
}

func (a *appmeshClient) EnsureRoute(route *appmesh.RouteData) error {
	resp, err := a.client.DescribeRoute(&appmesh.DescribeRouteInput{
		MeshName:          route.MeshName,
		VirtualRouterName: route.VirtualRouterName,
		RouteName:         route.RouteName,
	})
	if err != nil {
		if !isNotFound(err) {
			return err
		} else if !a.matcher.AreRoutesEqual(resp.Route, route) {
			_, err := a.client.UpdateRoute(&appmesh.UpdateRouteInput{
				MeshName:          route.MeshName,
				RouteName:         route.RouteName,
				VirtualRouterName: route.VirtualRouterName,
				Spec:              route.Spec,
			})
			return err
		} else {
			return nil
		}
	}
	_, err = a.client.CreateRoute(&appmesh.CreateRouteInput{
		MeshName:          route.MeshName,
		RouteName:         route.RouteName,
		VirtualRouterName: route.VirtualRouterName,
		Spec:              route.Spec,
	})
	return err
}

func (a *appmeshClient) EnsureVirtualNode(virtualNode *appmesh.VirtualNodeData) error {
	resp, err := a.client.DescribeVirtualNode(&appmesh.DescribeVirtualNodeInput{
		MeshName:        virtualNode.MeshName,
		VirtualNodeName: virtualNode.VirtualNodeName,
	})
	if err != nil {
		if !isNotFound(err) {
			return err
		} else if !a.matcher.AreVirtualNodesEqual(resp.VirtualNode, virtualNode) {
			_, err := a.client.UpdateVirtualNode(&appmesh.UpdateVirtualNodeInput{
				MeshName:        virtualNode.MeshName,
				VirtualNodeName: virtualNode.VirtualNodeName,
				Spec:            virtualNode.Spec,
			})
			return err
		} else {
			return nil
		}
	}
	_, err = a.client.CreateVirtualNode(&appmesh.CreateVirtualNodeInput{
		MeshName:        virtualNode.MeshName,
		VirtualNodeName: virtualNode.VirtualNodeName,
		Spec:            virtualNode.Spec,
	})
	return err
}

func (a *appmeshClient) DeleteAllDefaultRoutes(meshName string) error {
	meshNamePtr := aws2.String(meshName)
	var virtualRouterNames []*string
	listVirtualRoutersInput := &appmesh.ListVirtualRoutersInput{
		MeshName: meshNamePtr,
	}
	for {
		resp, err := a.client.ListVirtualRouters(listVirtualRoutersInput)
		if err != nil {
			return err
		}
		for _, virtualRouterRef := range resp.VirtualRouters {
			virtualRouterNames = append(virtualRouterNames, virtualRouterRef.VirtualRouterName)
		}
		if resp.NextToken == nil {
			break
		}
		listVirtualRoutersInput.NextToken = resp.NextToken
	}
	for _, virtualRouterName := range virtualRouterNames {
		listRoutesInput := &appmesh.ListRoutesInput{
			MeshName:          meshNamePtr,
			VirtualRouterName: virtualRouterName,
		}
		for {
			resp, err := a.client.ListRoutes(listRoutesInput)
			if err != nil {
				return err
			}
			for _, routeRef := range resp.Routes {
				if aws2.StringValue(routeRef.RouteName) == DefaultRouteName {
					_, err := a.client.DeleteRoute(&appmesh.DeleteRouteInput{
						MeshName:          meshNamePtr,
						VirtualRouterName: virtualRouterName,
						RouteName:         routeRef.RouteName,
					})
					if err != nil {
						return err
					}
				}
			}
			if resp.NextToken == nil {
				break
			}
			listRoutesInput.NextToken = resp.NextToken
		}
	}
	return nil
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
