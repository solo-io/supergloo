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
	"istio.io/istio/pkg/config/protocol"
	k8s_validation "k8s.io/apimachinery/pkg/util/validation"
)

//go:generate mockgen -source ./validation.go -destination ./mocks/mock_validation.go -package mock_failover_service_validation

/*
A valid FailoverService must satisfy the following constraints:

1. Must consist of at least 1 service.
2. All declared Services must exist.
3. All declared Services must be owned by a supported Mesh type (currently only Istio).
4. All declared Services must exist in the same VirtualMesh.
5. All declared Services must have OutlierDetection settings declared in a TP (grab this from the MeshService status).
6. Hostname must be populated.
7. Port must be populated.
8. Cluster must be populated.
9. Namespace must be populated.
*/
type FailoverServiceValidator interface {
	// Set the validation status for FailoverServices in the InputSnapshot
	Validate(snapshot failover.InputSnapshot)
}

var (
	MissingHostname  = eris.New("Missing required field \"hostname\".")
	MissingPort      = eris.New("Missing required field \"port\".")
	MissingCluster   = eris.New("Missing required field \"cluster\".")
	MissingNamespace = eris.New("Missing required field \"namespace\".")
	MissingServices  = eris.New("There must be at least one service declared for the FailoverService.")
	ClusterNotFound  = func(cluster string) error {
		return eris.Errorf("Declared cluster %s not found.", cluster)
	}
	ServiceNotFound = func(serviceRef *smh_core_types.ResourceRef) error {
		return eris.Errorf("Declared service %s.%s.%s not found in SMH discovery resources.",
			serviceRef.GetName(),
			serviceRef.GetNamespace(),
			serviceRef.GetCluster())
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
	ServiceWithoutParentVM = func(serviceRef *smh_core_types.ResourceRef, parentMesh *smh_discovery.Mesh) error {
		return eris.Errorf("Service %s.%s.%s with parent Mesh %s is not contained in a VirtualMesh.",
			serviceRef.GetName(), serviceRef.GetNamespace(), serviceRef.GetCluster(), parentMesh.GetName())
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

type serviceMeshPair struct {
	serviceRef *smh_core_types.ResourceRef
	mesh       *smh_discovery.Mesh
}

type failoverServiceValidator struct{}

func NewFailoverServiceValidator() FailoverServiceValidator {
	return &failoverServiceValidator{}
}

func (f *failoverServiceValidator) Validate(inputSnapshot failover.InputSnapshot) {
	for _, failoverService := range inputSnapshot.FailoverServices {
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
	if err := f.validateCluster(failoverService, inputSnapshot.KubeClusters); err != nil {
		multierr = multierror.Append(multierr, err)
	}
	if err := f.validateNamespace(failoverService); err != nil {
		multierr = multierror.Append(multierr, err)
	}
	if err := f.validateServices(failoverService, inputSnapshot.MeshServices, inputSnapshot.Meshes, inputSnapshot.VirtualMeshes); err != nil {
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
	meshServices []*smh_discovery.MeshService,
	meshes []*smh_discovery.Mesh,
	virtualMeshes []*smh_networking.VirtualMesh,
) error {
	services := failoverService.Spec.GetServices()
	if len(services) == 0 {
		return MissingServices
	}
	var multierr *multierror.Error
	var serviceParentMeshPairs []serviceMeshPair
	for _, serviceRef := range failoverService.Spec.GetServices() {
		meshService, err := f.findMeshService(serviceRef, meshServices)
		if err != nil {
			// Corresponding MeshService not found.
			multierr = multierror.Append(multierr, err)
			continue
		}
		if err := f.validateServiceOutlierDetection(meshService); err != nil {
			multierr = multierror.Append(multierr, err)
		}
		serviceParentMeshPair, err := f.validateParentMesh(meshService, meshes)
		if err != nil {
			multierr = multierror.Append(multierr, err)
		} else {
			serviceParentMeshPairs = append(serviceParentMeshPairs, serviceParentMeshPair)
		}
	}
	if err := f.validateFederation(serviceParentMeshPairs, virtualMeshes); err != nil {
		multierr = multierror.Append(multierr, err)
	}
	return multierr.ErrorOrNil()
}

func (f *failoverServiceValidator) findMeshService(
	serviceRef *smh_core_types.ResourceRef,
	allMeshServices []*smh_discovery.MeshService,
) (*smh_discovery.MeshService, error) {
	for _, meshService := range allMeshServices {
		kubeService := meshService.Spec.GetKubeService().GetRef()
		if serviceRef.GetName() == kubeService.GetName() &&
			serviceRef.GetNamespace() == kubeService.GetNamespace() &&
			serviceRef.GetCluster() == kubeService.GetCluster() {
			return meshService, nil
		}
	}
	return nil, ServiceNotFound(serviceRef)
}

func (f *failoverServiceValidator) validateServiceOutlierDetection(meshService *smh_discovery.MeshService) error {
	for _, tp := range meshService.Status.GetValidatedTrafficPolicies() {
		if tp.GetTrafficPolicySpec().GetOutlierDetection() != nil {
			return nil
		}
	}
	return MissingOutlierDetection(meshService)
}

func (f *failoverServiceValidator) validateParentMesh(
	meshService *smh_discovery.MeshService,
	allMeshes []*smh_discovery.Mesh,
) (serviceMeshPair, error) {
	meshRef := meshService.Spec.GetMesh()
	var parentMesh *smh_discovery.Mesh
	// Validate that mesh exists
	for _, mesh := range allMeshes {
		if meshRef.GetName() == mesh.GetName() &&
			meshRef.GetNamespace() == mesh.GetNamespace() {
			parentMesh = mesh
			break
		}
	}
	if parentMesh == nil {
		return serviceMeshPair{}, MeshNotFound(meshRef, meshService.Spec.GetKubeService().GetRef())
	}
	// Validate that mesh type is supported
	switch meshType := parentMesh.Spec.GetMeshType().(type) {
	case *types.MeshSpec_Istio1_5_:
	case *types.MeshSpec_Istio1_6_:
	default:
		return serviceMeshPair{}, UnsupportedMeshType(meshType)
	}
	return serviceMeshPair{
		serviceRef: meshService.Spec.GetKubeService().GetRef(),
		mesh:       parentMesh,
	}, nil
}

// TODO(harveyxia) Federation should update Mesh status with VirtualMesh ref
// Return error if services are in separate and non-federated meshes
func (f *failoverServiceValidator) validateFederation(
	serviceParentMeshPairs []serviceMeshPair,
	virtualMeshes []*smh_networking.VirtualMesh,
) error {
	var missingParentVMErrors []error
	var multierr *multierror.Error
	parentMeshes := v1alpha1sets2.NewMeshSet()
	parentVMs := v1alpha1sets.NewVirtualMeshSet()
	// Fetch all parent VirtualMeshes
	for _, serviceParentMeshPair := range serviceParentMeshPairs {
		parentMesh := serviceParentMeshPair.mesh
		parentMeshes.Insert(parentMesh)
		var parentVM *smh_networking.VirtualMesh
		for _, vm := range virtualMeshes {
			for _, meshRef := range vm.Spec.GetMeshes() {
				if parentMesh.GetName() == meshRef.GetName() && parentMesh.GetNamespace() == meshRef.GetNamespace() {
					parentVM = vm
					break
				}
			}
			if parentVM != nil {
				break
			}
		}
		if parentVM == nil {
			missingParentVMErrors = append(missingParentVMErrors, ServiceWithoutParentVM(serviceParentMeshPair.serviceRef, serviceParentMeshPair.mesh))
		} else {
			parentVMs.Insert(parentVM)
		}
	}
	// Validate that there's only one common parent mesh, else that there's only a single common parent VirtualMesh
	if len(parentMeshes.List()) > 1 {
		// Surface meshes with parent meshes as errors
		for _, err := range missingParentVMErrors {
			multierr = multierror.Append(multierr, err)
		}
		if len(parentVMs.List()) > 1 {
			multierr = multierror.Append(multierr, MultipleParentVirtualMeshes(parentVMs.List()))
		}
	}
	return multierr.ErrorOrNil()
}

func (f *failoverServiceValidator) validateCluster(
	failoverService *smh_networking.FailoverService,
	kubeClusters []*smh_discovery.KubernetesCluster,
) error {
	cluster := failoverService.Spec.GetCluster()
	if cluster == "" {
		return MissingCluster
	}
	for _, kubeCluster := range kubeClusters {
		if cluster == kubeCluster.GetName() {
			return nil
		}
	}
	return ClusterNotFound(cluster)
}

func (f *failoverServiceValidator) validateHostname(failoverService *smh_networking.FailoverService) error {
	hostname := failoverService.Spec.GetHostname()
	if hostname == "" {
		return MissingHostname
	}
	errStrings := k8s_validation.IsDNS1123Subdomain(hostname)
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
	if errStrings := k8s_validation.IsValidPortNum(int(port.GetPort())); errStrings != nil {
		multierr = multierror.Append(multierr, eris.New(strings.Join(errStrings, ", ")))
	}
	if errStrings := k8s_validation.IsValidPortName(port.GetName()); errStrings != nil {
		multierr = multierror.Append(multierr, eris.New(strings.Join(errStrings, ", ")))
	}
	if protocol.Parse(port.GetProtocol()) == protocol.Unsupported {
		multierr = multierror.Append(multierr, eris.Errorf("Invalid protocol for port: %s", port.GetProtocol()))
	}
	return multierr.ErrorOrNil()
}

func (f *failoverServiceValidator) validateNamespace(failoverService *smh_networking.FailoverService) error {
	if failoverService.Spec.GetNamespace() == "" {
		return MissingNamespace
	}
	return nil
}
