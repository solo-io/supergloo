package reconciliation

import (
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	"github.com/solo-io/skv2/contrib/pkg/output"
	"github.com/solo-io/skv2/pkg/ezkube"
)

var _ output.ErrorHandler = errHandler{}

type errHandler struct {
	inp input.Snapshot
}

func (e errHandler) HandleWriteError(resource ezkube.Object, err error) {
	// annotations := resource.GetAnnotations()
	//
	// switch gvk {
	// case "failoverservice.networking.mesh.gloo.solo.io":
	// 	e.inp.FailoverServices()
	// }

	// switch r := resource.(type) {
	// case *v1alpha2.VirtualMesh:
	// 	r.Status.Errors = append(r.Status.Errors, err.Error())
	// case *v1alpha2.FailoverService:
	// 	r.Status.Errors = append(r.Status.Errors, err.Error())
	// case *v1alpha2.TrafficPolicy:
	// 	r.Status.Errors = append(r.Status.Errors, err.Error())
	// case *v1alpha2.AccessPolicy:
	// 	r.Status.Errors = append(r.Status.Errors, err.Error())
	// }
}

func (e errHandler) HandleDeleteError(resource ezkube.Object, err error) {
	panic("unimplemented")
}

func (e errHandler) HandleListError(err error) {
	panic("unimplemented")
}

func (e errHandler) Errors() error {
	return nil
}
