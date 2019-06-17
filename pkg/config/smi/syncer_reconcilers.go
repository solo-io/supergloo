package smi

import (
	"context"

	"github.com/solo-io/supergloo/pkg/api/clientset"

	"github.com/solo-io/supergloo/pkg/config/utils"

	accessv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/smi/access/v1alpha1"
	specsv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/smi/specs/v1alpha1"
	splitv1alpha1 "github.com/solo-io/supergloo/pkg/api/external/smi/split/v1alpha1"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/translator/smi"
)

type Reconcilers interface {
	CanReconcile() bool
	ReconcileAll(ctx context.Context, config *smi.MeshConfig) error
}

type smiReconcilers struct {
	ownerLabels                map[string]string
	trafficTargetClientLoader  clientset.TrafficTargetClientLoader
	httpRouteGroupClientLoader clientset.HTTPRouteGroupClientLoader
	trafficSplitClientLoader   clientset.TrafficSplitClientLoader
}

func NewSMIReconcilers(ownerLabels map[string]string, trafficTargetClientLoader clientset.TrafficTargetClientLoader, httpRouteGroupClientLoader clientset.HTTPRouteGroupClientLoader, trafficSplitClientLoader clientset.TrafficSplitClientLoader) Reconcilers {
	return &smiReconcilers{ownerLabels: ownerLabels, trafficTargetClientLoader: trafficTargetClientLoader, httpRouteGroupClientLoader: httpRouteGroupClientLoader, trafficSplitClientLoader: trafficSplitClientLoader}
}

func (s *smiReconcilers) CanReconcile() bool {
	_, err := s.trafficTargetClientLoader()
	if err != nil {
		return false
	}
	_, err = s.httpRouteGroupClientLoader()
	if err != nil {
		return false
	}
	_, err = s.trafficSplitClientLoader()
	if err != nil {
		return false
	}
	return true
}

func (s *smiReconcilers) ReconcileAll(ctx context.Context, config *smi.MeshConfig) error {
	logger := contextutils.LoggerFrom(ctx)

	logger.Infof("TrafficTargets: %v"+
		"HTTPRouteGroups: %v"+
		"TrafficSplits: %v",
		config.SecurityConfig.TrafficTargets.Names(),
		config.SecurityConfig.HTTPRouteGroups.Names(),
		config.RoutingConfig.TrafficSplits.Names(),
	)
	utils.SetLabels(s.ownerLabels, config.SecurityConfig.TrafficTargets.AsResources()...)
	utils.SetLabels(s.ownerLabels, config.SecurityConfig.HTTPRouteGroups.AsResources()...)
	utils.SetLabels(s.ownerLabels, config.RoutingConfig.TrafficSplits.AsResources()...)

	trafficTargetClient, err := s.trafficTargetClientLoader()
	if err != nil {
		return err
	}
	httpRouteGroupClient, err := s.httpRouteGroupClientLoader()
	if err != nil {
		return err
	}
	trafficSplitClient, err := s.trafficSplitClientLoader()
	if err != nil {
		return err
	}

	if err := accessv1alpha1.NewTrafficTargetReconciler(trafficTargetClient).Reconcile(
		"",
		config.SecurityConfig.TrafficTargets,
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: s.ownerLabels,
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling trafficTargets")
	}

	if err := specsv1alpha1.NewHTTPRouteGroupReconciler(httpRouteGroupClient).Reconcile(
		"",
		config.SecurityConfig.HTTPRouteGroups,
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: s.ownerLabels,
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling trafficTargets")
	}

	if err := splitv1alpha1.NewTrafficSplitReconciler(trafficSplitClient).Reconcile(
		"",
		config.RoutingConfig.TrafficSplits,
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: s.ownerLabels,
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling trafficTargets")
	}

	return nil
}
