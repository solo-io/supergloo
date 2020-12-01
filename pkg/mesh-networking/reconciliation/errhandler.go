package reconciliation

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/go-multierror"
	"github.com/rotisserie/eris"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
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
				if vmesh, findErr := e.in.VirtualMeshes().Find(parentVMesh); findErr == nil {
					vmesh.Status.Errors = append(vmesh.Status.Errors, err.Error())
				}
			}
		case v1alpha2.AccessPolicy{}.GVK().String():
			for _, parentAP := range parents {
				if ap, findErr := e.in.VirtualMeshes().Find(parentAP); findErr == nil {
					ap.Status.Errors = append(ap.Status.Errors, err.Error())
				}
			}
		case v1alpha2.TrafficPolicy{}.GVK().String():
			for _, parentTP := range parents {
				if tp, findErr := e.in.VirtualMeshes().Find(parentTP); findErr == nil {
					tp.Status.Errors = append(tp.Status.Errors, err.Error())
				}
			}
		case v1alpha2.FailoverService{}.GVK().String():
			for _, parentFS := range parents {
				if fs, findErr := e.in.VirtualMeshes().Find(parentFS); findErr == nil {
					fs.Status.Errors = append(fs.Status.Errors, err.Error())
				}
			}
		}
	}
}

func (e *errHandler) Errors() error {
	return e.errs.ErrorOrNil()
}
