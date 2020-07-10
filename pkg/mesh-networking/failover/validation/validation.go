package validation

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	v1alpha1sets2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1/types"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	v1alpha1sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1/sets"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-networking/failover"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"istio.io/istio/pkg/config/protocol"
	"k8s.io/apimachinery/pkg/util/validation"
)

//go:generate mockgen -source ./validation.go -destination ./mocks/mock_validation.go -package mock_failover_service_validation

/*
A valid FailoverService must satisfy the following constraints:

1. TargetService must exist
2. Must consist of at least 1 failover service.
3. All declared failover services must exist.
4. All declared failover services must be owned by a supported Mesh type (currently only Istio).
5. All declared failover services must exist in the same VirtualMesh, or belong to a common parent Mesh.
6. All declared failover services must have OutlierDetection settings declared in a TP (grab this from the MeshService status).
*/
type FailoverServiceValidator interface {
	// Set the validation status for FailoverServices in the InputSnapshot
	Validate(snapshot failover.InputSnapshot)
}

var (
	MissingHostname         = eris.New("Missing required field \"hostname\".")
	MissingPort             = eris.New("Missing required field \"port\".")
	MissingMeshes           = eris.New("Missing required field \"meshes\".")
	MissingServices         = eris.New("There must be at least one service declared for the FailoverService.")
	FailoverServiceNotFound = func(serviceRef *v1.ClusterObjectRef) error {
		return eris.Errorf("Failover service %s.%s.%s not found in SMH discovery resources.",
			serviceRef.GetName(),
			serviceRef.GetNamespace(),
			serviceRef.GetClusterName())
	}
	MeshNotFound = func(meshRef *smh_core_types.ResourceRef, serviceRef *smh_core_types.ResourceRef) error {
		return eris.Errorf("Mesh %s.%s.%s for service %s.%s.%s not found in SMH discovery resources.",
			meshRef.GetName(),
			meshRef.GetNamespace(),
			meshRef.GetCluster(),
			serviceRef.GetName(),
			serviceRef.GetNamespace(),
			serviceRef.GetCluster())
	}
	UnsupportedMeshType = func(meshType interface{}) error {
		return eris.Errorf("Unsupported Mesh type %T", meshType)
	}
	MeshWithoutParentVM = func(mesh *smh_discovery.Mesh) error {
		return eris.Errorf("Mesh %s.%s is not grouped in a VirtualMesh.", mesh.GetName(), mesh.GetNamespace())
	}
	MultipleParentVirtualMeshes = func(virtualMeshes []*smh_networking.VirtualMesh) error {
		var virtualMeshNames []string
		for _, vm := range virtualMeshes {
			virtualMeshNames = append(virtualMeshNames, fmt.Sprintf("%s.%s", vm.GetName(), vm.GetNamespace()))
		}
		return eris.Errorf("Services belong to multiple different parent VirtualMeshes: [%s]", strings.Join(virtualMeshNames, ", "))
	}
	MissingOutlierDetection = func(meshService *smh_discovery.MeshService) error {
		return eris.Errorf("Service %s.%s.%s does not have any TrafficPolicies that apply OutlierDetection settings.",
			meshService.Spec.GetKubeService().GetRef().GetName(),
			meshService.Spec.GetKubeService().GetRef().GetNamespace(),
			meshService.Spec.GetKubeService().GetRef().GetCluster())
	}
)

type failoverServiceValidator struct{}

func NewFailoverServiceValidator() FailoverServiceValidator {
	return &failoverServiceValidator{}
}

func (f *failoverServiceValidator) Validate(inputSnapshot failover.InputSnapshot) {
	for _, failoverService := range inputSnapshot.FailoverServices.List() {
		failoverService.Status.ValidationStatus = f.validateSingle(failoverService, inputSnapshot)
		failoverService.Status.ObservedGeneration = failoverService.GetGeneration()
	}
}

func (f *failoverServiceValidator) validateSingle(
	failoverService *smh_networking.FailoverService,
	inputSnapshot failover.InputSnapshot,
) *smh_core_types.Status {
	var multierr *multierror.Error
	if err := f.validateHostname(failoverService); err != nil {
		multierr = multierror.Append(multierr, err)
	}
	if err := f.validatePort(failoverService); err != nil {
		multierr = multierror.Append(multierr, err)
	}
	if err := f.validateServices(failoverService, inputSnapshot.MeshServices.List(), inputSnapshot.Meshes); err != nil {
		multierr = multierror.Append(multierr, err)
	}
	if err := f.validateFederation(failoverService, inputSnapshot.MeshServices.List(), inputSnapshot.Meshes, inputSnapshot.VirtualMeshes); err != nil {
		multierr = multierror.Append(multierr, err)
	}
	if err := multierr.ErrorOrNil(); err != nil {
		return &smh_core_types.Status{
			State:   smh_core_types.Status_INVALID,
			Message: err.Error(),
		}
	} else {
		return &smh_core_types.Status{
			State: smh_core_types.Status_ACCEPTED,
		}
	}
}

func (f *failoverServiceValidator) validateServices(
	failoverService *smh_networking.FailoverService,
	allMeshServices []*smh_discovery.MeshService,
	meshes v1alpha1sets2.MeshSet,
) error {
	var multierr *multierror.Error
	// validate failover services
	failoverServices := failoverService.Spec.GetFailoverServices()
	if len(failoverServices) == 0 {
		return MissingServices
	}
	for _, serviceRef := range failoverServices {
		meshService, err := f.findMeshService(serviceRef, allMeshServices)
		if err != nil {
			// Corresponding MeshService not found.
			multierr = multierror.Append(multierr, FailoverServiceNotFound(serviceRef))
			continue
		}
		if err := f.validateServiceOutlierDetection(meshService); err != nil {
			multierr = multierror.Append(multierr, err)
		}

		meshRef := meshService.Spec.GetMesh()
		// Validate that mesh exists
		parentMesh, err := meshes.Find(failover.ResourceId{meshService.Spec.GetMesh()})
		if err != nil {
			multierr = multierror.Append(multierr, MeshNotFound(meshRef, meshService.Spec.GetKubeService().GetRef()))
			continue
		}
		if err := f.validateMesh(parentMesh); err != nil {
			multierr = multierror.Append(multierr, err)
		}
	}

	// TODO validateFederation(meshes, virtualMeshes, failoverService)
	//if err := f.validateFederation(serviceParentMeshPairs, virtualMeshes); err != nil {
	//	multierr = multierror.Append(multierr, err)
	//}
	return multierr.ErrorOrNil()
}

func (f *failoverServiceValidator) findMeshService(
	serviceRef *v1.ClusterObjectRef,
	allMeshServices []*smh_discovery.MeshService,
) (*smh_discovery.MeshService, error) {
	for _, meshService := range allMeshServices {
		kubeService := meshService.Spec.GetKubeService().GetRef()
		if serviceRef.GetName() == kubeService.GetName() &&
			serviceRef.GetNamespace() == kubeService.GetNamespace() &&
			serviceRef.GetClusterName() == kubeService.GetCluster() {
			return meshService, nil
		}
	}
	return nil, FailoverServiceNotFound(serviceRef)
}

func (f *failoverServiceValidator) validateServiceOutlierDetection(meshService *smh_discovery.MeshService) error {
	for _, tp := range meshService.Status.GetValidatedTrafficPolicies() {
		if tp.GetTrafficPolicySpec().GetOutlierDetection() != nil {
			return nil
		}
	}
	return MissingOutlierDetection(meshService)
}

func (f *failoverServiceValidator) validateMesh(
	mesh *smh_discovery.Mesh,
) error {
	// Validate that mesh type is supported
	switch meshType := mesh.Spec.GetMeshType().(type) {
	case *types.MeshSpec_Istio1_5_:
	case *types.MeshSpec_Istio1_6_:
	default:
		return UnsupportedMeshType(meshType)
	}
	return nil
}

// TODO(harveyxia) Federation should update Mesh status with VirtualMesh ref
// Valid only if FailoverService is composed of meshes and/or services belonging to
// a common mesh, or to meshes grouped under a common VirtualMesh.
func (f *failoverServiceValidator) validateFederation(
	failoverService *smh_networking.FailoverService,
	allMeshServices []*smh_discovery.MeshService,
	allMeshes v1alpha1sets2.MeshSet,
	allVirtualMeshes v1alpha1sets.VirtualMeshSet,
) error {
	// Surface these errors only if the FailoverService references multiple meshes.
	var missingParentVMErrors []error
	var multierr *multierror.Error
	referencedMeshes := v1alpha1sets2.NewMeshSet()
	referencedVMs := v1alpha1sets.NewVirtualMeshSet()
	if failoverService.Spec.GetMeshes() == nil {
		return MissingMeshes
	}
	// Process declared meshes
	for _, meshRef := range failoverService.Spec.GetMeshes() {
		mesh, err := allMeshes.Find(&v1.ClusterObjectRef{
			Name:      meshRef.GetName(),
			Namespace: meshRef.GetNamespace(),
		})
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		referencedMeshes.Insert(mesh)
	}
	// Process declared services
	for _, serviceRef := range failoverService.Spec.GetFailoverServices() {
		meshService, err := f.findMeshService(serviceRef, allMeshServices)
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		// TODO change type of MeshService.Spec.Mesh to ClusterObjectRef
		mesh, err := allMeshes.Find(&v1.ClusterObjectRef{
			Name:        meshService.Spec.GetMesh().GetName(),
			Namespace:   meshService.Spec.GetMesh().GetNamespace(),
			ClusterName: meshService.Spec.GetMesh().GetCluster(),
		})
		if err != nil {
			multierr = multierror.Append(multierr, err)
			continue
		}
		referencedMeshes.Insert(mesh)
	}
	// Compute referenced VirtualMeshes
	for _, mesh := range referencedMeshes.List() {
		vm := f.findVirtualMeshForMesh(mesh, allVirtualMeshes)
		if vm == nil {
			missingParentVMErrors = append(missingParentVMErrors, MeshWithoutParentVM(mesh))
		} else {
			referencedVMs.Insert(vm)
		}
	}
	// Validate that there's only one common parent mesh, else that there's only a single common parent VirtualMesh
	if len(referencedMeshes.List()) > 1 {
		// Surface meshes with parent meshes as errors
		for _, err := range missingParentVMErrors {
			multierr = multierror.Append(multierr, err)
		}
		if len(referencedVMs.List()) > 1 {
			multierr = multierror.Append(multierr, MultipleParentVirtualMeshes(referencedVMs.List()))
		}
	}
	return multierr.ErrorOrNil()
}

func (f *failoverServiceValidator) validateHostname(failoverService *smh_networking.FailoverService) error {
	hostname := failoverService.Spec.GetHostname()
	if hostname == "" {
		return MissingHostname
	}
	errStrings := validation.IsDNS1123Subdomain(hostname)
	if len(errStrings) > 0 {
		return eris.New(strings.Join(errStrings, ", "))
	}
	return nil
}

func (f *failoverServiceValidator) validatePort(failoverService *smh_networking.FailoverService) error {
	port := failoverService.Spec.GetPort()
	if port == nil {
		return MissingPort
	}
	var multierr *multierror.Error
	if errStrings := validation.IsValidPortNum(int(port.GetPort())); errStrings != nil {
		multierr = multierror.Append(multierr, eris.New(strings.Join(errStrings, ", ")))
	}
	if protocol.Parse(port.GetProtocol()) == protocol.Unsupported {
		multierr = multierror.Append(multierr, eris.Errorf("Invalid protocol for port: %s", port.GetProtocol()))
	}
	return multierr.ErrorOrNil()
}

func (f *failoverServiceValidator) findVirtualMeshForMesh(
	mesh *smh_discovery.Mesh,
	allVirtualMeshes v1alpha1sets.VirtualMeshSet,
) *smh_networking.VirtualMesh {
	virtualMeshes := allVirtualMeshes.List()
	for _, vm := range virtualMeshes {
		for _, meshRef := range vm.Spec.GetMeshes() {
			if mesh.GetName() == meshRef.GetName() && mesh.GetNamespace() == meshRef.GetNamespace() {
				return vm
			}
		}
	}
	return nil
}
