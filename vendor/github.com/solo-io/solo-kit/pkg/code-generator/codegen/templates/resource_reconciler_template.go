package templates

import (
	"text/template"
)

var ResourceReconcilerTemplate = template.Must(template.New("resource_client").Funcs(Funcs).Parse(`package {{ .Project.ProjectConfig.Version }}
import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/reconcile"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/utils/contextutils"
)

// Option to copy anything from the original to the desired before writing. Return value of false means don't update
type Transition{{ .Name }}Func func(original, desired *{{ .Name }}) (bool, error)

type {{ .Name }}Reconciler interface {
	Reconcile(namespace string, desiredResources {{ .Name }}List, transition Transition{{ .Name }}Func, opts clients.ListOpts) error
}

func {{ lower_camel .Name }}sToResources(list {{ .Name }}List) resources.ResourceList {
	var resourceList resources.ResourceList
	for _, {{ lower_camel .Name }} := range list {
		resourceList = append(resourceList, {{ lower_camel .Name }})
	}
	return resourceList
}

func New{{ .Name }}Reconciler(client {{ .Name }}Client) {{ .Name }}Reconciler {
	return &{{ lower_camel .Name }}Reconciler{
		base: reconcile.NewReconciler(client.BaseClient()),
	}
}

type {{ lower_camel .Name }}Reconciler struct {
	base reconcile.Reconciler
}

func (r *{{ lower_camel .Name }}Reconciler) Reconcile(namespace string, desiredResources {{ .Name }}List, transition Transition{{ .Name }}Func, opts clients.ListOpts) error {
	opts = opts.WithDefaults()
	opts.Ctx = contextutils.WithLogger(opts.Ctx, "{{ lower_camel .Name }}_reconciler")
	var transitionResources reconcile.TransitionResourcesFunc
	if transition != nil {
		transitionResources = func(original, desired resources.Resource) (bool, error) {
			return transition(original.(*{{ .Name }}), desired.(*{{ .Name }}))
		}
	}
	return r.base.Reconcile(namespace, {{ lower_camel .Name }}sToResources(desiredResources), transitionResources, opts)
}
`))
