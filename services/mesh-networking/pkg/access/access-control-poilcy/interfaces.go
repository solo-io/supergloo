package access_control_poilcy

import "context"

//go:generate mockgen -source ./interfaces.go -destination mocks/mock_interfaces.go

type AccessControlPolicyTranslator interface {
	Start(ctx context.Context) error
}
