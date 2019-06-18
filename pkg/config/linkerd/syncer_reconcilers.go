package linkerd

import (
	"context"

	"github.com/solo-io/supergloo/pkg/api/clientset"

	"github.com/solo-io/supergloo/pkg/config/utils"

	linkerdv1 "github.com/solo-io/supergloo/pkg/api/external/linkerd/v1"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/supergloo/pkg/translator/linkerd"
)

type Reconcilers interface {
	ReconcileAll(ctx context.Context, config *linkerd.MeshConfig) error
}

type linkerdReconcilers struct {
	ownerLabels map[string]string

	serviceProfileClientLoader clientset.ServiceProfileClientLoader
}

func NewLinkerdReconcilers(ownerLabels map[string]string,
	serviceProfileClientLoader clientset.ServiceProfileClientLoader) Reconcilers {
	return &linkerdReconcilers{
		ownerLabels:                ownerLabels,
		serviceProfileClientLoader: serviceProfileClientLoader,
	}
}

func (s *linkerdReconcilers) ReconcileAll(ctx context.Context, config *linkerd.MeshConfig) error {
	logger := contextutils.LoggerFrom(ctx)

	logger.Infof("ServiceProfiles: %v", config.ServiceProfiles.Names())
	utils.SetLabels(s.ownerLabels, config.ServiceProfiles.AsResources()...)
	serviceProfilClient, err := s.serviceProfileClientLoader()
	if err != nil {
		return err
	}

	if err := linkerdv1.NewServiceProfileReconciler(serviceProfilClient).Reconcile(
		"",
		config.ServiceProfiles,
		nil,
		clients.ListOpts{
			Ctx:      ctx,
			Selector: s.ownerLabels,
		},
	); err != nil {
		return errors.Wrapf(err, "reconciling serviceProfiles")
	}

	return nil
}
