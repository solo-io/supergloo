package restart

import (
	"context"

	corev1 "github.com/solo-io/external-apis/pkg/api/k8s/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/commands/mesh/flags"
	"github.com/solo-io/service-mesh-hub/pkg/meshctl/utils"
	skcorev1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Command(ctx context.Context, opts *flags.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart all pods in the specified mesh",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := utils.BuildClient(opts.Kubeconfig, opts.Kubeconfig)
			if err != nil {
				return err
			}
			return restartPods(ctx, c, &skcorev1.ObjectRef{
				Name:      opts.MeshName,
				Namespace: opts.MeshNamespace,
			})
		},
	}

	return cmd
}

func restartPods(ctx context.Context, c client.Client, meshRef *skcorev1.ObjectRef) error {
	podClient := corev1.NewPodClient(c)
	meshWorkloadClient := v1alpha2.NewMeshWorkloadClient(c)
	meshWorkloadList, err := meshWorkloadClient.ListMeshWorkload(ctx)
	if err != nil {
		return err
	}
	for _, meshWorkload := range meshWorkloadList.Items {
		// currently only supports restarting k8s workloads
		if !ezkube.RefsMatch(meshWorkload.Spec.Mesh, meshRef) || meshWorkload.Spec.GetKubernetes() == nil {
			continue
		}
		podLabels := meshWorkload.Spec.GetKubernetes().PodLabels
		// ignore if no workload selectors populated to avoid restarting all pods
		if len(podLabels) < 1 {
			continue
		}
		podList, err := podClient.ListPod(ctx, &client.ListOptions{
			LabelSelector: labels.Set(podLabels).AsSelector(),
		})
		if err != nil {
			return err
		}
		for _, pod := range podList.Items {
			err := podClient.DeletePod(ctx, client.ObjectKey{Name: pod.Name, Namespace: pod.Namespace})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
