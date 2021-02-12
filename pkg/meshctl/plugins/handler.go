package plugins

type Plugin interface {
	// Run takes in the arguments following the plugin name
	// and returns any error that occurred while running
	Run(args []string) error
}

type Handler interface {
	// Lookup takes the name of a plugin and returns a plugin
	// and a bool indicating if it was found
	Lookup(pluginName string) (Plugin, bool)
}

// Handle takes a handler and arguments in order to lookup and run a plugin.
func Handle(handler Handler, args []string) error {
	plugin, ok := handler.Lookup(args[0])
	if !ok {
		return nil
	}

	var pluginArgs []string
	if len(args) > 1 {
		pluginArgs = args[1:]
	}

	return plugin.Run(pluginArgs)
}
