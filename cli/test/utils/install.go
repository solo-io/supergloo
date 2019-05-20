package utils

import (
	"context"
	"time"

	. "github.com/onsi/gomega"
	skclients "github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
)

func CompleteInstall(ctx context.Context, client v1.InstallClient, installRef core.ResourceRef, delay time.Duration) {

LOOP:
	for {
		select {
		case <-time.After(delay):
			break LOOP
		case <-ctx.Done():
			return
		}
	}

	install, err := client.Read(installRef.Namespace, installRef.Name, skclients.ReadOpts{Ctx: ctx})
	Expect(err).ToNot(HaveOccurred())
	install.Status.State = core.Status_Accepted
	_, err = client.Write(install, skclients.WriteOpts{Ctx: ctx, OverwriteExisting: true})
	Expect(err).ToNot(HaveOccurred())
}
