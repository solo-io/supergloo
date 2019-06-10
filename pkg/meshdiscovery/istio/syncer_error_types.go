package istio

import "fmt"

type ErrorType string

const (
	ErrorType_DetectingMeshPolicy ErrorType = "DetectingMeshPolicy"
)

type SyncerError struct {
	Type    ErrorType
	Message string
}

func (e *SyncerError) Error() string {
	return fmt.Sprintf("istio discovery error: %v", e.Message)
}

func newErrorDetectingMeshPolicy(err error) *SyncerError {
	return &SyncerError{
		Type:    ErrorType_DetectingMeshPolicy,
		Message: fmt.Sprintf("detecting default MeshPolicy: %v", err),
	}
}

func IsErrorDetectingMeshPolicy(err error) bool {
	if syncerErr, ok := err.(*SyncerError); ok {
		return syncerErr.Type == ErrorType_DetectingMeshPolicy
	}
	return false
}
