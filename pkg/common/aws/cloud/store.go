package cloud

import (
	"sync"

	"github.com/aws/aws-app-mesh-controller-for-k8s/pkg/aws/throttle"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/rotisserie/eris"
	"github.com/solo-io/service-mesh-hub/pkg/common/aws/clients"
	credentials2 "github.com/solo-io/service-mesh-hub/pkg/common/aws/cloud/credentials"
	"github.com/solo-io/service-mesh-hub/pkg/mesh-discovery/compute-target/aws/clients/eks_temp"
)

var (
	AwsCloudNotFound = func(accountID, region string) error {
		return eris.Errorf("AWS session not found for account ID %s and region %s", accountID, region)
	}
)

type awsCloudStore struct {
	appmeshClientFactory clients.AppmeshClientFactory
	credsStore           credentials2.AwsCredentialsStore
	awsCloudStore        sync.Map // accountId string -> awsCloud AwsCloud
}

type sessionKey struct {
	accountId string
	region    string
}

func NewAwsCloudStore(
	appmeshClientFactory clients.AppmeshClientFactory,
) AwsCloudStore {
	return &awsCloudStore{
		appmeshClientFactory: appmeshClientFactory,
	}
}

func (a *awsCloudStore) Add(accountId string, creds *credentials.Credentials) {
	a.credsStore.Set(accountId, creds)
}

func (a *awsCloudStore) Get(accountId, region string) (*AwsCloud, error) {
	val, ok := a.awsCloudStore.Load(sessionKey{accountId: accountId, region: region})
	if !ok {
		awsCloud, err := a.instantiateNewCloud(accountId, region)
		if err != nil {
			return nil, err
		}
		return awsCloud, nil
	}
	cloud, ok := val.(*AwsCloud)
	if !ok {
		return nil, eris.New("Could not cast to AwsCloud.")
	}
	return cloud, nil
}

func (a *awsCloudStore) instantiateNewCloud(accountId, region string) (*AwsCloud, error) {
	creds, err := a.credsStore.Get(accountId)
	if err != nil {
		return nil, err
	}
	sess, err := a.buildSession(region, creds)
	if err != nil {
		return nil, err
	}
	awsCloud := &AwsCloud{
		Appmesh: a.appmeshClientFactory(appmesh.New(sess)),
		Eks:     eks_temp.NewEksClient(sess),
	}
	a.awsCloudStore.Store(sessionKey{accountId: accountId, region: region}, &awsCloud)
	return awsCloud, nil
}

func (a *awsCloudStore) buildSession(region string, creds *credentials.Credentials) (*session.Session, error) {
	sess, err := session.NewSession(&aws.Config{
		Credentials: creds,
		Region:      aws.String(region),
	})
	if err != nil {
		return nil, err
	}
	a.addThrottler(sess)
	return sess, nil
}

// Throttle both maximum burst and rate of requests.
func (a *awsCloudStore) addThrottler(sess *session.Session) {
	throttleCfg := throttle.NewDefaultServiceOperationsThrottleConfig()
	throttler := throttle.NewThrottler(throttleCfg)
	throttler.InjectHandlers(&sess.Handlers)
}
