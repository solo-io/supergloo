package istio

import (
	"context"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/api/clientset"
	"github.com/solo-io/supergloo/pkg/api/external/istio/networking/v1alpha3"
	rbacv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/istio/rbac/v1alpha1"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	"github.com/solo-io/supergloo/pkg/config/utils"
	"github.com/solo-io/supergloo/pkg/translator/istio"
)

type Reconcilers interface {
	ReconcileAll(ctx context.Context, config *istio.MeshConfig) error
}

type istioReconcilers struct {
	ownerLabels map[string]string

	rbacConfigClientLoader         clientset.RbacConfigClientLoader
	serviceRoleClientLoader        clientset.ServiceRoleClientLoader
	serviceRoleBindingClientLoader clientset.ServiceRoleBindingClientLoader
	meshPolicyClientLoader         clientset.MeshPolicyClientLoader
	destinationRuleClientLoader    clientset.DestinationRuleClientLoader
	virtualServiceClientLoader     clientset.VirtualServiceClientLoader
	tlsSecretReconciler            v1.TlsSecretReconciler
}

func NewIstioReconcilers(ownerLabels map[string]string,
	rbacConfigClientLoader clientset.RbacConfigClientLoader,
	serviceRoleClientLoader clientset.ServiceRoleClientLoader,
	serviceRoleBindingClientLoader clientset.ServiceRoleBindingClientLoader,
	meshPolicyClientLoader clientset.MeshPolicyClientLoader,
	destinationRuleClientLoader clientset.DestinationRuleClientLoader,
	virtualServiceClientLoader clientset.VirtualServiceClientLoader,
	tlsSecretReconciler v1.TlsSecretReconciler) Reconcilers {
	return &istioReconcilers{
		ownerLabels:                    ownerLabels,
		rbacConfigClientLoader:         rbacConfigClientLoader,
		serviceRoleClientLoader:        serviceRoleClientLoader,
		serviceRoleBindingClientLoader: serviceRoleBindingClientLoader,
		meshPolicyClientLoader:         meshPolicyClientLoader,
		destinationRuleClientLoader:    destinationRuleClientLoader,
		virtualServiceClientLoader:     virtualServiceClientLoader,
		tlsSecretReconciler:            tlsSecretReconciler,
	}
}

func (s *istioReconcilers) ReconcileAll(ctx context.Context, config *istio.MeshConfig) error {
	logger := contextutils.LoggerFrom(ctx)

	// lazy loading of clients
	// checks that the crd is registered if backed by a crd-based client
	rbacConfigClient, err := s.rbacConfigClientLoader()
	if err != nil {
		return err
	}
	destinationRuleClient, err := s.destinationRuleClientLoader()
	if err != nil {
		return err
	}
	virtualServiceClient, err := s.virtualServiceClientLoader()
	if err != nil {
		return err
	}
	serviceRoleClient, err := s.serviceRoleClientLoader()
	if err != nil {
		return err
	}
	serviceRoleBindingClient, err := s.serviceRoleBindingClientLoader()
	if err != nil {
		return err
	}

	// special case for MeshPolicy, only update existing if necessary
	defaultMeshPolicy := config.MeshPolicy
	if defaultMeshPolicy != nil {
		logger.Infof("MeshPolicy: %v", defaultMeshPolicy.Metadata.Name)
		utils.SetLabels(s.ownerLabels, defaultMeshPolicy)

		meshPolicyClient, err := s.meshPolicyClientLoader()
		if err != nil {
			return err
		}
		var skipWriting bool
		original, err := meshPolicyClient.Read(defaultMeshPolicy.Metadata.Name, clients.ReadOpts{Ctx: ctx})
		if err == nil {
			if original.Hash() != defaultMeshPolicy.Hash() {
				defaultMeshPolicy.Metadata.ResourceVersion = original.Metadata.ResourceVersion
			} else {
				skipWriting = true
			}
		}
		if !skipWriting {
			if _, err := meshPolicyClient.Write(defaultMeshPolicy, clients.WriteOpts{Ctx: ctx, OverwriteExisting: true}); err != nil {
				return err
			}
		}
	}

	// this list should always either be empty or contain the global rbac config
	var rbacConfigsToReconcile rbacv1alpha1.RbacConfigList
	if config.RbacConfig != nil {
		logger.Infof("RbacConfig: %v", config.RbacConfig.Metadata.Name)
		utils.SetLabels(s.ownerLabels, config.RbacConfig)
		rbacConfigsToReconcile = append(rbacConfigsToReconcile, config.RbacConfig)
	}
	if err := rbacv1alpha1.NewRbacConfigReconciler(rbacConfigClient).Reconcile(
		"",
		rbacConfigsToReconcile, // rbac config is a singleton
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: nil, // allows overwriting a user-created rbac config
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling default rbac config")
	}

	// this list should always either be empty or contain the global cacerts root cert secret
	var tlsSecretsToReconcile v1.TlsSecretList
	if config.RootCert != nil {
		logger.Infof("RootCert: %v", config.RootCert.Metadata.Name)
		utils.SetLabels(s.ownerLabels, config.RootCert)
		tlsSecretsToReconcile = append(tlsSecretsToReconcile, config.RootCert)
	}
	if err := s.tlsSecretReconciler.Reconcile(
		"",
		tlsSecretsToReconcile, // root cert is a singleton
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: s.ownerLabels,
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling cacerts root cert")
	}

	logger.Infof("DestinationRules: %v", config.DestinationRules.Names())
	utils.SetLabels(s.ownerLabels, config.DestinationRules.AsResources()...)
	if err := v1alpha3.NewDestinationRuleReconciler(destinationRuleClient).Reconcile(
		"",
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
	utils.SetLabels(s.ownerLabels, config.VirtualServices.AsResources()...)
	if err := v1alpha3.NewVirtualServiceReconciler(virtualServiceClient).Reconcile(
		"",
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
	utils.SetLabels(s.ownerLabels, config.ServiceRoles.AsResources()...)
	if err := rbacv1alpha1.NewServiceRoleReconciler(serviceRoleClient).Reconcile(
		"",
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
	utils.SetLabels(s.ownerLabels, config.ServiceRoleBindings.AsResources()...)
	if err := rbacv1alpha1.NewServiceRoleBindingReconciler(serviceRoleBindingClient).Reconcile(
		"",
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
