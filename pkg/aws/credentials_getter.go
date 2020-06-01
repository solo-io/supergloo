package aws

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

type awsCredentialsGetter struct {
	//meshToCreds map[string]*credentials.Credentials
	meshToCreds sync.Map
}

func NewCredentialsGetter() AwsCredentialsGetter {
	return &awsCredentialsGetter{}
}

func (c *awsCredentialsGetter) Get(accountID string) (*credentials.Credentials, error) {
	//creds, ok := c.meshToCreds[accountID]
	val, ok := c.meshToCreds.Load(accountID)
	if !ok {
		return nil, CredsNotFound(accountID)
	}
	creds, ok := val.(*credentials.Credentials)
	if !ok {
		return nil, eris.New("Could not cast AWS credentials value")
	}
	return creds, nil
}

func (c *awsCredentialsGetter) Set(accountID string, creds *credentials.Credentials) {
	c.meshToCreds.Store(accountID, creds)
	//c.meshToCreds[accountID] = creds
}

func (c *awsCredentialsGetter) Remove(accountId string) {
	c.meshToCreds.Delete(accountId)
	//delete(c.meshToCreds, accountId)
}
