package auth

import (
	"time"

	"github.com/avast/retry-go"
	"github.com/rotisserie/eris"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	k8sapiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

// create a kube config that can authorize to the target cluster as the service account from that target cluster
//go:generate mockgen -destination ./mocks/mock_remote_authority_config_creator.go github.com/solo-io/mesh-projects/pkg/auth RemoteAuthorityConfigCreator
type RemoteAuthorityConfigCreator interface {

	// Returns a `*rest.Config` that points to the same cluster as `targetClusterCfg`, but authorizes itself using the
	// bearer token belonging to the service account at `serviceAccountRef` in the target cluster
	//
	// NB: This function blocks the current go routine for up to 6 seconds while waiting for the service account's secret
	// to appear, by performing an exponential backoff retry loop
	ConfigFromRemoteServiceAccount(targetClusterCfg *rest.Config, serviceAccountRef *core.ResourceRef) (*rest.Config, error)
}

type remoteAuthorityConfigCreator struct {
	saClient     ServiceAccountClient
	secretClient SecretClient
}

func NewRemoteAuthorityConfigCreator(kubeClients kubernetes.Interface, writeNamespace string) RemoteAuthorityConfigCreator {
	return &remoteAuthorityConfigCreator{
		saClient:     kubeClients.CoreV1().ServiceAccounts(writeNamespace),
		secretClient: kubeClients.CoreV1().Secrets(writeNamespace),
	}
}

func NewRemoteAuthorityConfigCreatorForTest(saClient ServiceAccountClient, secretClient SecretClient) RemoteAuthorityConfigCreator {
	return &remoteAuthorityConfigCreator{
		saClient:     saClient,
		secretClient: secretClient,
	}
}

func (r *remoteAuthorityConfigCreator) ConfigFromRemoteServiceAccount(
	targetClusterCfg *rest.Config,
	serviceAccountRef *core.ResourceRef) (*rest.Config, error) {
	tokenSecret, err := r.waitForSecret(serviceAccountRef)
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
	serviceAccountRef *core.ResourceRef) (foundSecret *k8sapiv1.Secret, err error) {

	err = retry.Do(func() error {
		serviceAccount, err := r.saClient.Get(serviceAccountRef.Name, v1.GetOptions{})
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

		secret, err := r.secretClient.Get(secretName, v1.GetOptions{})
		if err != nil {
			return err
		}

		foundSecret = secret
		return nil
	}, secretLookupOpts...)

	return
}
