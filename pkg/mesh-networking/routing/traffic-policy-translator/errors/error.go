package errors

import "github.com/rotisserie/eris"

var (
	TrafficPolicyConflictError = eris.New("Found conflicting TrafficPolicy")
)
