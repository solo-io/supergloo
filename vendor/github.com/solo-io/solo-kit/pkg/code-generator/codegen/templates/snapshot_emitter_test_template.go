package templates

import (
	"text/template"
)

var ResourceGroupEmitterTestTemplate = template.Must(template.New("resource_group_emitter_test").Funcs(Funcs).Parse(`// +build solokit

package {{ .Project.ProjectConfig.Version }}

{{- /* we need to know if the tests require a crd client or a regular clientset */ -}}
{{- $clients := new_str_slice }}
{{- $needs_standard_client := false }}
{{- range .Resources}}
{{- $clients := (append_str_slice $clients (printf "%vClient"  (lower_camel .Name))) }}
{{- if (not .HasStatus) }}
{{- $needs_standard_client = true }}
{{- end}}
{{- end}}
{{- $clients := (join_str_slice $clients ", ") }}

import (
	"context"
	"os"
	"time"

	{{ .Imports }}
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/solo-kit/pkg/utils/log"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/test/helpers"
	"github.com/solo-io/solo-kit/test/setup"
	"github.com/solo-io/go-utils/kubeutils"
	kuberc "github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"k8s.io/client-go/rest"

	// Needed to run tests in GKE
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	// From https://github.com/kubernetes/client-go/blob/53c7adfd0294caa142d961e1f780f74081d5b15f/examples/out-of-cluster-client-configuration/main.go#L31
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
{{- if $needs_standard_client }}
	"k8s.io/client-go/kubernetes"
{{- end }}
)

var _ = Describe("{{ upper_camel .Project.ProjectConfig.Version }}Emitter", func() {
	if os.Getenv("RUN_KUBE_TESTS") != "1" {
		log.Printf("This test creates kubernetes resources and is disabled by default. To enable, set RUN_KUBE_TESTS=1 in your env.")
		return
	}
	var (
		namespace1          string
		namespace2          string
		name1, name2        = "angela"+helpers.RandString(3), "bob"+helpers.RandString(3)
		cfg                *rest.Config
		emitter            {{ .GoName }}Emitter
{{- range .Resources }}
		{{ lower_camel .Name }}Client {{ .ImportPrefix }}{{ .Name }}Client
{{- end}}
	)

	BeforeEach(func() {
		namespace1 = helpers.RandString(8)
		namespace2 = helpers.RandString(8)
		var err error
		cfg, err = kubeutils.GetConfig("", "")
		Expect(err).NotTo(HaveOccurred())
		err = setup.SetupKubeForTest(namespace1)
		Expect(err).NotTo(HaveOccurred())
		err = setup.SetupKubeForTest(namespace2)
		Expect(err).NotTo(HaveOccurred())


{{- if $needs_standard_client }}
		var kube kubernetes.Interface
{{- end}}


{{- range .Resources }}
		// {{ .Name }} Constructor
{{- if .HasStatus }}
		{{ lower_camel .Name }}ClientFactory := &factory.KubeResourceClientFactory{
			Crd: {{ .ImportPrefix }}{{ .Name }}Crd,
			Cfg: cfg,
		    SharedCache: kuberc.NewKubeCache(),
		}
{{- else }}
		kube, err = kubernetes.NewForConfig(cfg)
		Expect(err).NotTo(HaveOccurred())
{{/* TODO(ilackarms): Come with a way to specify that a resource is based on the secret client or configmap client*/}}
{{- if (eq .Name "Secret") }}
		{{ lower_camel .Name }}ClientFactory := &factory.KubeSecretClientFactory{
			Clientset: kube,
		}
{{- else }}
		{{ lower_camel .Name }}ClientFactory := &factory.KubeConfigMapClientFactory{
			Clientset: kube,
		}
{{- end }}
{{- end }}
		{{ lower_camel .Name }}Client, err = {{ .ImportPrefix }}New{{ .Name }}Client({{ lower_camel .Name }}ClientFactory)
		Expect(err).NotTo(HaveOccurred())
{{- end}}
		emitter = New{{ .GoName }}Emitter({{ $clients }})
	})
	AfterEach(func() {
		setup.TeardownKube(namespace1)
		setup.TeardownKube(namespace2)
{{- range .Resources }}
{{- if .ClusterScoped }}
		{{ lower_camel .Name }}Client.Delete(name1, clients.DeleteOpts{})
		{{ lower_camel .Name }}Client.Delete(name2, clients.DeleteOpts{})
{{- end }}
{{- end }}
	})
	It("tracks snapshots on changes to any resource", func() {
		ctx := context.Background()
		err := emitter.Register()
		Expect(err).NotTo(HaveOccurred())

		snapshots, errs, err := emitter.Snapshots([]string{namespace1, namespace2}, clients.WatchOpts{
			Ctx: ctx,
			RefreshRate: time.Second,
		})
		Expect(err).NotTo(HaveOccurred())

		var snap *{{ .GoName }}Snapshot
{{- range .Resources }}

		/*
			{{ .Name }}
		*/
		
		assertSnapshot{{ .PluralName }} := func(expect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List, unexpect{{ .PluralName }} {{ .ImportPrefix }}{{ .Name }}List) {
		drain:
			for {
				select {
				case snap = <-snapshots:
					for _, expected := range expect{{ .PluralName }} {
{{- if .ClusterScoped }}
						if _, err := snap.{{ .PluralName }}.Find(expected.Metadata.Ref().Strings()); err != nil {
{{- else }}
						if _, err := snap.{{ .PluralName }}.List().Find(expected.Metadata.Ref().Strings()); err != nil {
{{- end }}
							continue drain
						}
					}
					for _, unexpected := range unexpect{{ .PluralName }} {
{{- if .ClusterScoped }}
						if _, err := snap.{{ .PluralName }}.Find(unexpected.Metadata.Ref().Strings()); err == nil {
{{- else }}
						if _, err := snap.{{ .PluralName }}.List().Find(unexpected.Metadata.Ref().Strings()); err == nil {
{{- end }}
							continue drain
						}
					}
					break drain
				case err := <-errs:
					Expect(err).NotTo(HaveOccurred())
				case <-time.After(time.Second * 10):
{{- if .ClusterScoped }}
					nsList, _ := {{ lower_camel .Name }}Client.List(clients.ListOpts{})
					combined := nsList.ByNamespace()
{{- else }}
					nsList1, _ := {{ lower_camel .Name }}Client.List(namespace1, clients.ListOpts{})
					nsList2, _ := {{ lower_camel .Name }}Client.List(namespace2, clients.ListOpts{})
					combined := nsList1.ByNamespace()
					combined.Add(nsList2...)
{{- end }}
					Fail("expected final snapshot before 10 seconds. expected " + log.Sprintf("%v", combined))
				}
			}
		}	

{{- if .ClusterScoped }}
		{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }, nil)
{{- else }}
		{{ lower_camel .Name }}1a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}1b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name1), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }, nil)
{{- end }}

{{- if .ClusterScoped }}
		{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}2a }, nil)
{{- else }}
		{{ lower_camel .Name }}2a, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace1, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		{{ lower_camel .Name }}2b, err := {{ lower_camel .Name }}Client.Write({{ .ImportPrefix }}New{{ .Name }}(namespace2, name2), clients.WriteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b,  {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b  }, nil)
{{- end }}

{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a }, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a })
{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2a.Metadata.Namespace, {{ lower_camel .Name }}2a.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}2b.Metadata.Namespace, {{ lower_camel .Name }}2b.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}({{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b }, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b })
{{- end }}

{{- if .ClusterScoped }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}(nil, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}2a })
{{- else }}

		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1a.Metadata.Namespace, {{ lower_camel .Name }}1a.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())
		err = {{ lower_camel .Name }}Client.Delete({{ lower_camel .Name }}1b.Metadata.Namespace, {{ lower_camel .Name }}1b.Metadata.Name, clients.DeleteOpts{Ctx: ctx})
		Expect(err).NotTo(HaveOccurred())

		assertSnapshot{{ .PluralName }}(nil, {{ .ImportPrefix }}{{ .Name }}List{ {{ lower_camel .Name }}1a, {{ lower_camel .Name }}1b, {{ lower_camel .Name }}2a, {{ lower_camel .Name }}2b })
{{- end }}
{{- end}}
	})
})

`))
