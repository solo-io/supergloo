package templates

import (
	"text/template"
)

var ResourceClientTemplate = template.Must(template.New("resource_reconciler").Funcs(Funcs).Parse(`package {{ .Project.ProjectConfig.Version }}

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type {{ .Name }}Client interface {
	BaseClient() clients.ResourceClient
	Register() error
{{- if .ClusterScoped }}
	Read(name string, opts clients.ReadOpts) (*{{ .Name }}, error)
{{- else }}
	Read(namespace, name string, opts clients.ReadOpts) (*{{ .Name }}, error)
{{- end }}
	Write(resource *{{ .Name }}, opts clients.WriteOpts) (*{{ .Name }}, error)
{{- if .ClusterScoped }}
	Delete(name string, opts clients.DeleteOpts) error
	List(opts clients.ListOpts) ({{ .Name }}List, error)
	Watch(opts clients.WatchOpts) (<-chan {{ .Name }}List, <-chan error, error)
{{- else }}
	Delete(namespace, name string, opts clients.DeleteOpts) error
	List(namespace string, opts clients.ListOpts) ({{ .Name }}List, error)
	Watch(namespace string, opts clients.WatchOpts) (<-chan {{ .Name }}List, <-chan error, error)
{{- end }}
}

type {{ lower_camel .Name }}Client struct {
	rc clients.ResourceClient
}

func New{{ .Name }}Client(rcFactory factory.ResourceClientFactory) ({{ .Name }}Client, error) {
	return New{{ .Name }}ClientWithToken(rcFactory, "")
}

func New{{ .Name }}ClientWithToken(rcFactory factory.ResourceClientFactory, token string) ({{ .Name }}Client, error) {
	rc, err := rcFactory.NewResourceClient(factory.NewResourceClientParams{
		ResourceType: &{{ .Name }}{},
		Token: token,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating base {{ .Name }} resource client")
	}
	return New{{ .Name }}ClientWithBase(rc), nil
}

func New{{ .Name }}ClientWithBase(rc clients.ResourceClient) {{ .Name }}Client {
	return &{{ lower_camel .Name }}Client{
		rc: rc,
	}
}

func (client *{{ lower_camel .Name }}Client) BaseClient() clients.ResourceClient {
	return client.rc
}

func (client *{{ lower_camel .Name }}Client) Register() error {
	return client.rc.Register()
}

{{ if .ClusterScoped }}
func (client *{{ lower_camel .Name }}Client) Read(name string, opts clients.ReadOpts) (*{{ .Name }}, error) {
{{- else }}
func (client *{{ lower_camel .Name }}Client) Read(namespace, name string, opts clients.ReadOpts) (*{{ .Name }}, error) {
{{- end }}
	opts = opts.WithDefaults()
{{ if .ClusterScoped }}
	resource, err := client.rc.Read("", name, opts)
{{- else }}
	resource, err := client.rc.Read(namespace, name, opts)
{{- end }}
	if err != nil {
		return nil, err
	}
	return resource.(*{{ .Name }}), nil
}

func (client *{{ lower_camel .Name }}Client) Write({{ lower_camel .Name }} *{{ .Name }}, opts clients.WriteOpts) (*{{ .Name }}, error) {
	opts = opts.WithDefaults()
	resource, err := client.rc.Write({{ lower_camel .Name }}, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*{{ .Name }}), nil
}

{{ if .ClusterScoped }}
func (client *{{ lower_camel .Name }}Client) Delete(name string, opts clients.DeleteOpts) error {
{{- else }}
func (client *{{ lower_camel .Name }}Client) Delete(namespace, name string, opts clients.DeleteOpts) error {
{{- end }}
	opts = opts.WithDefaults()
{{ if .ClusterScoped }}
	return client.rc.Delete("", name, opts)
{{- else }}
	return client.rc.Delete(namespace, name, opts)
{{- end }}
}

{{ if .ClusterScoped }}
func (client *{{ lower_camel .Name }}Client) List(opts clients.ListOpts) ({{ .Name }}List, error) {
{{- else }}
func (client *{{ lower_camel .Name }}Client) List(namespace string, opts clients.ListOpts) ({{ .Name }}List, error) {
{{- end }}
	opts = opts.WithDefaults()
{{ if .ClusterScoped }}
	resourceList, err := client.rc.List("", opts)
{{- else }}
	resourceList, err := client.rc.List(namespace, opts)
{{- end }}
	if err != nil {
		return nil, err
	}
	return convertTo{{ .Name }}(resourceList), nil
}

{{ if .ClusterScoped }}
func (client *{{ lower_camel .Name }}Client) Watch(opts clients.WatchOpts) (<-chan {{ .Name }}List, <-chan error, error) {
{{- else }}
func (client *{{ lower_camel .Name }}Client) Watch(namespace string, opts clients.WatchOpts) (<-chan {{ .Name }}List, <-chan error, error) {
{{- end }}
	opts = opts.WithDefaults()
{{ if .ClusterScoped }}
	resourcesChan, errs, initErr := client.rc.Watch("", opts)
{{- else }}
	resourcesChan, errs, initErr := client.rc.Watch(namespace, opts)
{{- end }}
	if initErr != nil {
		return nil, nil, initErr
	}
	{{ lower_camel .PluralName }}Chan := make(chan {{ .Name }}List)
	go func() {
		for {
			select {
			case resourceList := <-resourcesChan:
				{{ lower_camel .PluralName }}Chan <- convertTo{{ .Name }}(resourceList)
			case <-opts.Ctx.Done():
				close({{ lower_camel .PluralName }}Chan)
				return
			}
		}
	}()
	return {{ lower_camel .PluralName }}Chan, errs, nil
}

func convertTo{{ .Name }}(resources resources.ResourceList) {{ .Name }}List {
	var {{ lower_camel .Name }}List {{ .Name }}List
	for _, resource := range resources {
		{{ lower_camel .Name }}List = append({{ lower_camel .Name }}List, resource.(*{{ .Name }}))
	}
	return {{ lower_camel .Name }}List
}

`))
