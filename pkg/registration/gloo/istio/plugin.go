package istio

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/api/extensions/v1beta1"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	kubev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type glooIstioMtlsPlugin struct {
	cs *clientset.Clientset
}

func NewGlooIstioMtlsPlugin(cs *clientset.Clientset) *glooIstioMtlsPlugin {
	return &glooIstioMtlsPlugin{cs: cs}
}

func (pl *glooIstioMtlsPlugin) HandleMeshes(ctx context.Context, ingress *v1.MeshIngress, meshes v1.MeshList) error {
	ctx = contextutils.WithLoggerValues(ctx,
		zap.String("plugin", "istio-gloo-mtls"),
		zap.String("mesh-ingress", ingress.Metadata.Ref().Key()),
	)
	logger := contextutils.LoggerFrom(ctx)

	deployment, err := pl.cs.Kube.ExtensionsV1beta1().Deployments(ingress.InstallationNamespace).Get("gateway-proxy", kubev1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to find deployemt for gateway-proxy in %s", ingress.InstallationNamespace)
	}
	var istioMeshes v1.MeshList
	for _, mesh := range meshes {
		if istioMesh := mesh.GetIstio(); istioMesh != nil {
			istioMeshes = append(istioMeshes, mesh)
		}
	}

	// This mutates the deployment
	update, err := shouldUpdateDeployment(deployment, istioMeshes)
	if err != nil {
		return err
	}

	if !update {
		return nil
	}

	logger.Infof("about to modify deployment for %s.%s", deployment.Namespace, deployment.Name)
	_, err = pl.cs.Kube.ExtensionsV1beta1().Deployments(ingress.InstallationNamespace).Update(deployment)
	if err != nil {
		return errors.Wrapf(err, "unable to rewrite deployment after update")
	}

	return nil
}

func shouldUpdateDeployment(deployment *v1beta1.Deployment, targetMeshes v1.MeshList) (bool, error) {
	var volumes VolumeList = deployment.Spec.Template.Spec.Volumes
	gatewayProxyContainer := deployment.Spec.Template.Spec.Containers[0]
	var mounts VolumeMountList = gatewayProxyContainer.VolumeMounts

	secretVolumeInfo, err := makeSecretVolumesForMeshes(targetMeshes)
	if err != nil {
		return false, err
	}
	oldDeploymentVolumes := volumesToDeploymentInfo(volumes, mounts)

	added, deleted := diff(secretVolumeInfo, oldDeploymentVolumes)
	updated := len(added) > 0 || len(deleted) > 0

	if !updated {
		return false, nil
	}
	for _, v := range added {
		volumes = append(volumes, v.Volume)
		mounts = append(mounts, v.VolumeMount)
	}

	for _, v := range deleted {
		tempVolumes := make(VolumeList, len(volumes))
		copy(tempVolumes, volumes)
		for i, possibleDelete := range tempVolumes {
			if v.Volume.Name == possibleDelete.Name {
				volumes = volumes.Remove(i)
			}
		}

		tempVolumeMounts := make(VolumeMountList, len(mounts))
		copy(tempVolumeMounts, mounts)
		for i, possibleDelete := range tempVolumeMounts {
			if v.VolumeMount.Name == possibleDelete.Name {
				mounts = mounts.Remove(i)
			}
		}
	}
	gatewayProxyContainer.VolumeMounts = mounts
	deployment.Spec.Template.Spec.Containers[0] = gatewayProxyContainer
	deployment.Spec.Template.Spec.Volumes = volumes
	return true, nil
}
