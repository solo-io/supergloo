package appmesh

import (
	"strings"

	"github.com/ghodss/yaml"
	gloov1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	"github.com/solo-io/go-utils/protoutils"
	v1 "github.com/solo-io/supergloo/pkg/api/v1"
	kubev1 "k8s.io/api/core/v1"
)

// Contains the resources for a complete test scenario
type TestResourceSet map[string]PodsServicesUpstreamsTuple

func (t TestResourceSet) MustGetPodList() v1.PodList {
	var podList v1.PodList
	for _, tuple := range t {
		podList = append(podList, tuple.MustGetPodList()...)
	}
	return podList
}

func (t TestResourceSet) MustGetPods() []*kubev1.Pod {
	var pods []*kubev1.Pod
	for _, tuple := range t {
		pods = append(pods, tuple.MustGetPods()...)
	}
	return pods
}

func (t TestResourceSet) MustGetServices() []*kubev1.Service {
	var services []*kubev1.Service
	for _, tuple := range t {
		services = append(services, tuple.MustGetServices()...)
	}
	return services
}

func (t TestResourceSet) MustGetUpstreams() gloov1.UpstreamList {
	var upstreamList gloov1.UpstreamList
	for _, tuple := range t {
		upstreamList = append(upstreamList, tuple.MustGetUpstreamList()...)
	}
	return upstreamList
}

// Represents a set of related test resources
type PodsServicesUpstreamsTuple struct {
	Pods, Services, Upstreams []string
}

func (t *PodsServicesUpstreamsTuple) MustGetPodList() v1.PodList {
	var podList v1.PodList
	for _, podYaml := range t.Pods {
		for _, podManifest := range strings.Split(podYaml, "---") {
			var podObj v1.Pod
			err := yaml.Unmarshal([]byte(podManifest), &podObj)
			if err != nil {
				panic("failed to unmarshal test pod")
			}
			podList = append(podList, &podObj)
		}
	}
	return podList
}

func (t *PodsServicesUpstreamsTuple) MustGetPods() []*kubev1.Pod {
	var pods []*kubev1.Pod
	for _, podYaml := range t.Pods {
		for _, podManifest := range strings.Split(podYaml, "---") {
			var podObj kubev1.Pod
			err := yaml.Unmarshal([]byte(podManifest), &podObj)
			if err != nil {
				panic("failed to unmarshal test pod")
			}
			pods = append(pods, &podObj)
		}
	}
	return pods
}

func (t *PodsServicesUpstreamsTuple) MustGetServices() []*kubev1.Service {
	var services []*kubev1.Service
	for _, svcYaml := range t.Services {
		for _, svcManifest := range strings.Split(svcYaml, "---") {
			var svcObj kubev1.Service
			err := yaml.Unmarshal([]byte(svcManifest), &svcObj)
			if err != nil {
				panic("failed to unmarshal test service")
			}
			services = append(services, &svcObj)
		}
	}
	return services
}

func (t *PodsServicesUpstreamsTuple) MustGetUpstreamList() gloov1.UpstreamList {
	var upstreamList gloov1.UpstreamList
	for _, upstreamYaml := range t.Upstreams {
		for _, usManifest := range strings.Split(upstreamYaml, "---") {
			var us gloov1.Upstream
			err := protoutils.UnmarshalYaml([]byte(usManifest), &us)
			if err != nil {
				panic("failed to unmarshal test upstream")
			}
			upstreamList = append(upstreamList, &us)
		}
	}
	return upstreamList
}
