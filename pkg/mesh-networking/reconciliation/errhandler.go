package reconciliation

import (
	"context"

	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/contrib/pkg/output/errhandlers"
	"github.com/solo-io/skv2/pkg/controllerutils"
	"github.com/solo-io/skv2/pkg/ezkube"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ output.ErrorHandler = errHandler{}

type errHandler struct {
	client client.Client
	errhandlers.AppendingErrHandler
}

func (e errHandler) HandleWriteError(resource ezkube.Object, err error) {
	e.AppendingErrHandler.HandleWriteError(resource, err)

	switch r := resource.(type) {
	case *v1alpha2.VirtualMesh:
		r.Status.Errors = append(r.Status.Errors, e.AppendingErrHandler.Errors().Error())
	case *v1alpha2.FailoverService:
		r.Status.Errors = append(r.Status.Errors, e.AppendingErrHandler.Errors().Error())
	case *v1alpha2.TrafficPolicy:
		r.Status.Errors = append(r.Status.Errors, e.AppendingErrHandler.Errors().Error())
	case *v1alpha2.AccessPolicy:
		r.Status.Errors = append(r.Status.Errors, e.AppendingErrHandler.Errors().Error())
	}

	if _, err := controllerutils.UpdateStatus(context.TODO(), e.client, resource); err != nil {
		e.AppendingErrHandler.HandleWriteError(resource, err)
	}
}
