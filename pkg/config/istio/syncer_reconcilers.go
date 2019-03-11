package istio

import (
	"context"

	"github.com/solo-io/supergloo/pkg/translator/istio"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	policyv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/authorization/v1alpha1"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	rbacv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
)

type Reconcilers interface {
	ReconcileAll(ctx context.Context, writeNamespace string, config *istio.MeshConfig) error
}

type istioReconcilers struct {
	ownerLabels map[string]string

	rbacConfigReconciler         rbacv1alpha1.RbacConfigReconciler
	serviceRoleReconciler        rbacv1alpha1.ServiceRoleReconciler
	serviceRoleBindingReconciler rbacv1alpha1.ServiceRoleBindingReconciler
	meshPolicyReconciler         policyv1alpha1.MeshPolicyReconciler
	destinationRuleReconciler    v1alpha3.DestinationRuleReconciler
	virtualServiceReconciler     v1alpha3.VirtualServiceReconciler
}

func NewIstioReconcilers(ownerLabels map[string]string,
	rbacConfigReconciler rbacv1alpha1.RbacConfigReconciler,
	serviceRoleReconciler rbacv1alpha1.ServiceRoleReconciler,
	serviceRoleBindingReconciler rbacv1alpha1.ServiceRoleBindingReconciler,
	meshPolicyReconciler policyv1alpha1.MeshPolicyReconciler,
	destinationRuleReconciler v1alpha3.DestinationRuleReconciler,
	virtualServiceReconciler v1alpha3.VirtualServiceReconciler) Reconcilers {
	return &istioReconcilers{
		ownerLabels:                  ownerLabels,
		rbacConfigReconciler:         rbacConfigReconciler,
		serviceRoleReconciler:        serviceRoleReconciler,
		serviceRoleBindingReconciler: serviceRoleBindingReconciler,
		meshPolicyReconciler:         meshPolicyReconciler,
		destinationRuleReconciler:    destinationRuleReconciler,
		virtualServiceReconciler:     virtualServiceReconciler,
	}
}

func (s *istioReconcilers) ReconcileAll(ctx context.Context, writeNamespace string, config *istio.MeshConfig) error {
	logger := contextutils.LoggerFrom(ctx)

	logger.Infof("MeshPolicy: %v", config.MeshPolicy.Metadata.Name)
	s.setLabels(config.MeshPolicy)
	if err := s.meshPolicyReconciler.Reconcile(
		writeNamespace,
		policyv1alpha1.MeshPolicyList{config.MeshPolicy}, // mesh policy is a singleton
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: s.ownerLabels,
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling default mesh policy")
	}

	logger.Infof("RbacConfig: %v", config.RbacConfig.Metadata.Name)
	s.setLabels(config.RbacConfig)
	if err := s.rbacConfigReconciler.Reconcile(
		writeNamespace,
		rbacv1alpha1.RbacConfigList{config.RbacConfig}, // rbac config is a singleton
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: s.ownerLabels,
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling default rbac config")
	}

	logger.Infof("DestinationRules: %v", config.DestinationRules.Names())
	s.setLabels(config.DestinationRules.AsResources()...)
	if err := s.destinationRuleReconciler.Reconcile(
		writeNamespace,
		config.DestinationRules,
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: s.ownerLabels,
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling destination rules")
	}

	logger.Infof("VirtualServices: %v", config.VirtualServices.Names())
	s.setLabels(config.VirtualServices.AsResources()...)
	if err := s.virtualServiceReconciler.Reconcile(
		writeNamespace,
		config.VirtualServices,
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: s.ownerLabels,
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling virtual services")
	}

	logger.Infof("ServiceRoles: %v", config.ServiceRoles.Names())
	s.setLabels(config.ServiceRoles.AsResources()...)
	if err := s.serviceRoleReconciler.Reconcile(
		writeNamespace,
		config.ServiceRoles,
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: s.ownerLabels,
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling service roles")
	}

	logger.Infof("ServiceRoleBindings: %v", config.ServiceRoleBindings.Names())
	s.setLabels(config.ServiceRoleBindings.AsResources()...)
	if err := s.serviceRoleBindingReconciler.Reconcile(
		writeNamespace,
		config.ServiceRoleBindings,
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: s.ownerLabels,
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling service role bindings")
	}

	return nil
}

// set labels on all resources, required for our reconciler
func (s *istioReconcilers) setLabels(rcs ...resources.Resource) {
	for _, res := range rcs {
		resources.UpdateMetadata(res, func(meta *core.Metadata) {
			if meta.Labels == nil {
				meta.Labels = make(map[string]string)
			}
			for k, v := range s.ownerLabels {
				meta.Labels[k] = v
			}
		})
	}
}
