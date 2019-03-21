package gloo

import (
	"context"
	"fmt"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/reporter"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	kubev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type glooConfigSyncer struct {
	reporter reporter.Reporter
	cs       *clientset.Clientset
}

var (
	defaultMode int32 = 420
	optional    bool  = true
)

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
	mounts := deployment.Spec.Template.Spec.Containers[0].VolumeMounts

	newDeploymentVolumes, err := ResourcesToDeploymentInfo(targetMeshes, meshes)
	oldDeploymentVolumes := VolumesToDeploymentInfo(volumes, mounts)

	added, deleted := diff(newDeploymentVolumes, oldDeploymentVolumes)

	if len(added) > 0 || len(deleted) > 0 {

	}

	return nil
}
