package auth

import (
	"context"
	"time"

	"github.com/avast/retry-go"
	"github.com/rotisserie/eris"
	smh_core_types "github.com/solo-io/service-mesh-hub/pkg/api/core.smh.solo.io/v1alpha1/types"
	k8s_core "github.com/solo-io/service-mesh-hub/pkg/api/kubernetes/core/v1"
	"github.com/solo-io/service-mesh-hub/pkg/filesystem/files"
	k8s_core_types "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// visible for testing
	SecretTokenKey = "token"
)

var (
	// exponential backoff retry with an initial period of 0.1s for 7 iterations, which will mean a cumulative retry period of ~6s
	secretLookupOpts = []retry.Option{
		retry.Delay(time.Millisecond * 100),
		retry.Attempts(7),
		retry.DelayType(retry.BackOffDelay),
	}
)

func NewRemoteAuthorityConfigCreator(
	secretClient k8s_core.SecretClient,
	serviceAccountClient k8s_core.ServiceAccountClient,
) RemoteAuthorityConfigCreator {
	return &remoteAuthorityConfigCreator{
		serviceAccountClient: serviceAccountClient,
		secretClient:         secretClient,
	}
}

type remoteAuthorityConfigCreator struct {
	secretClient         k8s_core.SecretClient
	serviceAccountClient k8s_core.ServiceAccountClient
	fileReader           files.FileReader
}

func (r *remoteAuthorityConfigCreator) ConfigFromRemoteServiceAccount(
	ctx context.Context,
	targetClusterCfg *rest.Config,
	serviceAccountRef *smh_core_types.ResourceRef,
) (*rest.Config, error) {

	tokenSecret, err := r.waitForSecret(ctx, serviceAccountRef)
	if err != nil {
		return nil, SecretNotReady(err)
	}

	serviceAccountToken, ok := tokenSecret.Data[SecretTokenKey]
	if !ok {
		return nil, MalformedSecret
	}

	// make a copy of the config we were handed, with all user credentials removed
	// https://github.com/kubernetes/client-go/blob/9bbcc2938d41daa40d3080a1b6524afbe4e27bd9/rest/config.go#L542
	newCfg := rest.AnonymousClientConfig(targetClusterCfg)

	// authorize ourselves as the service account we were given
	newCfg.BearerToken = string(serviceAccountToken)

	return newCfg, nil
}

func (r *remoteAuthorityConfigCreator) waitForSecret(
	ctx context.Context,
	serviceAccountRef *smh_core_types.ResourceRef,
) (foundSecret *k8s_core_types.Secret, err error) {

	err = retry.Do(func() error {
		serviceAccount, err := r.serviceAccountClient.GetServiceAccount(
			ctx,
			client.ObjectKey{Name: serviceAccountRef.Name, Namespace: serviceAccountRef.Namespace},
		)
		if err != nil {
			return err
		}

		if len(serviceAccount.Secrets) == 0 {
			return eris.Errorf("service account %+v does not have a token secret associated with it", serviceAccountRef)
		}
		if len(serviceAccount.Secrets) != 1 {
			return eris.Errorf("service account %+v unexpectedly has more than one secret", serviceAccountRef)
		}

		secretName := serviceAccount.Secrets[0].Name

		secret, err := r.secretClient.GetSecret(ctx, client.ObjectKey{Name: secretName, Namespace: serviceAccountRef.Namespace})
		if err != nil {
			return err
		}

		foundSecret = secret
		return nil
	}, secretLookupOpts...)

	return
}
