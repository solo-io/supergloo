package test_utils

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "github.com/onsi/gomega"

	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	istio_client_networking_types "istio.io/client-go/pkg/apis/networking/v1alpha3"
	k8s_meta_types "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func GetData() []string {
	var ret []string
	entries, err := ioutil.ReadDir("testdata")
	Expect(err).NotTo(HaveOccurred())
	for _, f := range entries {
		ret = append(ret, filepath.Base(f.Name()))
	}
	return ret
}
func GetInputMeshServices(f string) []*smh_discovery.MeshService {
	var svc []*smh_discovery.MeshService
	GetInput(f, "ms", &svc)
	return svc
}
func GetOutputMeshServices(f string) []*smh_discovery.MeshService {
	var svc []*smh_discovery.MeshService
	GetOutput(f, "ms", &svc)
	return svc
}
func GetInputTrafficPolicies(f string) []*smh_networking.TrafficPolicy {
	var svc []*smh_networking.TrafficPolicy
	GetInput(f, "tp", &svc)
	return svc
}
func GetOutputTrafficPolicies(f string) []*smh_networking.TrafficPolicy {
	var tps []*smh_networking.TrafficPolicy
	GetOutput(f, "tp", &tps)
	return tps
}
func GetOutputDestinationRules(f string) []*istio_client_networking_types.DestinationRule {
	var drs []*istio_client_networking_types.DestinationRule
	GetOutput(f, "dr", &drs)
	for _, dr := range drs {
		dr.TypeMeta = k8s_meta_types.TypeMeta{}
	}
	return drs
}

func GetOutputVirtualServices(f string) []*istio_client_networking_types.VirtualService {
	var vss []*istio_client_networking_types.VirtualService
	GetOutput(f, "vs", &vss)

	// clean up type meta, as it doesn't show up in translation since its embedded in the type
	for _, vs := range vss {
		vs.TypeMeta = k8s_meta_types.TypeMeta{}
	}
	return vss
}

func GetInput(f, suffix string, out interface{}) {
	reader, err := os.Open("testdata/" + f + "/input-" + suffix + ".yaml")
	Expect(err).NotTo(HaveOccurred())
	defer reader.Close()
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 1024)
	err = decoder.Decode(out)
	Expect(err).NotTo(HaveOccurred())
}
func GetOutput(f, suffix string, out interface{}) {
	reader, err := os.Open("testdata/" + f + "/output-" + suffix + ".yaml")
	if err == nil {
		defer reader.Close()
		decoder := yaml.NewYAMLOrJSONDecoder(reader, 1024)
		err = decoder.Decode(out)
		Expect(err).NotTo(HaveOccurred())
	}
}
