package templates

import (
	"text/template"
)

var XdsTemplate = template.Must(template.New("xds_template").Funcs(Funcs).Parse(`package {{ .Project.ProjectConfig.Version }}

import (
	"context"
	"errors"
	"fmt"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	discovery "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/client"
	"github.com/solo-io/solo-kit/pkg/api/v1/control-plane/server"
)

// Type Definitions:

const {{ upper_camel .MessageType }}Type = cache.TypePrefix + "/{{ .ProtoPackage }}.{{ upper_camel .MessageType }}"

/* Defined a resource - to be used by snapshot */
type {{ upper_camel .MessageType }}XdsResourceWrapper struct {
	// TODO(yuval-k): This is public for mitchellh hashstructure to work properly. consider better alternatives.
	Resource *{{ upper_camel .MessageType }}
}

// Make sure the Resource interface is implemented
var _ cache.Resource = &{{ upper_camel .MessageType }}XdsResourceWrapper{}

func New{{ upper_camel .MessageType }}XdsResourceWrapper(resourceProto *{{ upper_camel .MessageType }}) *{{ upper_camel .MessageType }}XdsResourceWrapper {
	return &{{ upper_camel .MessageType }}XdsResourceWrapper{
		Resource: resourceProto,
	}
}

func (e *{{ upper_camel .MessageType }}XdsResourceWrapper) Self() cache.XdsResourceReference {
	return cache.XdsResourceReference{Name: e.Resource.{{ upper_camel .NameField }}, Type: {{ upper_camel .MessageType }}Type}
}

func (e *{{ upper_camel .MessageType }}XdsResourceWrapper) ResourceProto() cache.ResourceProto {
	return e.Resource
}

{{- if .NoReferences }}
func (e *{{ upper_camel .MessageType }}XdsResourceWrapper) References() []cache.XdsResourceReference {
	return nil
}
{{- else }}
	// This method is not implemented as it requires domain knowledge and cannot be auto generated.
	// Please copy it, and implement it in a different file (so it doesn't get overwritten).
	// Alternativly, specify the annotation @solo-kit:resource.no_references in the comments for the 
	// {{ upper_camel .MessageType }} to indicate that there are no references.
	//	func (e *{{ upper_camel .MessageType }}XdsResourceWrapper) References() []cache.XdsResourceReference {
	//		panic("not implemented")
	//	}
{{- end }}

// Define a type record. This is used by the generic client library.
var {{ upper_camel .MessageType }}TypeRecord = client.NewTypeRecord(
	{{ upper_camel .MessageType }}Type,
	
	// Return an empty message, that can be used to deserialize bytes into it.
	func() cache.ResourceProto { return &{{ upper_camel .MessageType }}{} },
	
	// Covert the message to a resource suitable for use for protobuf's Any.
	func(r cache.ResourceProto) cache.Resource {
		return &{{ upper_camel .MessageType }}XdsResourceWrapper{Resource: r.(*{{ upper_camel .MessageType }})}
	},
)

// Server Implementation:

// Wrap the generic server and implement the type sepcific methods:
type {{ lower_camel .Name }}Server struct {
	server.Server
}

func New{{ upper_camel .Name }}Server(genericServer server.Server) {{ upper_camel .Name }}Server {
	return &{{ lower_camel .Name }}Server{Server: genericServer}
}

func (s *{{ lower_camel .Name }}Server) Stream{{ upper_camel .MessageType }}(stream {{ upper_camel .Name }}_Stream{{ upper_camel .MessageType }}Server) error {
	return s.Server.Stream(stream, {{ upper_camel .MessageType }}Type)
}

func (s *{{ lower_camel .Name }}Server) Fetch{{ upper_camel .MessageType }}(ctx context.Context, req *discovery.DiscoveryRequest) (*discovery.DiscoveryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unavailable, "empty request")
	}
	req.TypeUrl = {{ upper_camel .MessageType }}Type
	return s.Server.Fetch(ctx, req)
}

func (s *{{ lower_camel .Name }}Server) Incremental{{ upper_camel .MessageType }}(_ {{ upper_camel .Name }}_Incremental{{ upper_camel .MessageType }}Server) error {
	return errors.New("not implemented")
}


// Client Implementation: Generate a strongly typed client over the generic client

// The apply functions receives resources and returns an error if they were applied correctly.
// In theory the configuration can become valid in the future (i.e. eventually consistent), but I don't think we need to worry about that now
// As our current use cases only have one configuration resource, so no interactions are expected.
type Apply{{ upper_camel .MessageType }} func(version string, resources []*{{ upper_camel .MessageType }}) error

// Convert the strongly typed apply to a generic apply.
func apply{{ upper_camel .MessageType }}(typedApply Apply{{ upper_camel .MessageType }}) func(cache.Resources) error {
	return func(resources cache.Resources) error {

		var configs []*{{ upper_camel .MessageType }}
		for _, r := range resources.Items {
			if proto, ok := r.ResourceProto().(*{{ upper_camel .MessageType }}); !ok {
				return fmt.Errorf("resource %s of type %s incorrect", r.Self().Name, r.Self().Type)
			} else {
				configs = append(configs, proto)
			}
		}

		return typedApply(resources.Version, configs)
	}
}

func New{{ upper_camel .MessageType }}Client(nodeinfo *core.Node, typedApply Apply{{ upper_camel .MessageType }}) client.Client {
	return client.NewClient(nodeinfo, {{ upper_camel .MessageType }}TypeRecord, apply{{ upper_camel .MessageType }}(typedApply))
}

`))
