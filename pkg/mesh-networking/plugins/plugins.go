package plugins

import (
	"github.com/solo-io/service-mesh-hub/pkg/api/networking.smh.solo.io/snapshot/input"
	"github.com/solo-io/smh/pkg/mesh-networking/translation/utils/hostutils"
)

// parameters for initializing plugins
type Parameters struct {
	ClusterDomains hostutils.ClusterDomainRegistry
	Snapshot       input.Snapshot
}

func Register(constructor Constructor) {
	registeredPlugins = append(registeredPlugins, constructor)
}

// Note: Translator plugins should be added here by the plugin in the init() function.
var registeredPlugins []Constructor

type Constructor func(params Parameters) Plugin

func makePlugins(params Parameters) []Plugin {
	var plugins []Plugin
	for _, pluginFactory := range registeredPlugins {
		plugin := pluginFactory(params)
		plugins = append(plugins, plugin)
	}
	return plugins
}

// the plugin Factory initializes Translator plugins on each reconcile
type Factory interface {
	// return a set of plugins built from the given snapshot.
	MakePlugins(params Parameters) []Plugin
}

type factory struct{}

func NewFactory() Factory {
	return &factory{}
}

func (f *factory) MakePlugins(params Parameters) []Plugin {
	return makePlugins(params)
}

// Plugins modify the output VirtualService corresponding to the input MeshService.
type Plugin interface {
	// unique identifier for plugin
	PluginName() string
}

type RegisterField func(fieldPtr, val interface{}) error
