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

type glooMtlsSyncer struct {
	cs       *clientset.Clientset
	reporter reporter.Reporter
}

func NewGlooRegistrationSyncer(reporter reporter.Reporter, cs *clientset.Clientset) v1.RegistrationSyncer {
	return &glooMtlsSyncer{reporter: reporter, cs: cs}
}

func (s *glooMtlsSyncer) Sync(ctx context.Context, snap *v1.RegistrationSnapshot) error {
	ctx = contextutils.WithLogger(ctx, fmt.Sprintf("gloo-config-sync-%v", snap.Hash()))
	logger := contextutils.LoggerFrom(ctx)
	logger.Infof("begin sync %v: %v", snap.Hash(), snap.Stringer())
	defer logger.Infof("end sync %v", snap.Hash())
	logger.Debugf("full snapshot: %v", snap)

	var glooMeshIngresses v1.MeshIngressList
	for _, meshIngress := range snap.Meshingresses.List() {
		if _, ok := meshIngress.MeshIngressType.(*v1.MeshIngress_Gloo); ok {
			glooMeshIngresses = append(glooMeshIngresses, meshIngress)
		}
	}

	errs := reporter.ResourceErrors{}
	for _, glooIngress := range glooMeshIngresses {
		if err := s.handleGlooMeshIngressConfig(ctx, glooIngress, snap.Meshes.List()); err != nil {
			errs.AddError(glooIngress, err)
			logger.Errorf("unable to update gloo ingress %v, %s", glooIngress.Metadata, err)
		}
	}

	logger.Infof("sync completed successfully!")
	return s.reporter.WriteReports(ctx, errs, nil)
}

func (s *glooMtlsSyncer) handleGlooMeshIngressConfig(ctx context.Context, ingress *v1.MeshIngress, meshes v1.MeshList) error {
	logger := contextutils.LoggerFrom(ctx)

	glooMeshIngress, isGloo := ingress.MeshIngressType.(*v1.MeshIngress_Gloo)
	if !isGloo {
		return errors.Errorf("only gloo mesh ingress currently supported")
	}
	targetMeshes := ingress.Meshes

	deployment, err := s.cs.Kube.ExtensionsV1beta1().Deployments(glooMeshIngress.Gloo.InstallationNamespace).Get("gateway-proxy", kubev1.GetOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to find deployemt for gateway-proxy in %s", glooMeshIngress.Gloo.InstallationNamespace)
	}

	update, err := shouldUpdateDeployment(deployment, targetMeshes, meshes)
	if err != nil {
		return err
	}

	if !update {
		return nil
	}

	logger.Infof("about to modify deployment for %s.%s", deployment.Namespace, deployment.Name)
	_, err = s.cs.Kube.ExtensionsV1beta1().Deployments(glooMeshIngress.Gloo.InstallationNamespace).Update(deployment)
	if err != nil {
		return errors.Wrapf(err, "unable to rewrite deployment after update")
	}

	return nil
}

func shouldUpdateDeployment(deployment *v1beta1.Deployment, targetMeshes []*core.ResourceRef, meshes v1.MeshList) (bool, error) {
	var volumes VolumeList = deployment.Spec.Template.Spec.Volumes
	gatewayProxyContainer := deployment.Spec.Template.Spec.Containers[0]
	var mounts VolumeMountList = gatewayProxyContainer.VolumeMounts

	newDeploymentVolumes, err := resourcesToDeploymentInfo(targetMeshes, meshes)
	if err != nil {
		return false, err
	}
	oldDeploymentVolumes := volumesToDeploymentInfo(volumes, mounts)

	added, deleted := diff(newDeploymentVolumes, oldDeploymentVolumes)
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
