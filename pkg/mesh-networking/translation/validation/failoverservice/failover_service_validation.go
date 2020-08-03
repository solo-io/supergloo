package failoverservice

import (
	"fmt"
	"strings"

	"github.com/rotisserie/eris"
	discoveryv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2"
	discoveryv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha2/sets"
	networkingv1alpha2 "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2"
	networkingv1alpha2sets "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha2/sets"
	skv2core "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	v1alpha1sets "github.com/solo-io/skv2/pkg/api/multicluster.solo.io/v1alpha1/sets"
	"github.com/solo-io/skv2/pkg/ezkube"
	"istio.io/istio/pkg/config/protocol"
	"k8s.io/apimachinery/pkg/util/validation"
)

//go:generate mockgen -source ./failover_service_validation.go -destination ./mocks/mock_failover_service_validation.go -package mock_validation

/*
A valid FailoverService must satisfy the following constraints:

1. TargetService must exist
2. Must consist of at least 1 failover service.
3. All declared failover services must exist.
4. All declared failover services must be owned by a supported Mesh type (currently only Istio).
5. All declared failover services must exist in the same VirtualMesh, or belong to a common parent Mesh.
6. All declared failover services must have OutlierDetection settings declared in a TP (grab this from the MeshService status).
7. All targeted Meshes must be of a supported type.
*/
type FailoverServiceValidator interface {
	// Set the validation status for FailoverServices in the Inputs
	Validate(inputs Inputs, failoverService *networkingv1alpha2.FailoverServiceSpec) []error
}

type Inputs struct {
	MeshServices discoveryv1alpha2sets.MeshServiceSet
	// For validation
	KubeClusters  v1alpha1sets.KubernetesClusterSet
	Meshes        discoveryv1alpha2sets.MeshSet
	VirtualMeshes networkingv1alpha2sets.VirtualMeshSet
}

var (
	MissingHostname        = eris.New("Missing required field \"hostname\".")
	MissingPort            = eris.New("Missing required field \"port\".")
	MissingMeshes          = eris.New("Missing required field \"meshes\".")
	MissingServices        = eris.New("There must be at least one service declared for the FailoverService.")
	BackingServiceNotFound = func(serviceRef *skv2core.ClusterObjectRef) error {
		return eris.Errorf("Backing service %s.%s.%s not found in SMH discovery resources.",
			serviceRef.GetName(),
			serviceRef.GetNamespace(),
			serviceRef.GetClusterName())
	}
	MeshNotFound = func(meshRef *skv2core.ObjectRef, serviceRef *skv2core.ClusterObjectRef) error {
		return eris.Errorf("Mesh %s.%s for service %s.%s.%s not found in SMH discovery resources.",
			meshRef.GetName(),
			meshRef.GetNamespace(),
			serviceRef.GetName(),
			serviceRef.GetNamespace(),
			serviceRef.GetClusterName())
	}
	UnsupportedMeshType = func(meshType interface{}, mesh *discoveryv1alpha2.Mesh) error {
		return eris.Errorf("Unsupported Mesh type %T for Mesh %s.%s", meshType, mesh.GetName(), mesh.GetNamespace())
	}
	UnsupportedServiceType = func(serviceType interface{}) error {
		return eris.Errorf("Unsupported service type %T", serviceType)
	}
	MeshWithoutParentVM = func(mesh *discoveryv1alpha2.Mesh) error {
		return eris.Errorf("Mesh %s.%s is not grouped in a VirtualMesh.", mesh.GetName(), mesh.GetNamespace())
	}
	MultipleParentVirtualMeshes = func(virtualMeshes []*networkingv1alpha2.VirtualMesh) error {
		var virtualMeshNames []string
		for _, vm := range virtualMeshes {
			virtualMeshNames = append(virtualMeshNames, fmt.Sprintf("%s.%s", vm.GetName(), vm.GetNamespace()))
		}
		return eris.Errorf("Services belong to multiple different parent VirtualMeshes: [%s]", strings.Join(virtualMeshNames, ", "))
	}
	MissingOutlierDetection = func(meshService *discoveryv1alpha2.MeshService) error {
		return eris.Errorf("Service %s.%s.%s does not have any TrafficPolicies that apply OutlierDetection settings.",
			meshService.Spec.GetKubeService().GetRef().GetName(),
			meshService.Spec.GetKubeService().GetRef().GetNamespace(),
			meshService.Spec.GetKubeService().GetRef().GetClusterName())
	}
)

type failoverServiceValidator struct {
}

func NewFailoverServiceValidator() FailoverServiceValidator {
	return &failoverServiceValidator{}
}

func (f *failoverServiceValidator) Validate(inputs Inputs, failoverService *networkingv1alpha2.FailoverServiceSpec) []error {
	var errs []error
	if err := f.validateHostname(failoverService); err != nil {
		errs = append(errs, err)
	}
	if portErrs := f.validatePort(failoverService); portErrs != nil {
		errs = append(errs, portErrs...)
	}
	if serviceErrs := f.validateServices(failoverService, inputs.MeshServices.List(), inputs.Meshes); serviceErrs != nil {
		errs = append(errs, serviceErrs...)
	}
	if federationErrs := f.validateFederation(failoverService, inputs.MeshServices.List(), inputs.Meshes, inputs.VirtualMeshes); federationErrs != nil {
		errs = append(errs, federationErrs...)
	}
	if meshErrs := f.validateMeshes(failoverService, inputs.Meshes); meshErrs != nil {
		errs = append(errs, meshErrs...)
	}
	return errs
}

func (f *failoverServiceValidator) validateMeshes(
	failoverService *networkingv1alpha2.FailoverServiceSpec,
	meshes discoveryv1alpha2sets.MeshSet,
) []error {
	var errs []error
	for _, meshRef := range failoverService.Meshes {
		mesh, err := meshes.Find(meshRef)
		if err != nil {
			errs = append(errs, err)
		}
		if err := f.validateMesh(mesh); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (f *failoverServiceValidator) validateServices(
	failoverService *networkingv1alpha2.FailoverServiceSpec,
	allMeshServices []*discoveryv1alpha2.MeshService,
	meshes discoveryv1alpha2sets.MeshSet,
) []error {
	var errs []error
	componentServices := failoverService.GetBackingServices()
	if len(componentServices) == 0 {
		return []error{MissingServices}
	}
	for _, typedServiceRef := range componentServices {
		if typedServiceRef.GetKubeService() == nil {
			errs = append(errs, UnsupportedServiceType(typedServiceRef.GetBackingServiceType()))
			continue
		}
		serviceRef := typedServiceRef.GetKubeService()
		meshService, err := f.findMeshService(serviceRef, allMeshServices)
		if err != nil {
			// Corresponding MeshService not found.
			errs = append(errs, BackingServiceNotFound(serviceRef))
			continue
		}
		if err := f.validateServiceOutlierDetection(meshService); err != nil {
			errs = append(errs, err)
		}
		meshRef := meshService.Spec.GetMesh()
		// Approve that mesh exists
		parentMesh, err := meshes.Find(meshService.Spec.GetMesh())
		if err != nil {
			errs = append(errs, MeshNotFound(meshRef, meshService.Spec.GetKubeService().GetRef()))
			continue
		}
		if err := f.validateMesh(parentMesh); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (f *failoverServiceValidator) findMeshService(
	serviceRef *skv2core.ClusterObjectRef,
	allMeshServices []*discoveryv1alpha2.MeshService,
) (*discoveryv1alpha2.MeshService, error) {
	for _, meshService := range allMeshServices {
		if ezkube.ClusterRefsMatch(serviceRef, meshService.Spec.GetKubeService().GetRef()) {
			return meshService, nil
		}
	}
	return nil, BackingServiceNotFound(serviceRef)
}

func (f *failoverServiceValidator) validateServiceOutlierDetection(meshService *discoveryv1alpha2.MeshService) error {
	for _, tp := range meshService.Status.GetAppliedTrafficPolicies() {
		if tp.GetSpec().GetOutlierDetection() != nil {
			return nil
		}
	}
	return MissingOutlierDetection(meshService)
}

func (f *failoverServiceValidator) validateMesh(
	mesh *discoveryv1alpha2.Mesh,
) error {
	switch meshType := mesh.Spec.GetMeshType().(type) {
	case *discoveryv1alpha2.MeshSpec_Istio_:
	default:
		return UnsupportedMeshType(meshType, mesh)
	}
	return nil
}

// Valid only if FailoverService is composed of meshes and/or services belonging to
// a common mesh, or to meshes grouped under a common VirtualMesh.
func (f *failoverServiceValidator) validateFederation(
	failoverService *networkingv1alpha2.FailoverServiceSpec,
	allMeshServices []*discoveryv1alpha2.MeshService,
	allMeshes discoveryv1alpha2sets.MeshSet,
	allVirtualMeshes networkingv1alpha2sets.VirtualMeshSet,
) []error {
	// Surface these errors only if the FailoverService references multiple meshes.
	var missingParentVMErrors []error
	var errs []error
	referencedMeshes := discoveryv1alpha2sets.NewMeshSet()
	referencedVMs := networkingv1alpha2sets.NewVirtualMeshSet()
	if failoverService.GetMeshes() == nil {
		return []error{MissingMeshes}
	}
	// Process declared meshes
	for _, meshRef := range failoverService.GetMeshes() {
		mesh, err := allMeshes.Find(meshRef)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		referencedMeshes.Insert(mesh)
	}
	// Process declared services
	for _, typedServiceRef := range failoverService.GetBackingServices() {
		if typedServiceRef.GetKubeService() == nil {
			// Error already reported when validating component services.
			continue
		}
		serviceRef := typedServiceRef.GetKubeService()
		meshService, err := f.findMeshService(serviceRef, allMeshServices)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		mesh, err := allMeshes.Find(meshService.Spec.Mesh)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		referencedMeshes.Insert(mesh)
	}
	// Compute referenced VirtualMeshes
	for _, mesh := range referencedMeshes.List() {
		if len(mesh.Status.AppliedVirtualMeshes) < 1 {
			missingParentVMErrors = append(missingParentVMErrors, MeshWithoutParentVM(mesh))
		}
		for _, appliedVM := range mesh.Status.AppliedVirtualMeshes {
			vm, err := allVirtualMeshes.Find(appliedVM.Ref)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			referencedVMs.Insert(vm)
		}
	}
	// Approve that there's only one common parent mesh, else that there's only a single common parent VirtualMesh
	if len(referencedMeshes.List()) > 1 {
		// Surface meshes without parent meshes as errors
		for _, err := range missingParentVMErrors {
			errs = append(errs, err)
		}
		if len(referencedVMs.List()) > 1 {
			errs = append(errs, MultipleParentVirtualMeshes(referencedVMs.List()))
		}
	}
	return errs
}

func (f *failoverServiceValidator) validateHostname(failoverService *networkingv1alpha2.FailoverServiceSpec) error {
	hostname := failoverService.GetHostname()
	if hostname == "" {
		return MissingHostname
	}
	errStrings := validation.IsDNS1123Subdomain(hostname)
	if len(errStrings) > 0 {
		return eris.New(strings.Join(errStrings, ", "))
	}
	return nil
}

func (f *failoverServiceValidator) validatePort(failoverService *networkingv1alpha2.FailoverServiceSpec) []error {
	var errs []error
	port := failoverService.GetPort()
	if port == nil {
		return []error{MissingPort}
	}
	if errStrings := validation.IsValidPortNum(int(port.GetNumber())); errStrings != nil {
		errs = append(errs, eris.New(strings.Join(errStrings, ", ")))
	}
	if protocol.Parse(port.GetProtocol()) == protocol.Unsupported {
		errs = append(errs, eris.Errorf("Invalid protocol for port: %s", port.GetProtocol()))
	}
	return errs
}

func (f *failoverServiceValidator) findVirtualMeshForMesh(
	mesh *discoveryv1alpha2.Mesh,
	allVirtualMeshes networkingv1alpha2sets.VirtualMeshSet,
) *networkingv1alpha2.VirtualMesh {
	virtualMeshes := allVirtualMeshes.List()
	for _, vm := range virtualMeshes {
		for _, meshRef := range vm.Spec.GetMeshes() {
			// A Mesh can be grouped into at most one VirtualMesh.
			if ezkube.RefsMatch(mesh, meshRef) {
				return vm
			}
		}
	}
	return nil
}
