package templates

import (
	"text/template"
)

var ResourceClientTestTemplate = template.Must(template.New("resource_client_test").Funcs(Funcs).Parse(`// +build solokit

package {{ .Project.ProjectConfig.Version }}

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/tests/typed"
)

var _ = Describe("{{ .Name }}Client", func() {
{{- if (not .ClusterScoped) }}
	var (
		namespace string
	)
{{- end }}
	for _, test := range []typed.ResourceClientTester{
		&typed.KubeRcTester{Crd: {{ .Name }}Crd},
{{- /* cluster-scoped resources are currently only supported by crd client */}}
{{- if (not .ClusterScoped) }}
		&typed.ConsulRcTester{},
		&typed.FileRcTester{},
		&typed.MemoryRcTester{},
		&typed.VaultRcTester{},
		&typed.KubeSecretRcTester{},
		&typed.KubeConfigMapRcTester{},
{{- end }}
	} {
		Context("resource client backed by "+test.Description(), func() {
			var (
				client {{ .Name }}Client
				err    error
				name1, name2, name3 = "foo"+helpers.RandString(3), "boo"+helpers.RandString(3), "goo"+helpers.RandString(3)
			)


{{- if .ClusterScoped }}
{{/* cluster-scoped resources get no namespace, must delete individual resources*/}}
			BeforeEach(func() {
				factory := test.Setup("")
				client, err = NewClusterResourceClient(factory)
				Expect(err).NotTo(HaveOccurred())
			})
			AfterEach(func() {
				client.Delete(name1, clients.DeleteOpts{})
				client.Delete(name2, clients.DeleteOpts{})
				client.Delete(name3, clients.DeleteOpts{})
			})
			It("CRUDs {{ .Name }}s "+test.Description(), func() {
				{{ .Name }}ClientTest(client, name1, name2, name3)
			})
{{- else }}
{{/* non-cluster-scoped resources get a namespace and then the ns is deleted*/}}
			BeforeEach(func() {
				namespace = helpers.RandString(6)
				factory := test.Setup(namespace)
				client, err = New{{ .Name }}Client(factory)
				Expect(err).NotTo(HaveOccurred())
			})
			AfterEach(func() {
				test.Teardown(namespace)
			})
			It("CRUDs {{ .Name }}s "+test.Description(), func() {
				{{ .Name }}ClientTest(namespace, client, name1, name2, name3)
			})
{{- end }}
		})
	}
})


{{- if .ClusterScoped }}
func {{ .Name }}ClientTest(client {{ .Name }}Client, name1, name2, name3 string) {
{{- else }}
func {{ .Name }}ClientTest(namespace string, client {{ .Name }}Client, name1, name2, name3 string) {
{{- end }}
	err := client.Register()
	Expect(err).NotTo(HaveOccurred())

	name := name1

{{- if .ClusterScoped }}
	input := New{{ .Name }}("", name)
{{- else }}
	input := New{{ .Name }}(namespace, name)
	input.Metadata.Namespace = namespace
{{- end }}
	r1, err := client.Write(input, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())

	_, err = client.Write(input, clients.WriteOpts{})
	Expect(err).To(HaveOccurred())
	Expect(errors.IsExist(err)).To(BeTrue())

	Expect(r1).To(BeAssignableToTypeOf(&{{ .Name }}{}))
	Expect(r1.GetMetadata().Name).To(Equal(name))

{{- if (not .ClusterScoped) }}
	Expect(r1.GetMetadata().Namespace).To(Equal(namespace))
{{- end }}
	Expect(r1.Metadata.ResourceVersion).NotTo(Equal(input.Metadata.ResourceVersion))
	Expect(r1.Metadata.Ref()).To(Equal(input.Metadata.Ref()))
	{{- range .Fields }}
		{{- if and (not (eq .Name "metadata")) (not .IsOneof) }}
	Expect(r1.{{ upper_camel .Name }}).To(Equal(input.{{ upper_camel .Name }}))
		{{- end }}
	{{- end }}

	_, err = client.Write(input, clients.WriteOpts{
		OverwriteExisting: true,
	})
	Expect(err).To(HaveOccurred())

	input.Metadata.ResourceVersion = r1.GetMetadata().ResourceVersion
	r1, err = client.Write(input, clients.WriteOpts{
		OverwriteExisting: true,
	})
	Expect(err).NotTo(HaveOccurred())


{{- if .ClusterScoped }}
	read, err := client.Read(name, clients.ReadOpts{})
{{- else }}
	read, err := client.Read(namespace, name, clients.ReadOpts{})
{{- end }}
	Expect(err).NotTo(HaveOccurred())
	Expect(read).To(Equal(r1))


{{- if (not .ClusterScoped) }}
	_, err = client.Read("doesntexist", name, clients.ReadOpts{})
	Expect(err).To(HaveOccurred())
	Expect(errors.IsNotExist(err)).To(BeTrue())
{{- end }}

	name = name2
	input = &{{ .Name }}{}

	input.Metadata = core.Metadata{
		Name:      name,
{{- if (not .ClusterScoped) }}
		Namespace: namespace,
{{- end }}
	}

	r2, err := client.Write(input, clients.WriteOpts{})
	Expect(err).NotTo(HaveOccurred())


{{- if .ClusterScoped }}
	list, err := client.List(clients.ListOpts{})
{{- else }}
	list, err := client.List(namespace, clients.ListOpts{})
{{- end }}
	Expect(err).NotTo(HaveOccurred())
	Expect(list).To(ContainElement(r1))
	Expect(list).To(ContainElement(r2))


{{- if .ClusterScoped }}
	err = client.Delete("adsfw", clients.DeleteOpts{})
{{- else }}
	err = client.Delete(namespace, "adsfw", clients.DeleteOpts{})
{{- end }}
	Expect(err).To(HaveOccurred())
	Expect(errors.IsNotExist(err)).To(BeTrue())


{{- if .ClusterScoped }}
	err = client.Delete("adsfw", clients.DeleteOpts{
{{- else }}
	err = client.Delete(namespace, "adsfw", clients.DeleteOpts{
{{- end }}
		IgnoreNotExist: true,
	})
	Expect(err).NotTo(HaveOccurred())


{{- if .ClusterScoped }}
	err = client.Delete(r2.GetMetadata().Name, clients.DeleteOpts{})
{{- else }}
	err = client.Delete(namespace, r2.GetMetadata().Name, clients.DeleteOpts{})
{{- end }}
	Expect(err).NotTo(HaveOccurred())

	Eventually(func() {{ .Name }}List {
{{- if .ClusterScoped }}
		list, err = client.List(clients.ListOpts{})
{{- else }}
		list, err = client.List(namespace, clients.ListOpts{})
{{- end }}
		Expect(err).NotTo(HaveOccurred())
		return list
	}, time.Second * 10).Should(ContainElement(r1))
	Eventually(func() {{ .Name }}List {
{{- if .ClusterScoped }}
		list, err = client.List(clients.ListOpts{})
{{- else }}
		list, err = client.List(namespace, clients.ListOpts{})
{{- end }}
		Expect(err).NotTo(HaveOccurred())
		return list
	}, time.Second * 10).ShouldNot(ContainElement(r2))

{{- if .ClusterScoped }}
	w, errs, err := client.Watch(clients.WatchOpts{
{{- else }}
	w, errs, err := client.Watch(namespace, clients.WatchOpts{
{{- end }}
		RefreshRate: time.Hour,
	})
	Expect(err).NotTo(HaveOccurred())

	var r3 resources.Resource
	wait := make(chan struct{})
	go func() {
		defer close(wait)
		defer GinkgoRecover()

		resources.UpdateMetadata(r2, func(meta *core.Metadata) {
			meta.ResourceVersion = ""
		})
		r2, err = client.Write(r2, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())

		name = name3
		input = &{{ .Name }}{}
		Expect(err).NotTo(HaveOccurred())
		input.Metadata = core.Metadata{
			Name:      name,
{{- if (not .ClusterScoped) }}
			Namespace: namespace,
{{- end }}
		}

		r3, err = client.Write(input, clients.WriteOpts{})
		Expect(err).NotTo(HaveOccurred())
	}()
	<-wait

	select {
	case err := <-errs:
		Expect(err).NotTo(HaveOccurred())
	case list = <-w:
	case <-time.After(time.Millisecond * 5):
		Fail("expected a message in channel")
	}

drain:
	for {
		select {
		case list = <-w:
		case err := <-errs:
			Expect(err).NotTo(HaveOccurred())
		case <-time.After(time.Millisecond * 500):
			break drain
		}
	}

	Expect(list).To(ContainElement(r1))
	Expect(list).To(ContainElement(r2))
	Expect(list).To(ContainElement(r3))
}
`))
