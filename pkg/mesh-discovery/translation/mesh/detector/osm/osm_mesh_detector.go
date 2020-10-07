package osm

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/input"

	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/mesh/detector"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/translation/utils"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/utils/dockerutils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

const (
	osmControlPlaneDeploymentName = "osm-controller"
)

type meshDetector struct {
	ctx context.Context
}

func NewMeshDetector(
	ctx context.Context,
) detector.MeshDetector {
	return &meshDetector{
		ctx: ctx,
	}
}

// returns a mesh for each deployment that contains the osm controller image
func (d *meshDetector) DetectMeshes(in input.Snapshot) (v1alpha2.MeshSlice, error) {
	var meshes v1alpha2.MeshSlice
	var errs error
	for _, deployment := range in.Deployments().List() {
		mesh, err := d.detectMesh(deployment)
		if err != nil {
			errs = multierror.Append(errs, err)
		}
		if mesh == nil {
			continue
		}
		meshes = append(meshes, mesh)
	}
	return meshes, errs
}

func (d *meshDetector) detectMesh(deployment *appsv1.Deployment) (*v1alpha2.Mesh, error) {
	version, err := getOsmControllerVersion(deployment)
	if err != nil {
		return nil, err
	}
	if version == "" {
		return nil, nil
	}
	return &v1alpha2.Mesh{
		ObjectMeta: utils.DiscoveredObjectMeta(deployment),
		Spec: v1alpha2.MeshSpec{
			MeshType: &v1alpha2.MeshSpec_Osm{
				Osm: &v1alpha2.MeshSpec_OSM{
					Installation: &v1alpha2.MeshSpec_MeshInstallation{
						Namespace: deployment.Namespace,
						Cluster:   deployment.ClusterName,
						PodLabels: deployment.Spec.Selector.MatchLabels,
						Version:   version,
					},
				},
			},
		},
	}, nil
}

func getOsmControllerVersion(deployment *appsv1.Deployment) (string, error) {
	for _, container := range deployment.Spec.Template.Spec.Containers {
		if isOsmController(deployment, &container) {
			parsedImage, err := dockerutils.ParseImageName(container.Image)
			if err != nil {
				return "", eris.Wrapf(err, "failed to parse osm-controller image tag: %s", container.Image)
			}
			version := parsedImage.Tag
			if parsedImage.Digest != "" {
				version = parsedImage.Digest
			}
			return version, nil
		}
	}
	return "", nil
}

// Return true if deployment is inferred to be an OSM deployment
func isOsmController(deployment *appsv1.Deployment, container *corev1.Container) bool {
	return deployment.GetName() == osmControlPlaneDeploymentName &&
		strings.Contains(container.Image, osmControlPlaneDeploymentName)
}
