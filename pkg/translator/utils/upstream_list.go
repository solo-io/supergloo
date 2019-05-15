package utils

import (
	"reflect"
	"sort"

	"github.com/solo-io/go-utils/stringutils"

	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/errors"
)

// one (kube) service that maps to multiple upstreams
type UpstreamService struct {
	Host      string
	LabelSets []map[string]string
	Ports     []uint32
	Upstreams gloov1.UpstreamList // the upstreams this service was created from
}

func UpstreamServicesByHost(upstreams gloov1.UpstreamList) (map[string]*UpstreamService, error) {
	services := make(map[string]*UpstreamService)
	for _, us := range upstreams {
		host, err := GetHostForUpstream(us)
		if err != nil {
			return nil, err
		}
		if service, ok := services[host]; ok {
			service.Upstreams = append(service.Upstreams, us)
			continue
		}
		service, err := ServiceFromHost(host, upstreams)
		if err != nil {
			return nil, err
		}
		services[host] = service
	}
	return services, nil
}

func GetUpstreamHostPortsLabels(us *gloov1.Upstream) (string, uint32, map[string]string, error) {
	labels := GetLabelsForUpstream(us)
	host, err := GetHostForUpstream(us)
	if err != nil {
		return "", 0, nil, errors.Wrapf(err, "getting host for upstream")
	}
	port, err := GetPortForUpstream(us)
	if err != nil {
		return "", 0, nil, errors.Wrapf(err, "getting port for upstream")
	}
	return host, port, labels, nil
}

// only selects the first upstream in each list with a unique host, drop the others
func ServiceFromHost(host string, upstreams gloov1.UpstreamList) (*UpstreamService, error) {
	service := &UpstreamService{Host: host}

	for _, us := range upstreams {
		usHost, port, labels, err := GetUpstreamHostPortsLabels(us)
		if err != nil {
			return nil, err
		}
		if host != usHost {
			continue
		}
		var duplicateLabels, duplicatePort bool
		for _, foundLabels := range service.LabelSets {
			if reflect.DeepEqual(labels, foundLabels) {
				duplicateLabels = true
				break
			}
		}
		for _, foundPort := range service.Ports {
			if port == foundPort {
				duplicatePort = true
				break
			}
		}
		if !duplicatePort {
			service.Ports = append(service.Ports, port)
		}
		if !duplicateLabels {
			service.LabelSets = append(service.LabelSets, labels)
		}
		service.Upstreams = append(service.Upstreams, us)
	}

	return service, nil
}

func HostsForUpstreams(upstreams gloov1.UpstreamList) ([]string, error) {
	var hosts []string
	for _, us := range upstreams {
		host, err := GetHostForUpstream(us)
		if err != nil {
			return nil, errors.Wrapf(err, "getting host for upstream")
		}
		hosts = append(hosts, host)
	}
	hosts = stringutils.Unique(hosts)
	sort.Strings(hosts)
	return hosts, nil
}

func PortsFromUpstreams(upstreams gloov1.UpstreamList) ([]uint32, error) {
	var ports []uint32
addPorts:
	for _, us := range upstreams {
		port, err := GetPortForUpstream(us)
		if err != nil {
			return nil, err
		}
		for _, p := range ports {
			if p == port {
				continue addPorts
			}
		}
		ports = append(ports, port)
	}
	return ports, nil
}

func LabelsFromUpstreams(upstreams gloov1.UpstreamList) ([]map[string]string, error) {
	var labels []map[string]string
addLabels:
	for _, us := range upstreams {
		label := GetLabelsForUpstream(us)
		for _, p := range labels {
			if reflect.DeepEqual(p, label) {
				continue addLabels
			}
		}
		labels = append(labels, label)
	}
	return labels, nil
}
