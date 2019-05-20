package helpers

import (
	"context"
	"time"

	"github.com/solo-io/go-utils/contextutils"
	"github.com/solo-io/go-utils/errors"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/supergloo/cli/pkg/helpers/clients"
)

var TimeoutError = errors.Errorf("timed out while waiting for install to transition to 'accepted' status")

// TODO: generalize and move to go-utils (or solo-kit?)
// Blocks until the install transitions to the given status, times out, or any error occurs
func WaitForInstallStatus(ctx context.Context, installRef core.ResourceRef, desiredState core.Status_State, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	contextutils.LoggerFrom(ctx).Infof("Waiting for installation to transition to status '%s'...", desiredState)

	installListChan, errorChan, err := clients.MustInstallClient().Watch(installRef.Namespace, skclients.WatchOpts{Ctx: ctx})
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return TimeoutError
		case err, ok := <-errorChan:
			if ok && err != nil {
				return errors.Wrapf(err, "unexpected error while watching installs")
			}
		case installList := <-installListChan:
			for _, install := range installList {
				if this := install.Metadata.Ref(); !(&this).Equal(&installRef) {
					continue
				}

				if install.Status.State == desiredState {
					contextutils.LoggerFrom(ctx).Infof("Installation reached status '%s'", desiredState)
					return nil
				}
			}
		}
	}
}
