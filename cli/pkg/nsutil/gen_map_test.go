package nsutil

import (
	"github.com/solo-io/supergloo/cli/pkg/cmd/options"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Generate Options", func() {
	nsrMap := make(options.NsResourceMap)
	nsrMap["ns1"] = &options.NsResource{
		Meshes:  []string{"m1", "m2"},
		Secrets: []string{"s1", "s2"},
		MeshesByInstallNs: []options.ResourceRef{
			options.ResourceRef{
				Name:      "m1",
				Namespace: "ns1",
			},
		},
	}
	nsrMap["ns2"] = &options.NsResource{
		Meshes:  []string{},
		Secrets: []string{"s3"},
		MeshesByInstallNs: []options.ResourceRef{
			options.ResourceRef{
				Name:      "m2",
				Namespace: "ns1",
			},
		},
	}

	It("should create the correct Mesh options and map", func() {
		// Note that mesh "m2" is installed in ns2 but its CRD is in ns1
		genOpts, resMap := generateMeshSelectOptions(nsrMap)
		Expect(genOpts).To(Equal([]string{
			"ns1, m1",
			"ns2, m2",
		}))
		expectedMap := make(ResMap)
		expectedMap["ns1, m1"] = ResSelect{
			displayName:      "m1",
			displayNamespace: "ns1",
			resourceRef: options.ResourceRef{
				Name:      "m1",
				Namespace: "ns1",
			},
		}
		expectedMap["ns2, m2"] = ResSelect{
			displayName:      "m2",
			displayNamespace: "ns2",
			resourceRef: options.ResourceRef{
				Name:      "m2",
				Namespace: "ns1",
			},
		}
		Expect(resMap).To(Equal(expectedMap))
	})

	It("should create the correct Secret options and map", func() {
		genOpts, resMap := generateCommonResourceSelectOptions("secret", nsrMap)
		Expect(genOpts).To(Equal([]string{
			"ns1, s1",
			"ns1, s2",
			"ns2, s3",
		}))
		expectedMap := make(ResMap)
		expectedMap["ns1, s1"] = ResSelect{
			displayName:      "s1",
			displayNamespace: "ns1",
			resourceRef: options.ResourceRef{
				Name:      "s1",
				Namespace: "ns1",
			},
		}
		expectedMap["ns1, s2"] = ResSelect{
			displayName:      "s2",
			displayNamespace: "ns1",
			resourceRef: options.ResourceRef{
				Name:      "s2",
				Namespace: "ns1",
			},
		}
		expectedMap["ns2, s3"] = ResSelect{
			displayName:      "s3",
			displayNamespace: "ns2",
			resourceRef: options.ResourceRef{
				Name:      "s3",
				Namespace: "ns2",
			},
		}
		Expect(resMap).To(Equal(expectedMap))
	})
})
