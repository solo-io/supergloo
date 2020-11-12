package checks

import (
	"context"
	"fmt"
	"strings"

	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/skv2/pkg/ezkube"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type networkingCrdCheck struct{}

func NewNetworkingCrdCheck() Check {
	return &networkingCrdCheck{}
}

func (c *networkingCrdCheck) GetDescription() string {
	return "Gloo Mesh networking configuration resources are in a valid state"
}

func (c *networkingCrdCheck) Run(ctx context.Context, client client.Client, _ string) *Failure {
	var allErrors []error
	tpList, err := v1alpha2.NewTrafficPolicyClient(client).ListTrafficPolicy(ctx)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	apList, err := v1alpha2.NewAccessPolicyClient(client).ListAccessPolicy(ctx)
	if err != nil {
		allErrors = append(allErrors, err)
	}
	fsList, err := v1alpha2.NewFailoverServiceClient(client).ListFailoverService(ctx)
	if err != nil {
		allErrors = append(allErrors, err)
	}

	if errs := c.checkTrafficPolicies(tpList); errs != nil {
		allErrors = append(allErrors, errs...)
	}
	if errs := c.checkAccessPolicies(apList); errs != nil {
		allErrors = append(allErrors, errs...)
	}
	if errs := c.checkFailoverServices(fsList); errs != nil {
		allErrors = append(allErrors, errs...)
	}

	if len(allErrors) > 0 {
		return &Failure{
			Errors: allErrors,
			Hint:   c.buildHint(),
		}
	}
	return nil
}

func (c *networkingCrdCheck) checkTrafficPolicies(tpList *v1alpha2.TrafficPolicyList) []error {
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

func (c *networkingCrdCheck) checkAccessPolicies(apList *v1alpha2.AccessPolicyList) []error {
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

func (c *networkingCrdCheck) checkFailoverServices(fsList *v1alpha2.FailoverServiceList) []error {
	if fsList == nil {
		return nil
	}
	var errs []error
	for _, fs := range fsList.Items {
		fs := fs // pike
		if err := c.checkStatus(fs.Kind, &fs, fs.Generation, &fs.Status); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// This struct captures metadata common to all networking CRD statuses
type networkingCrdStatus interface {
	GetObservedGeneration() int64
	GetState() v1alpha2.ApprovalState
}

func (c *networkingCrdCheck) checkStatus(kind string, id ezkube.ResourceId, generation int64, status networkingCrdStatus) error {
	var errorStrings []string
	resourceKey := fmt.Sprintf("%s %s.%s", kind, id.GetName(), id.GetNamespace())
	if status.GetState() != v1alpha2.ApprovalState_ACCEPTED {
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
