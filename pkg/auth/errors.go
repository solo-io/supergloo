package auth

import (
	"github.com/rotisserie/eris"
)

var (
	MalformedSecret = eris.New("service account secret does not contain a bearer token")
	SecretNotReady  = func(err error) error {
		return eris.Wrap(err, "secret for service account not ready yet")
	}
)
