package checks

import (
	"context"
	"fmt"
	"strings"

	"github.com/rotisserie/eris"
	commonv1 "github.com/solo-io/gloo-mesh/pkg/api/common.mesh.gloo.solo.io/v1"
	networkingv1 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

type networkingCrdCheck struct{}

func NewNetworkingCrdCheck() Check {
	return &networkingCrdCheck{}
}

func (c *networkingCrdCheck) GetDescription() string {
	return "Gloo Mesh networking configuration resources are in a valid state"
}

func (c *networkingCrdCheck) Run(ctx context.Context, checkCtx CheckContext) *Failure {
	client := checkCtx.Client()
	failure := new(Failure)
	tpList, err := networkingv1.NewTrafficPolicyClient(client).ListTrafficPolicy(ctx)
	if err != nil {
		failure.AddError(err)
	}
	apList, err := networkingv1.NewAccessPolicyClient(client).ListAccessPolicy(ctx)
	if err != nil {
		failure.AddError(err)
	}
	vmList, err := networkingv1.NewVirtualMeshClient(client).ListVirtualMesh(ctx)
	if err != nil {
		failure.AddError(err)
	}

	if errs := c.checkTrafficPolicies(tpList); errs != nil {
		failure.AddError(errs...)
	}
	if errs := c.checkAccessPolicies(apList); errs != nil {
		failure.AddError(errs...)
	}
	if errs := c.checkVirtualMeshes(vmList); errs != nil {
		failure.AddError(errs...)
	}

	if len(failure.Errors) > 0 {
		failure.AddHint(c.buildHint(), "")
	}
	return failure
}

func (c *networkingCrdCheck) checkTrafficPolicies(tpList *networkingv1.TrafficPolicyList) []error {
	if tpList == nil {
		return nil
	}
	var errs []error
	for _, tp := range tpList.Items {
		tp := tp // pike
		if err := c.checkStatus(tp.Kind, &tp, tp.Generation, &tp.Status); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (c *networkingCrdCheck) checkAccessPolicies(apList *networkingv1.AccessPolicyList) []error {
	if apList == nil {
		return nil
	}
	var errs []error
	for _, ap := range apList.Items {
		ap := ap // pike
		if err := c.checkStatus(ap.Kind, &ap, ap.Generation, &ap.Status); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (c *networkingCrdCheck) checkVirtualMeshes(vmList *networkingv1.VirtualMeshList) []error {
	if vmList == nil {
		return nil
	}
	var errs []error
	for _, vm := range vmList.Items {
		vm := vm // pike
		if err := c.checkStatus(vm.Kind, &vm, vm.Generation, &vm.Status); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// This struct captures metadata common to all networking CRD statuses
type networkingCrdStatus interface {
	GetObservedGeneration() int64
	GetState() commonv1.ApprovalState
}

func (c *networkingCrdCheck) checkStatus(kind string, id ezkube.ResourceId, generation int64, status networkingCrdStatus) error {
	var errorStrings []string
	resourceKey := fmt.Sprintf("%s %s.%s", kind, id.GetName(), id.GetNamespace())
	if status.GetState() != commonv1.ApprovalState_ACCEPTED {
		errorStrings = append(errorStrings, fmt.Sprintf("approval state is %s", status.GetState().String()))
	}
	if status.GetObservedGeneration() != generation {
		errorStrings = append(errorStrings, fmt.Sprintf("observed generation does not match generation"))
	}
	if len(errorStrings) > 0 {
		return eris.Errorf("%s: %s", resourceKey, strings.Join(errorStrings, ", "))
	}
	return nil
}

func (c *networkingCrdCheck) buildHint() string {
	return fmt.Sprintf(`check the status of Gloo Mesh networking resources with "kubectl get <resource-type> -Aoyaml"`)
}
