package gloo

import (
	"context"
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"k8s.io/api/extensions/v1beta1"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	kubev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type glooConfigSyncer struct {
	cs       *clientset.Clientset
	reporter reporter.Reporter
}

func NewGlooRegistrationSyncer(reporter reporter.Reporter, cs *clientset.Clientset) v1.RegistrationSyncer {
	return &glooConfigSyncer{reporter: reporter, cs: cs}
}

func (s *glooConfigSyncer) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("gloo-config-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v: %v", snap.Hash(), snap.Stringer())
	defer logger.Infof("end sync %v", snap.Hash())
	logger.Debugf("full snapshot: %v", snap)

	glooMeshIngresses := make(v1.MeshIngressList, 0)
	for _, meshIngress := range snap.Meshingresses.List() {
		if _, ok := meshIngress.MeshIngressType.(*v1.MeshIngress_Gloo); ok {
			glooMeshIngresses = append(glooMeshIngresses, meshIngress)
		}
	}

	errs := reporter.ResourceErrors{}
	for _, glooIngress := range glooMeshIngresses {
		if err := s.handleGlooMeshIngressConfig(glooIngress, snap.Meshes.List()); err != nil {
			errs.AddError(glooIngress, err)
			logger.Errorf("unable to update gloo ingress %v, %s", glooIngress.Metadata, err)
		}
	}

	logger.Infof("sync completed successfully!")
	return s.reporter.WriteReports(ctx, errs, nil)
}

func (s *glooConfigSyncer) handleGlooMeshIngressConfig(ingress *v1.MeshIngress, meshes v1.MeshList) error {

	glooMeshIngress, isGloo := ingress.MeshIngressType.(*v1.MeshIngress_Gloo)
	if !isGloo {
		return errors.Errorf("only gloo mesh ingress currently supported")
	}
	targetMeshes := ingress.Meshes

	deployment, err := s.cs.Kube.ExtensionsV1beta1().Deployments(glooMeshIngress.Gloo.InstallationNamespace).Get("gateway-proxy", kubev1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to find deployemt for gateway-proxy in %s", glooMeshIngress.Gloo.InstallationNamespace)
	}

	update, err := ShouldUpdateDeployment(deployment, targetMeshes, meshes)
	if err != nil {
		return err
	}

	if update {
		_, err := s.cs.Kube.ExtensionsV1beta1().Deployments(glooMeshIngress.Gloo.InstallationNamespace).Update(deployment)
		if err != nil {
			return errors.Wrapf(err, "unable to rewrite deployment after update")
		}
	}

	return nil
}

func ShouldUpdateDeployment(deployment *v1beta1.Deployment, targetMeshes []*core.ResourceRef, meshes v1.MeshList) (bool, error) {
	var volumes VolumeList = deployment.Spec.Template.Spec.Volumes
	gatewayProxyContainer := deployment.Spec.Template.Spec.Containers[0]
	var mounts VolumeMountList = gatewayProxyContainer.VolumeMounts

	newDeploymentVolumes, err := ResourcesToDeploymentInfo(targetMeshes, meshes)
	if err != nil {
		return false, err
	}
	oldDeploymentVolumes := VolumesToDeploymentInfo(volumes, mounts)

	added, deleted := Diff(newDeploymentVolumes, oldDeploymentVolumes)
	updated := len(added) > 0 || len(deleted) > 0

	if updated {
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
	}
	return updated, nil
}
