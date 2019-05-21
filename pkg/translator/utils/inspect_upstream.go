package utils

import (
	"fmt"

	"github.com/solo-io/gloo/projects/gloo/pkg/api/v1/plugins/kubernetes"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/solo-kit/pkg/errors"
)

func GetLabelsForUpstream(us *gloov1.Upstream) map[string]string {
	switch specType := us.UpstreamSpec.UpstreamType.(type) {
	case *gloov1.UpstreamSpec_Kube:
		return specType.Kube.Selector
	}
	// default to using the labels from the upstream
	return us.Metadata.Labels
}

func GetNamespaceForUpstream(us *gloov1.Upstream) string {
	switch specType := us.UpstreamSpec.UpstreamType.(type) {
	case *gloov1.UpstreamSpec_Kube:
		return specType.Kube.ServiceNamespace
	}
	// default to using the labels from the upstream
	return us.Metadata.Namespace
}

func GetUpstreamKubeSpec(us *gloov1.Upstream) (*kubernetes.UpstreamSpec, error) {
	switch specType := us.UpstreamSpec.UpstreamType.(type) {
	case *gloov1.UpstreamSpec_Kube:
		return specType.Kube, nil
	}
	return nil, errors.Errorf("not a kube upstream")
}

func GetHostForUpstream(us *gloov1.Upstream) (string, error) {
	hosts, err := GetHostsForUpstream(us)
	if err != nil {
		return "", errors.Wrapf(err, "failed to get hosts for upstream %s", us.Metadata.Ref())
	}
	if len(hosts) < 1 {
		return "", errors.Errorf("failed to get hosts for upstream %s", us.Metadata.Ref())
	}
	return hosts[0], nil
}

// TODO (ilackarms): consider multi-cluster here
func ServiceHost(serviceName, serviceNamespace string) string {
	return fmt.Sprintf("%v.%v.svc.cluster.local", serviceName, serviceNamespace)
}

func GetHostsForUpstream(us *gloov1.Upstream) ([]string, error) {
	switch specType := us.UpstreamSpec.UpstreamType.(type) {
	case *gloov1.UpstreamSpec_Aws:
		return nil, errors.Errorf("aws not implemented")
	case *gloov1.UpstreamSpec_Azure:
		return nil, errors.Errorf("azure not implemented")
	case *gloov1.UpstreamSpec_Kube:
		return []string{
			ServiceHost(specType.Kube.ServiceName, specType.Kube.ServiceNamespace),
			specType.Kube.ServiceName,
		}, nil
	case *gloov1.UpstreamSpec_Static:
		var hosts []string
		for _, h := range specType.Static.Hosts {
			hosts = append(hosts, h.Addr)
		}
		return hosts, nil
	}
	return nil, errors.Errorf("unsupported upstream type %v", us)
}

func GetPortForUpstream(us *gloov1.Upstream) (uint32, error) {
	switch specType := us.UpstreamSpec.UpstreamType.(type) {
	case *gloov1.UpstreamSpec_Aws:
		return 0, errors.Errorf("aws not implemented")
	case *gloov1.UpstreamSpec_Azure:
		return 0, errors.Errorf("azure not implemented")
	case *gloov1.UpstreamSpec_Kube:
		return specType.Kube.ServicePort, nil
	case *gloov1.UpstreamSpec_Static:
		// TODO(ilackarms): handle cases where port changes between hosts
		for _, h := range specType.Static.Hosts {
			return h.Port, nil
		}
		return 0, errors.Errorf("no hosts found on static upstream")
	}
	return 0, errors.Errorf("unknown upstream type")
}
