package mc_manager

import "context"

//go:generate mockgen -source rest_interfaces.go -destination ./mocks/rest_interfaces.go

type RestAPIHandler interface {
	APIAdded(ctx context.Context) error
	APIRemoved(ctx context.Context) error
}

type APIClientGetter interface {
	GetClientForAPI(ctx context.Context) error
}
