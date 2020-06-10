package cert_manager

import (
	"context"
	"fmt"
	"strings"

	"github.com/avast/retry-go"
	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/contextutils"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_discovery_types "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	smh_networking_types "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/types"
	smh_security "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
	smh_security_types "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1/types"
	container_runtime "github.com/solo-io/service-mesh-hub/pkg/container-runtime"
	"github.com/solo-io/service-mesh-hub/pkg/kube/metadata"
	"github.com/solo-io/service-mesh-hub/pkg/kube/multicluster"
	vm_validation "github.com/solo-io/service-mesh-hub/services/mesh-networking/pkg/validation"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	UnsupportedMeshTypeError = func(mesh *smh_discovery.Mesh) error {
		return eris.Errorf("unsupported mesh type: %T found", mesh.Spec.GetMeshType())
	}
	UnableToGatherCertConfigInfo = func(
		err error,
		mesh *smh_discovery.Mesh,
		vm *smh_networking.VirtualMesh,
	) error {
		return eris.Wrapf(err, "unable to produce cert config info for mesh %s in virtual mesh %s",
			mesh.GetName(), vm.GetName())
	}
)

type virtualMeshCsrManager struct {
	meshClient              smh_discovery.MeshClient
	meshRefFinder           vm_validation.VirtualMeshFinder
	csrClientFactory        smh_security.VirtualMeshCertificateSigningRequestClientFactory
	dynamicClientGetter     multicluster.DynamicClientGetter
	istioCertConfigProducer IstioCertConfigProducer
}

func NewVirtualMeshCsrProcessor(
	dynamicClientGetter multicluster.DynamicClientGetter,
	meshClient smh_discovery.MeshClient,
	meshRefFinder vm_validation.VirtualMeshFinder,
	csrClientFactory smh_security.VirtualMeshCertificateSigningRequestClientFactory,
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
	vm *smh_networking.VirtualMesh,
) smh_networking_types.VirtualMeshStatus {
	logger := contextutils.LoggerFrom(ctx)
	meshes, err := m.meshRefFinder.GetMeshesForVirtualMesh(ctx, vm)
	if err != nil {
		logger.Debugf("Hit error case 1 %s", err.Error())
		vm.Status.CertificateStatus = &smh_core_types.Status{
			State:   smh_core_types.Status_PROCESSING_ERROR,
			Message: err.Error(),
		}
		return vm.Status
	}
	if err = m.attemptCsrCreate(ctx, vm, meshes); err != nil {
		logger.Debugf("Hit error case 2 %s", err.Error())
		vm.Status.CertificateStatus = &smh_core_types.Status{
			State:   smh_core_types.Status_PROCESSING_ERROR,
			Message: err.Error(),
		}
		return vm.Status
	}
	logger.Debugf("Accepted")
	vm.Status.CertificateStatus = &smh_core_types.Status{
		State: smh_core_types.Status_ACCEPTED,
	}
	return vm.Status
}

func (m *virtualMeshCsrManager) attemptCsrCreate(
	ctx context.Context,
	vm *smh_networking.VirtualMesh,
	meshes []*smh_discovery.Mesh,
) error {
	logger := contextutils.LoggerFrom(ctx)
	logger.Debugf("Got %d member meshes", len(meshes))
	for _, mesh := range meshes {
		logger.Debugf("Processing mesh %s.%s", mesh.Name, mesh.Namespace)
		var (
			certConfig *smh_security_types.VirtualMeshCertificateSigningRequestSpec_CertConfig
			meshType   smh_core_types.MeshType
			err        error
		)
		switch mesh.Spec.GetMeshType().(type) {
		case *smh_discovery_types.MeshSpec_Istio1_5_, *smh_discovery_types.MeshSpec_Istio1_6_:
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

		if err = csrClient.CreateVirtualMeshCertificateSigningRequest(ctx, &smh_security.VirtualMeshCertificateSigningRequest{
			ObjectMeta: metav1.ObjectMeta{
				Name:      m.buildCsrName(strings.ToLower(meshType.String()), vm.GetName()),
				Namespace: container_runtime.GetWriteNamespace(),
			},
			Spec: smh_security_types.VirtualMeshCertificateSigningRequestSpec{
				VirtualMeshRef: &smh_core_types.ResourceRef{
					Name:      vm.GetName(),
					Namespace: vm.GetNamespace(),
				},
				CertConfig: certConfig,
			},
			Status: smh_security_types.VirtualMeshCertificateSigningRequestStatus{
				ComputedStatus: &smh_core_types.Status{
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
	return strings.ReplaceAll(fmt.Sprintf("%s-%s-cert-request", meshType, virtualMeshName), "_", "-")
}
