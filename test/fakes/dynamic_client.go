package fakes

import (
	linkerd_config "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	"github.com/onsi/gomega"
	smi_config "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha3"
	discovery_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/discovery.zephyr.solo.io/v1alpha1"
	networking_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/networking.zephyr.solo.io/v1alpha1"
	security_v1alpha1 "github.com/solo-io/mesh-projects/pkg/api/security.zephyr.solo.io/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// make an in-memory dynamic client.Client compatible with our types
func InMemoryClient(objs ...runtime.Object) client.Client {
	s := scheme.Scheme

	err := (&runtime.SchemeBuilder{
		linkerd_config.AddToScheme,
		smi_config.AddToScheme,
		security_v1alpha1.AddToScheme,
		networking_v1alpha1.AddToScheme,
		discovery_v1alpha1.AddToScheme,
	}).AddToScheme(s)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return fake.NewFakeClientWithScheme(s, objs...)
}
