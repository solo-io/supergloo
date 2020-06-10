package fakes

import (
	linkerd_config "github.com/linkerd/linkerd2/controller/gen/apis/serviceprofile/v1alpha2"
	"github.com/onsi/gomega"
	smi_config "github.com/servicemeshinterface/smi-sdk-go/pkg/apis/split/v1alpha1"
	smh_discovery "github.com/solo-io/service-mesh-hub/pkg/api/discovery.smh.solo.io/v1alpha1"
	smh_networking "github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/v1alpha1"
	security_v1alpha1 "github.com/solo-io/service-mesh-hub/pkg/api/security.smh.solo.io/v1alpha1"
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
		smh_networking.AddToScheme,
		smh_discovery.AddToScheme,
	}).AddToScheme(s)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())

	return fake.NewFakeClientWithScheme(s, objs...)
}
