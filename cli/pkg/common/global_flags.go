package common

import "k8s.io/client-go/rest"

type GlobalFlagConfig struct {
	MasterWriteNamespace string
	MasterKubeConfig     *rest.Config
}
