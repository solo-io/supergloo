package rest_manager

import (
	"context"

	"github.com/rotisserie/eris"
)

//go:generate mockgen -source interfaces.go -destination ./mocks/interfaces.go

type RestAPIProvider string

const (
	AppMesh = "AppMesh"
)

func (a RestAPIProvider) IsProviderValid() error {
	switch a {
	case AppMesh:
		return nil
	default:
		return eris.Errorf("Unrecognized REST API provider: %s", a)
	}
}

type RestAPIHandler interface {
	APIAdded(ctx context.Context, apiProvider RestAPIProvider) error
	APIRemoved(ctx context.Context, apiProvider RestAPIProvider) error
}

type RestAPIClientGetter interface {
	GetClientForRestAPI(ctx context.Context, apiProvider RestAPIProvider) error
}
