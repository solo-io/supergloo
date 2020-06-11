package credentials

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/rotisserie/eris"
)

var (
	CredsNotFound = func(accountID string) error {
		return eris.Errorf("AWS credentials not found for mesh with name %s", accountID)
	}
)

type awsCredentialsStore struct {
	store sync.Map
}

func NewCredentialsGetter() AwsCredentialsStore {
	return &awsCredentialsStore{}
}

func (c *awsCredentialsStore) Get(accountID string) (*credentials.Credentials, error) {
	val, ok := c.store.Load(accountID)
	if !ok {
		return nil, CredsNotFound(accountID)
	}
	creds, ok := val.(*credentials.Credentials)
	if !ok {
		return nil, eris.New("Could not cast AWS credentials value.")
	}
	return creds, nil
}

func (c *awsCredentialsStore) Set(accountID string, creds *credentials.Credentials) {
	c.store.Store(accountID, creds)
}

func (c *awsCredentialsStore) Remove(accountId string) {
	c.store.Delete(accountId)
}
