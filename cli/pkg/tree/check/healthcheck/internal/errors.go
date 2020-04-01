package internal

import (
	"github.com/rotisserie/eris"
	"github.com/solo-io/mesh-projects/pkg/api/core.zephyr.solo.io/v1alpha1/types"
	"github.com/solo-io/mesh-projects/pkg/version"
)

var (
	GenericCheckFailed = func(err error) error {
		return eris.Wrap(err, "check failed with unexpected error")
	}
	NamespaceDoesNotExist = func(ns string) error {
		return eris.Errorf("namespace '%s' does not exist", ns)
	}
	KubernetesApiServerUnreachable = func(err error) error {
		return eris.Wrap(err, "the Kubernetes API server is not reachable")
	}
	KubernetesServerVersionUnsupported = func(minorVersion string) error {
		return eris.Errorf("Kubernetes version 1.%s unsupported; only Kubernetes version 1.%d and later is supported", minorVersion, version.MinimumSupportedKubernetesMinorVersion)
	}
	FederationRecordingHasFailed = func(meshServiceName, installNamespace string, nonAcceptedStatus types.Status_State) error {
		return eris.Errorf("failed to write federation metadata to mesh service '%s.%s'; status is '%s'", meshServiceName, installNamespace, nonAcceptedStatus.String())
	}
	NoServiceMeshHubComponentsExist = eris.New("no Service Mesh Hub components are installed yet")
)
