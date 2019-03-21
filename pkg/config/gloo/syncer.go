package gloo

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	corev1 "k8s.io/api/core/v1"
	kubev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type glooConfigSyncer struct {
	reporter reporter.Reporter
	cs       *clientset.Clientset
}

func NewGlooConfigSyncer(reporter reporter.Reporter, cs *clientset.Clientset) v1.ConfigSyncer {
	return &glooConfigSyncer{reporter: reporter, cs: cs}
}

func (s *glooConfigSyncer) Sync(ctx context.Context, snap *v1.ConfigSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("gloo-config-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v: %v", snap.Hash(), snap.Stringer())
	defer logger.Infof("end sync %v", snap.Hash())
	logger.Debugf("full snapshot: %v", snap)

	var resourceErrs reporter.ResourceErrors

	glooMeshIngresses := make(v1.MeshIngressList, 0)

	for _, meshIngress := range snap.Meshingresses.List() {
		if _, ok := meshIngress.MeshIngressType.(*v1.MeshIngress_Gloo); ok {
			glooMeshIngresses = append(glooMeshIngresses, meshIngress)
		}
	}

	// finally, write reports
	if err := s.reporter.WriteReports(ctx, resourceErrs, nil); err != nil {
		return errors.Wrapf(err, "writing reports")
	}

	logger.Infof("sync completed successfully!")
	return nil
}

func (s *glooConfigSyncer) handleGlooMeshIngressConfig(ingress v1.MeshIngress, meshes v1.MeshList) error {

	glooMeshIngress, isGloo := ingress.MeshIngressType.(*v1.MeshIngress_Gloo)
	if !isGloo {
		return errors.Errorf("only gloo mesh ingress currently supported")
	}
	targetMeshes := ingress.Meshes

	deployment, err := s.cs.Kube.ExtensionsV1beta1().Deployments(glooMeshIngress.Gloo.InstallationNamespace).Get("gateway-proxy", kubev1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to find deployemt for gateway-proxy in %s", glooMeshIngress.Gloo.InstallationNamespace)
	}
	volumes := deployment.Spec.Template.Spec.Volumes
	gatewayProxyContainer := deployment.Spec.Template.Spec.Containers[0]
	mounts := gatewayProxyContainer.VolumeMounts

	newDeploymentVolumes, err := ResourcesToDeploymentInfo(targetMeshes, meshes)
	oldDeploymentVolumes := VolumesToDeploymentInfo(volumes, mounts)

	added, deleted := diff(newDeploymentVolumes, oldDeploymentVolumes)
	updated := len(added) > 0 || len(deleted) > 0

	if updated {
		for _, v := range added {
			volumes = append(volumes, v.Volume)
			mounts = append(mounts, v.VolumeMount)
		}

		for _, v := range deleted {
			var tempVolumes []corev1.Volume
			copy(tempVolumes, volumes)
			for i, possibleDelete := range tempVolumes {
				if v.Volume.Name == possibleDelete.Name {
					removeVolume(volumes, i)
				}
			}
			var tempVolumeMounts []corev1.VolumeMount
			copy(tempVolumeMounts, mounts)
			for i, possibleDelete := range tempVolumeMounts {
				if v.VolumeMount.Name == possibleDelete.Name {
					removeVolumeMount(mounts, i)
				}
			}
		}
		gatewayProxyContainer.VolumeMounts = mounts
		deployment.Spec.Template.Spec.Containers[0] = gatewayProxyContainer
		deployment.Spec.Template.Spec.Volumes = volumes

		_, err := s.cs.Kube.ExtensionsV1beta1().Deployments(glooMeshIngress.Gloo.InstallationNamespace).Update(deployment)
		if err != nil {
			return errors.Wrapf(err, "unable to rewrite deployment after update")
		}
	}

	return nil
}

func removeVolume(s []corev1.Volume, i int) []corev1.Volume {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}

func removeVolumeMount(s []corev1.VolumeMount, i int) []corev1.VolumeMount {
	s[i] = s[len(s)-1]
	return s[:len(s)-1]
}
