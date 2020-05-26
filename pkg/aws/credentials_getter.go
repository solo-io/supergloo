package aws

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/rotisserie/eris"
)

var (
	CredsNotFound = func(accountID string) error {
		return eris.Errorf("AWS credentials not found for mesh with name %s", accountID)
	}
)

type awsCredentialsGetter struct {
	meshToCreds map[string]*credentials.Credentials
}

func NewCredentialsGetter() *awsCredentialsGetter {
	return &awsCredentialsGetter{
		meshToCreds: make(map[string]*credentials.Credentials),
	}
}

func (c *awsCredentialsGetter) Get(accountID string) (*credentials.Credentials, error) {
	creds, ok := c.meshToCreds[accountID]
	if !ok {
		return nil, CredsNotFound(accountID)
	}
	return creds, nil
}

func (c *awsCredentialsGetter) Set(accountID string, creds *credentials.Credentials) {
	c.meshToCreds[accountID] = creds
}

func (c *awsCredentialsGetter) Remove(accountId string) {
	delete(c.meshToCreds, accountId)
}
