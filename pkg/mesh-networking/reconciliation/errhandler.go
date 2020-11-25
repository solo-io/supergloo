package reconciliation

import (
	"encoding/json"

	discoveryv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/discovery.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/input"
	networkingv1alpha2 "github.com/solo-io/gloo-mesh/pkg/api/networking.mesh.gloo.solo.io/v1alpha2"
	"github.com/solo-io/gloo-mesh/pkg/mesh-networking/translation/utils/metautils"
	"github.com/solo-io/skv2/contrib/pkg/output"
	v1 "github.com/solo-io/skv2/pkg/api/core.skv2.solo.io/v1"
	"github.com/solo-io/skv2/pkg/ezkube"
)

var _ output.ErrorHandler = errHandler{}

type errHandler struct {
	inp input.Snapshot
}

func (e errHandler) HandleWriteError(resource ezkube.Object, err error) {
	annotations := resource.GetAnnotations()
	parentsStr, ok := annotations[metautils.ParentLabelkey]
	if !ok {
		return
	}
	var allParents map[string][]*v1.ObjectRef
	if err := json.Unmarshal([]byte(parentsStr), &allParents); err != nil {
		return
	}

	for gvk, parents := range allParents {
		switch gvk {
		case discoveryv1alpha2.Mesh{}.GVK().String():
			for _, parentMesh := range parents {
				for _, mesh := range e.inp.Meshes().List(func(mesh *discoveryv1alpha2.Mesh) bool {
					return parentMesh.Equal(mesh)
				}) {
					// TODO(ryantking): Append error to mesh
					_ = mesh
				}
			}
		case discoveryv1alpha2.TrafficTarget{}.GVK().String():
			for _, parentTarget := range parents {
				for _, target := range e.inp.TrafficTargets().List(func(target *discoveryv1alpha2.TrafficTarget) bool {
					return parentTarget.Equal(target)
				}) {
					// TODO(ryantking): Append error to traffic target
					_ = target
				}
			}
		case networkingv1alpha2.VirtualMesh{}.GVK().String():
			for _, parentVMesh := range parents {
				for _, vmesh := range e.inp.VirtualMeshes().List(func(vmesh *networkingv1alpha2.VirtualMesh) bool {
					return parentVMesh.Equal(vmesh)
				}) {
					vmesh.Status.Errors = append(vmesh.Status.Errors, err.Error())
				}
			}
		case networkingv1alpha2.AccessPolicy{}.GVK().String():
			for _, parentAP := range parents {
				for _, ap := range e.inp.AccessPolicies().List(func(ap *networkingv1alpha2.AccessPolicy) bool {
					return parentAP.Equal(ap)
				}) {
					ap.Status.Errors = append(ap.Status.Errors, err.Error())
				}
			}
		case networkingv1alpha2.TrafficPolicy{}.GVK().String():
			for _, parentTP := range parents {
				for _, tp := range e.inp.TrafficPolicies().List(func(tp *networkingv1alpha2.TrafficPolicy) bool {
					return parentTP.Equal(tp)
				}) {
					tp.Status.Errors = append(tp.Status.Errors, err.Error())
				}
			}
		case networkingv1alpha2.FailoverService{}.GVK().String():
			for _, parentFS := range parents {
				for _, fs := range e.inp.FailoverServices().List(func(fs *networkingv1alpha2.FailoverService) bool {
					return parentFS.Equal(fs)
				}) {
					fs.Status.Errors = append(fs.Status.Errors, err.Error())
				}
			}
		}
	}
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
