package multicluster

import "context"

type apiHandler struct {
}

func (a *apiHandler) APIAdded(ctx context.Context) error {
	panic("implement me")
}

func (a *apiHandler) APIRemoved(ctx context.Context) error {
	panic("implement me")
}
