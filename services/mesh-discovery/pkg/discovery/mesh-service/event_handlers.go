package mesh_service

import (
	"context"

	"github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/mesh-projects/pkg/logging"
	corev1 "k8s.io/api/core/v1"
)

/**********
* kind-specific handlers that just forward events on to their callback
* we're not attaching these methods to the parent object because function overloading
* is apparently too hard for Go >:(
**********/

type ServiceEventHandler struct {
	Ctx                 context.Context
	ClusterName         string
	HandleServiceUpsert func(service *corev1.Service) error
}

func (s *ServiceEventHandler) Create(obj *corev1.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, s.ClusterName).Debugf("Handling event: %+v", obj)
	return s.HandleServiceUpsert(obj)
}

func (s *ServiceEventHandler) Update(old, new *corev1.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, s.ClusterName).Debugf("Handling event: old %+v new %+v", old, new)
	return s.HandleServiceUpsert(new)
}

func (s *ServiceEventHandler) Delete(obj *corev1.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, s.ClusterName).Warnf("Ignoring delete event: %+v", obj)
	return nil
}

func (s *ServiceEventHandler) Generic(obj *corev1.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, s.ClusterName).Errorf("Ignoring generic event: %+v", obj)
	return nil
}

type MeshWorkloadEventHandler struct {
	Ctx                      context.Context
	ClusterName              string
	HandleMeshWorkloadUpsert func(meshWorkload *v1alpha1.MeshWorkload) error
}

func (m *MeshWorkloadEventHandler) Create(obj *v1alpha1.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, m.ClusterName).Debugf("Handling event: %+v", obj)
	return m.HandleMeshWorkloadUpsert(obj)
}

func (m *MeshWorkloadEventHandler) Update(old, new *v1alpha1.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, m.ClusterName).Debugf("Handling event: old %+v new %+v", old, new)
	return m.HandleMeshWorkloadUpsert(new)
}

func (m *MeshWorkloadEventHandler) Delete(obj *v1alpha1.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, m.ClusterName).Debugf("Ignoring delete event:: %+v", obj)
	return nil
}

func (m *MeshWorkloadEventHandler) Generic(obj *v1alpha1.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, m.ClusterName).Errorf("Ignoring generic event: %+v", obj)
	return nil
}
