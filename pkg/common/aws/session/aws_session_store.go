package session

import (
	"sync"

	"github.com/aws/aws-app-mesh-controller-for-k8s/pkg/aws/throttle"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/session/credentials"
)

var (
	SessionNotFound = func(accountID, region string) error {
		return eris.Errorf("AWS session not found for account ID %s and region %s", accountID, region)
	}
)

type awsSessionStore struct {
	credentialsStore credentials.AwsCredentialsStore
	store            sync.Map // accountId string -> session *session.Session
}

func NewAwsSessionStore(
	credentialsStore credentials.AwsCredentialsStore,
) AwsSessionStore {
	return &awsSessionStore{
		credentialsStore: credentialsStore,
	}
}

type sessionKey struct {
	accountId string
	region    string
}

func (a *awsSessionStore) Add(accountId, region string) error {
	existingSess, err := a.Get(accountId, region)
	if existingSess != nil {
		return nil
	}

	creds, err := a.credentialsStore.Get(accountId)
	if err != nil {
		return err
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	})
	if err != nil {
		return err
	}
	a.addThrottler(sess)
	a.store.Store(sessionKey{accountId: accountId, region: region}, sess)
	return nil
}

func (a *awsSessionStore) Get(accountId, region string) (*session.Session, error) {
	val, ok := a.store.Load(sessionKey{accountId: accountId, region: region})
	if !ok {
		return nil, SessionNotFound(accountId, region)
	}
	sess, ok := val.(*session.Session)
	if !ok {
		return nil, eris.New("Could not cast to AWS session.")
	}
	return sess, nil
}

// Throttle both maximum burst and rate of requests.
func (a *awsSessionStore) addThrottler(sess *session.Session) {
	throttleCfg := throttle.NewDefaultServiceOperationsThrottleConfig()
	throttler := throttle.NewThrottler(throttleCfg)
	throttler.InjectHandlers(&sess.Handlers)
}
