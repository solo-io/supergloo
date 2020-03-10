package cert_manager

import (
	"context"
	"fmt"
	"strings"

	"github.com/avast/retry-go"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	security_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/mesh-projects/pkg/clients/zephyr/discovery"
	zephyr_security "github.com/solo-io/mesh-projects/pkg/clients/zephyr/security"
	"github.com/solo-io/mesh-projects/pkg/env"
	mc_manager "github.com/solo-io/mesh-projects/services/common/multicluster/manager"
	group_validation "github.com/solo-io/mesh-projects/services/mesh-networking/pkg/validation"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	DynamicClientDoesNotExistForClusterError = func(clusterName string) error {
		return eris.Errorf("could not find a dynamic client for cluster %s", clusterName)
	}
	UnsupportedMeshTypeError = func(mesh *discovery_v1alpha1.Mesh) error {
		return eris.Errorf("unsupported mesh type: %T found", mesh.Spec.GetMeshType())
	}
	UnableToGatherCertConfigInfo = func(
		err error,
		mesh *discovery_v1alpha1.Mesh,
		mg *networking_v1alpha1.MeshGroup,
	) error {
		return eris.Wrapf(err, "unable to produce cert config info for mesh %s in mesh group %s",
			mesh.GetName(), mg.GetName())
	}
)

type meshGroupCsrManager struct {
	meshClient              zephyr_discovery.MeshClient
	meshRefFinder           group_validation.GroupMeshFinder
	csrClientFactory        zephyr_security.MeshGroupCSRClientFactory
	dynamicClientGetter     mc_manager.DynamicClientGetter
	istioCertConfigProducer IstioCertConfigProducer
}

func NewMeshGroupCsrProcessor(
	dynamicClientGetter mc_manager.DynamicClientGetter,
	meshClient zephyr_discovery.MeshClient,
	meshRefFinder group_validation.GroupMeshFinder,
	csrClientFactory zephyr_security.MeshGroupCSRClientFactory,
	istioCertConfigProducer IstioCertConfigProducer,
) MeshGroupCertificateManager {
	return &meshGroupCsrManager{
		meshClient:              meshClient,
		dynamicClientGetter:     dynamicClientGetter,
		csrClientFactory:        csrClientFactory,
		meshRefFinder:           meshRefFinder,
		istioCertConfigProducer: istioCertConfigProducer,
	}
}

func (m *meshGroupCsrManager) InitializeCertificateForMeshGroup(
	ctx context.Context,
	mg *networking_v1alpha1.MeshGroup,
) networking_types.MeshGroupStatus {
	meshes, err := m.meshRefFinder.GetMeshesForGroup(ctx, mg)
	if err != nil {
		mg.Status.CertificateStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_INVALID,
			Message: err.Error(),
		}
		return mg.Status
	}
	if err = m.attemptCsrCreate(ctx, mg, meshes); err != nil {
		mg.Status.CertificateStatus = &core_types.ComputedStatus{
			Status:  core_types.ComputedStatus_INVALID,
			Message: err.Error(),
		}
		return mg.Status
	}
	mg.Status.CertificateStatus = &core_types.ComputedStatus{
		Status: core_types.ComputedStatus_ACCEPTED,
	}
	return mg.Status
}

func (m *meshGroupCsrManager) attemptCsrCreate(
	ctx context.Context,
	mg *networking_v1alpha1.MeshGroup,
	meshes []*discovery_v1alpha1.Mesh,
) error {
	for _, mesh := range meshes {
		var (
			certConfig *security_types.CertConfig
			meshType   core_types.MeshType
			err        error
		)
		switch mesh.Spec.GetMeshType().(type) {
		case *discovery_types.MeshSpec_Istio:
			meshType = core_types.MeshType_ISTIO
			certConfig, err = m.istioCertConfigProducer.ConfigureCertificateInfo(mg, mesh)
		default:
			return UnsupportedMeshTypeError(mesh)
		}
		if err != nil {
			return UnableToGatherCertConfigInfo(err, mesh, mg)
		}

		clusterName := mesh.Spec.GetCluster().GetName()
		// TODO: check KubernetesCluster resource to see if this retry is worth it
		dynamicClient, ok := m.dynamicClientGetter.GetClientForCluster(clusterName, retry.Attempts(6))
		if !ok {
			return DynamicClientDoesNotExistForClusterError(clusterName)
		}
		csrClient := m.csrClientFactory(dynamicClient)
		_, err = csrClient.Get(ctx, m.buildCsrName(strings.ToLower(meshType.String()), mg.GetName()), env.GetWriteNamespace())
		if !errors.IsNotFound(err) {
			if err != nil {
				return err
			}
			// TODO: Handle this case better
			// CSR already exists, continue
			continue
		}

		if err = csrClient.Create(ctx, &security_v1alpha1.MeshGroupCertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      m.buildCsrName(strings.ToLower(meshType.String()), mg.GetName()),
				Namespace: env.GetWriteNamespace(),
			},
			Spec: security_types.MeshGroupCertificateSigningRequestSpec{
				MeshGroupRef: &core_types.ResourceRef{
					Name:      mg.GetName(),
					Namespace: mg.GetNamespace(),
				},
				CertConfig: certConfig,
			},
			Status: security_types.MeshGroupCertificateSigningRequestStatus{
				ComputedStatus: &core_types.ComputedStatus{
					Message: "awaiting automated csr generation",
				},
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m *meshGroupCsrManager) buildCsrName(meshType, meshGroupName string) string {
	return fmt.Sprintf("%s-%s-cert-request", meshType, meshGroupName)
}
