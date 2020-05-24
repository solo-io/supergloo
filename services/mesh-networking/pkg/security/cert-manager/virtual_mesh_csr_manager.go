package cert_manager

import (
	"context"
	"fmt"
	"strings"

	"github.com/avast/retry-go"
	"github.com/rotisserie/eris"
	zephyr_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	zephyr_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1/types"
	zephyr_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1"
	zephyr_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.zephyr.solo.io/v1alpha1/types"
	zephyr_security "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1"
	zephyr_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.zephyr.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/kube/multicluster"
	vm_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/validation"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	UnsupportedMeshTypeError = func(mesh *zephyr_discovery.Mesh) error {
		return eris.Errorf("unsupported mesh type: %T found", mesh.Spec.GetMeshType())
	}
	UnableToGatherCertConfigInfo = func(
		err error,
		mesh *zephyr_discovery.Mesh,
		vm *zephyr_networking.VirtualMesh,
	) error {
		return eris.Wrapf(err, "unable to produce cert config info for mesh %s in virtual mesh %s",
			mesh.GetName(), vm.GetName())
	}
)

type virtualMeshCsrManager struct {
	meshClient              zephyr_discovery.MeshClient
	meshRefFinder           vm_validation.VirtualMeshFinder
	csrClientFactory        zephyr_security.VirtualMeshCertificateSigningRequestClientFactory
	dynamicClientGetter     multicluster.DynamicClientGetter
	istioCertConfigProducer IstioCertConfigProducer
}

func NewVirtualMeshCsrProcessor(
	dynamicClientGetter multicluster.DynamicClientGetter,
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
	vm *zephyr_networking.VirtualMesh,
) zephyr_networking_types.VirtualMeshStatus {
	meshes, err := m.meshRefFinder.GetMeshesForVirtualMesh(ctx, vm)
	if err != nil {
		vm.Status.CertificateStatus = &zephyr_core_types.Status{
			State:   zephyr_core_types.Status_PROCESSING_ERROR,
			Message: err.Error(),
		}
		return vm.Status
	}
	if err = m.attemptCsrCreate(ctx, vm, meshes); err != nil {
		vm.Status.CertificateStatus = &zephyr_core_types.Status{
			State:   zephyr_core_types.Status_PROCESSING_ERROR,
			Message: err.Error(),
		}
		return vm.Status
	}
	vm.Status.CertificateStatus = &zephyr_core_types.Status{
		State: zephyr_core_types.Status_ACCEPTED,
	}
	return vm.Status
}

func (m *virtualMeshCsrManager) attemptCsrCreate(
	ctx context.Context,
	vm *zephyr_networking.VirtualMesh,
	meshes []*zephyr_discovery.Mesh,
) error {
	for _, mesh := range meshes {
		var (
			certConfig *zephyr_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig
			meshType   zephyr_core_types.MeshType
			err        error
		)
		switch mesh.Spec.GetMeshType().(type) {
		case *zephyr_discovery_types.MeshSpec_Istio1_5_, *zephyr_discovery_types.MeshSpec_Istio1_6_:
			meshType, err = metadata.MeshToMeshType(mesh)
			if err != nil {
				return err
			}

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
		_, err = csrClient.GetVirtualMeshCertificateSigningRequest(
			ctx,
			client.ObjectKey{
				Name:      m.buildCsrName(strings.ToLower(meshType.String()), vm.GetName()),
				Namespace: container_runtime.GetWriteNamespace(),
			},
		)
		if !errors.IsNotFound(err) {
			if err != nil {
				return err
			}
			// TODO: Handle this case better
			// CSR already exists, continue
			continue
		}

		if err = csrClient.CreateVirtualMeshCertificateSigningRequest(ctx, &zephyr_security.VirtualMeshCertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      m.buildCsrName(strings.ToLower(meshType.String()), vm.GetName()),
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: zephyr_security_types.VirtualMeshCertificateSigningRequestSpec{
				VirtualMeshRef: &zephyr_core_types.ResourceRef{
					Name:      vm.GetName(),
					Namespace: vm.GetNamespace(),
				},
				CertConfig: certConfig,
			},
			Status: zephyr_security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &zephyr_core_types.Status{
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
