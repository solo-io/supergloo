package appmesh

import (
	"context"
	"strings"

	"github.com/rotisserie/eris"
	v1beta2sets "github.com/solo-io/external-apis/pkg/api/appmesh/appmesh.k8s.aws/v1beta2/sets"
	v1sets "github.com/solo-io/external-apis/pkg/api/k8s/core/v1/sets"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	decorator "github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/decorator"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/workload/types"
	"github.com/solo-io/skv2/contrib/pkg/sets"
	"go.uber.org/zap"
)

const (
	appMeshVirtualNodeName = "APPMESH_VIRTUAL_NODE_NAME"
)

type appMeshWorkloadDecorator struct {
	ctx          context.Context
	virtualNodes v1beta2sets.VirtualNodeSet
}

func NewAppMeshWorkloadDecorator(
	ctx context.Context,
	virtualNodes v1beta2sets.VirtualNodeSet,
) decorator.WorkloadDecorator {
	return &appMeshWorkloadDecorator{
		ctx:          ctx,
		virtualNodes: virtualNodes,
	}
}

func (d *appMeshWorkloadDecorator) DecorateWorkload(discoveredWorkload *v1alpha2.Workload, kubeWorkload types.Workload, mesh *v1alpha2.Mesh, pods v1sets.PodSet) {
	if mesh.Spec.GetAwsAppMesh() == nil {
		// Not an App Mesh workload, do not decorate
		return
	}

	virtualNodeArn, err := getVirtualNodeArn(kubeWorkload, pods, d.virtualNodes)
	if err != nil {
		contextutils.LoggerFrom(d.ctx).Debugw("Failed to get VirtualNodeARN", zap.String("workload", sets.Key(discoveredWorkload)))
		return
	}

	if discoveredWorkload.Spec.AppMesh == nil {
		discoveredWorkload.Spec.AppMesh = &v1alpha2.WorkloadSpec_AppMesh{
			VirtualNodeArn: virtualNodeArn,
		}
	}
}

func getVirtualNodeArn(kubeWorkload types.Workload, workloadPods v1sets.PodSet, virtualNodes v1beta2sets.VirtualNodeSet) (string, error) {
	virtualNodeName := getVirtualNodeName(workloadPods)
	if virtualNodeName == "" {
		return "", eris.Errorf("found no virtual node name for kube workload %s", sets.Key(kubeWorkload))
	}

	for _, node := range virtualNodes.List() {
		arn := node.Status.VirtualNodeARN
		if arn == nil {
			// VirtualNode resource has not been processed by the App Mesh controller
			continue
		}

		/*
			Virtual Node name: "mesh/mesh-name/virtualNode/virtual-node-name"
			Virtual Node ARN: arn:aws:appmesh:<region>:<account-id>:mesh/mesh-name/virtualNode/virtual-node-name

			Thus we lose some specificity around the region and AWS Account ID.
		*/
		if strings.HasSuffix(*arn, virtualNodeName) {
			return *arn, nil
		}
	}

	return "", eris.Errorf("found no VirtualNode matching Virtual Node Name %s", virtualNodeName)
}

func getVirtualNodeName(workloadPods v1sets.PodSet) string {
	for _, pod := range workloadPods.List() {
		for _, container := range pod.Spec.Containers {
			for _, envVar := range container.Env {
				if envVar.Name == appMeshVirtualNodeName {
					return envVar.Value
				}
			}
		}
	}
	return ""
}
