package k8s

import (
	"context"

	zephyr_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	"github.com/solo-io/service-mesh-hub/pkg/logging"
	k8s_core "k8s.io/api/core/v1"
)

/**********
* kind-specific handlers that just forward events on to their callback
* we're not attaching these methods to the parent object because function overloading
* is apparently too hard for Go >:(
**********/

type ServiceEventHandler struct {
	Ctx                 context.Context
	HandleServiceUpsert func(service *k8s_core.Service) error
}

func (s *ServiceEventHandler) CreateService(obj *k8s_core.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, obj).Debugf("Handling event")
	return s.HandleServiceUpsert(obj)
}

func (s *ServiceEventHandler) UpdateService(old, new *k8s_core.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, new).Debugf("Handling event")
	return s.HandleServiceUpsert(new)
}

func (s *ServiceEventHandler) DeleteService(obj *k8s_core.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, obj).Warnf("Ignoring event")
	return nil
}

func (s *ServiceEventHandler) GenericService(obj *k8s_core.Service) error {
	logging.BuildEventLogger(s.Ctx, logging.CreateEvent, obj).Errorf("Ignoring event")
	return nil
}

type MeshWorkloadEventHandler struct {
	Ctx                      context.Context
	HandleMeshWorkloadUpsert func(meshWorkload *zephyr_discovery.MeshWorkload) error
}

func (m *MeshWorkloadEventHandler) CreateMeshWorkload(obj *zephyr_discovery.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, obj).Debugf("Handling event")
	return m.HandleMeshWorkloadUpsert(obj)
}

func (m *MeshWorkloadEventHandler) UpdateMeshWorkload(old, new *zephyr_discovery.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, new).Debugf("Handling event")
	return m.HandleMeshWorkloadUpsert(new)
}

func (m *MeshWorkloadEventHandler) DeleteMeshWorkload(obj *zephyr_discovery.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, obj).Debugf("Ignoring event")
	return nil
}

func (m *MeshWorkloadEventHandler) GenericMeshWorkload(obj *zephyr_discovery.MeshWorkload) error {
	logging.BuildEventLogger(m.Ctx, logging.CreateEvent, obj).Errorf("Ignoring event")
	return nil
}
