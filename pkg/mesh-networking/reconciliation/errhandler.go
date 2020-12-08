package reconciliation

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	input "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input/networking"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/skv2/contrib/pkg/output"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

var _ output.ErrorHandler = &errHandler{}

type errHandler struct {
	ctx  context.Context
	in   input.Snapshot
	errs *multierror.Error
}

func newErrHandler(ctx context.Context, inp input.Snapshot) *errHandler {
	return &errHandler{ctx, inp, &multierror.Error{}}
}

func (e *errHandler) HandleWriteError(resource ezkube.Object, err error) {
	e.handleError(resource, eris.Wrap(err, "write error"))
}

func (e *errHandler) HandleDeleteError(resource ezkube.Object, err error) {
	e.handleError(resource, eris.Wrap(err, "delete error"))
}

func (e *errHandler) HandleListError(err error) {
	e.errs = multierror.Append(e.errs, err)
}

func (e *errHandler) handleError(resource ezkube.Object, err error) {
	e.errs = multierror.Append(e.errs, err)

	annotations := resource.GetAnnotations()
	parentsStr, ok := annotations[metautils.ParentLabelkey]
	if !ok {
		return
	}
	var allParents map[string][]*v1.ObjectRef
	if err := json.Unmarshal([]byte(parentsStr), &allParents); err != nil {
		contextutils.LoggerFrom(e.ctx).Errorf("internal error: could not unmarshal %q annotation", metautils.ParentLabelkey)
		return
	}

	for gvk, parents := range allParents {
		switch gvk {
		case v1alpha2.VirtualMesh{}.GVK().String():
			for _, parentVMesh := range parents {
				vmesh, findErr := e.in.VirtualMeshes().Find(parentVMesh)
				if findErr != nil {
					contextutils.LoggerFrom(e.ctx).Errorf("internal error: resource for parent not found: %s", parentVMesh.String())
					continue
				}

				vmesh.Status.Errors = append(vmesh.Status.Errors, err.Error())
				vmesh.Status.State = v1alpha2.ApprovalState_FAILED
			}
		case v1alpha2.AccessPolicy{}.GVK().String():
			for _, parentAP := range parents {
				ap, findErr := e.in.VirtualMeshes().Find(parentAP)
				if findErr != nil {
					contextutils.LoggerFrom(e.ctx).Errorf("internal error: resource for parent not found: %s", parentAP.String())
					continue
				}

				ap.Status.Errors = append(ap.Status.Errors, err.Error())
				ap.Status.State = v1alpha2.ApprovalState_FAILED
			}
		case v1alpha2.TrafficPolicy{}.GVK().String():
			for _, parentTP := range parents {
				tp, findErr := e.in.VirtualMeshes().Find(parentTP)
				if findErr != nil {
					contextutils.LoggerFrom(e.ctx).Errorf("internal error: resource for parent not found: %s", parentTP.String())
					continue
				}

				tp.Status.Errors = append(tp.Status.Errors, err.Error())
				tp.Status.State = v1alpha2.ApprovalState_FAILED
			}
		case v1alpha2.FailoverService{}.GVK().String():
			for _, parentFS := range parents {
				fs, findErr := e.in.VirtualMeshes().Find(parentFS)
				if findErr != nil {
					contextutils.LoggerFrom(e.ctx).Errorf("internal error: resource for parent not found: %s", parentFS.String())
					continue
				}

				fs.Status.Errors = append(fs.Status.Errors, err.Error())
				fs.Status.State = v1alpha2.ApprovalState_FAILED
			}
		}
	}
}

func (e *errHandler) Errors() error {
	return e.errs.ErrorOrNil()
}
