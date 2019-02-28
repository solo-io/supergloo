package testutils

import (
	. "github.com/onsi/gomega"
	kubev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
	v1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	)

func WaitForIstioTeardown(ns string) {
	Eventually(func() []kubev1.Service {
		svcs, err := MustKubeClient().CoreV1().Services(ns).List(v1.ListOptions{})
		if err != nil {
			// namespace is gone
			return []kubev1.Service{}
		}
		return svcs.Items
	}, time.Second*30).Should(BeEmpty())

	Eventually(func() []v1beta1.CustomResourceDefinition{
		svcs, err := MustApiExtsClient().ApiextensionsV1beta1().CustomResourceDefinitions().List(v1.ListOptions{})
		if err != nil {
			// namespace is gone
			return []v1beta1.CustomResourceDefinition{}
		}
		var defs []v1beta1.CustomResourceDefinition
		for _, def := range svcs.Items {
			if strings.HasSuffix(def.Name, ".istio.io") {
				defs = append(defs, def)
			}
		}
		return defs
	}, time.Second*30).Should(BeEmpty())
}
