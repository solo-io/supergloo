package mesh_service

import (
	"context"

	"github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	corev1 "k8s.io/api/core/v1"
)

/**********
* kind-specific handlers that just forward events on to their callback
* we're not attaching these methods to the parent object because function overloading
* is apparently too hard for Go >:(
**********/

type ServiceEventHandler struct {
	Ctx                 context.Context
	HandleServiceUpsert func(service *corev1.Service) error
}

func (s *ServiceEventHandler) CreateService(obj *corev1.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, obj).Debugf("Handling event")
	return s.HandleServiceUpsert(obj)
}

func (s *ServiceEventHandler) UpdateService(old, new *corev1.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, new).Debugf("Handling event")
	return s.HandleServiceUpsert(new)
}

func (s *ServiceEventHandler) DeleteService(obj *corev1.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, obj).Warnf("Ignoring event")
	return nil
}

func (s *ServiceEventHandler) GenericService(obj *corev1.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, obj).Errorf("Ignoring event")
	return nil
}

type MeshWorkloadEventHandler struct {
	Ctx                      context.Context
	HandleMeshWorkloadUpsert func(meshWorkload *v1alpha1.MeshWorkload) error
}

func (m *MeshWorkloadEventHandler) CreateMeshWorkload(obj *v1alpha1.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, obj).Debugf("Handling event")
	return m.HandleMeshWorkloadUpsert(obj)
}

func (m *MeshWorkloadEventHandler) UpdateMeshWorkload(old, new *v1alpha1.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, new).Debugf("Handling event")
	return m.HandleMeshWorkloadUpsert(new)
}

func (m *MeshWorkloadEventHandler) DeleteMeshWorkload(obj *v1alpha1.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, obj).Debugf("Ignoring event")
	return nil
}

func (m *MeshWorkloadEventHandler) GenericMeshWorkload(obj *v1alpha1.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, obj).Errorf("Ignoring event")
	return nil
}
