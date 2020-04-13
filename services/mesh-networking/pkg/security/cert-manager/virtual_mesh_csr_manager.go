package cert_manager

import (
	"context"
	"fmt"
	"strings"

	"github.com/avast/retry-go"
	"github.com/rotisserie/eris"
	core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	discovery_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	networking_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	security_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/clients/zephyr/discovery"
	"github.com/solo-io/service-mesh-hub/pkg/env"
	mc_manager "github.com/solo-io/service-mesh-hub/services/common/multicluster/manager"
	vm_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/validation"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	UnsupportedMeshTypeError = func(mesh *discovery_v1alpha1.Mesh) error {
		return eris.Errorf("unsupported mesh type: %T found", mesh.Spec.GetMeshType())
	}
	UnableToGatherCertConfigInfo = func(
		err error,
		mesh *discovery_v1alpha1.Mesh,
		vm *networking_v1alpha1.VirtualMesh,
	) error {
		return eris.Wrapf(err, "unable to produce cert config info for mesh %s in virtual mesh %s",
			mesh.GetName(), vm.GetName())
	}
)

type virtualMeshCsrManager struct {
	meshClient              zephyr_discovery.MeshClient
	meshRefFinder           vm_validation.VirtualMeshFinder
	csrClientFactory        zephyr_security.VirtualMeshCertificateSigningRequestClientFactory
	dynamicClientGetter     mc_manager.DynamicClientGetter
	istioCertConfigProducer IstioCertConfigProducer
}

func NewVirtualMeshCsrProcessor(
	dynamicClientGetter mc_manager.DynamicClientGetter,
	meshClient zephyr_discovery.MeshClient,
	meshRefFinder vm_validation.VirtualMeshFinder,
	csrClientFactory zephyr_security.VirtualMeshCertificateSigningRequestClientFactory,
	istioCertConfigProducer IstioCertConfigProducer,
) VirtualMeshCertificateManager {
	return &virtualMeshCsrManager{
		meshClient:              meshClient,
		dynamicClientGetter:     dynamicClientGetter,
		csrClientFactory:        csrClientFactory,
		meshRefFinder:           meshRefFinder,
		istioCertConfigProducer: istioCertConfigProducer,
	}
}

func (m *virtualMeshCsrManager) InitializeCertificateForVirtualMesh(
	ctx context.Context,
	vm *networking_v1alpha1.VirtualMesh,
) networking_types.VirtualMeshStatus {
	meshes, err := m.meshRefFinder.GetMeshesForVirtualMesh(ctx, vm)
	if err != nil {
		vm.Status.CertificateStatus = &core_types.Status{
			State:   core_types.Status_PROCESSING_ERROR,
			Message: err.Error(),
		}
		return vm.Status
	}
	if err = m.attemptCsrCreate(ctx, vm, meshes); err != nil {
		vm.Status.CertificateStatus = &core_types.Status{
			State:   core_types.Status_PROCESSING_ERROR,
			Message: err.Error(),
		}
		return vm.Status
	}
	vm.Status.CertificateStatus = &core_types.Status{
		State: core_types.Status_ACCEPTED,
	}
	return vm.Status
}

func (m *virtualMeshCsrManager) attemptCsrCreate(
	ctx context.Context,
	vm *networking_v1alpha1.VirtualMesh,
	meshes []*discovery_v1alpha1.Mesh,
) error {
	for _, mesh := range meshes {
		var (
			certConfig *security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig
			meshType   core_types.MeshType
			err        error
		)
		switch mesh.Spec.GetMeshType().(type) {
		case *discovery_types.MeshSpec_Istio:
			meshType = core_types.MeshType_ISTIO
			certConfig, err = m.istioCertConfigProducer.ConfigureCertificateInfo(vm, mesh)
		default:
			return UnsupportedMeshTypeError(mesh)
		}
		if err != nil {
			return UnableToGatherCertConfigInfo(err, mesh, vm)
		}

		clusterName := mesh.Spec.GetCluster().GetName()
		// TODO: check KubernetesCluster resource to see if this retry is worth it
		dynamicClient, err := m.dynamicClientGetter.GetClientForCluster(ctx, clusterName, retry.Attempts(6))
		if err != nil {
			return err
		}
		csrClient := m.csrClientFactory(dynamicClient)
		_, err = csrClient.Get(ctx, m.buildCsrName(strings.ToLower(meshType.String()), vm.GetName()), env.GetWriteNamespace())
		if !errors.IsNotFound(err) {
			if err != nil {
				return err
			}
			// TODO: Handle this case better
			// CSR already exists, continue
			continue
		}

		if err = csrClient.Create(ctx, &security_v1alpha1.VirtualMeshCertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      m.buildCsrName(strings.ToLower(meshType.String()), vm.GetName()),
				Namespace: env.GetWriteNamespace(),
			},
			Spec: security_types.VirtualMeshCertificateSigningRequestSpec{
				VirtualMeshRef: &core_types.ResourceRef{
					Name:      vm.GetName(),
					Namespace: vm.GetNamespace(),
				},
				CertConfig: certConfig,
			},
			Status: security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &core_types.Status{
					Message: "awaiting automated csr generation",
				},
			},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m *virtualMeshCsrManager) buildCsrName(meshType, virtualMeshName string) string {
	return fmt.Sprintf("%s-%s-cert-request", meshType, virtualMeshName)
}
