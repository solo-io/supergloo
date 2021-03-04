---
title: Extending Meshctl
menuTitle: Extending Meshctl
weight: 120
---

{{% notice note %}}
Gloo Mesh Enterprise is required for this feature.
{{% /notice %}}

In this guide we will walk you through how to get up and running with meshctl plugins. Meshctl plugins extend the core features of meshctl for advanced Gloo Mesh Enterprise use cases such as building and deploying [Wasm filters]({{% versioned_link_path fromRoot="/guides/wasm_extension" %}}) to Istio workloads and fetching access logs from services in your mesh. To utilize these plugins, we will follow these steps:

1. Initialize the meshctl plugin manager
1. Search for available plugins
1. Install the meshctl-wasm plugin
1. Upgrade and uninstall the meshctl-wasm plugin (optional)

## Before you begin

All you need to get started is meshctl 1.0.0 or greater. If you haven't already, you can install or upgrade meshctl with the following:

```shell
curl -sL http://run.solo.io/meshctl/install | sh -
```

## Initialize the plugin manager

First we will install the Gloo Mesh Enterprise CLI plugin manager by running the following command:

```bash
meshctl init-plugin-manager
```

You should see the following output:

```console
The meshctl plugin manager was successfully installed 🎉
```

The next step is to add the path the to `plugin` binaries to your $PATH environment variable. 

```bash
export PATH=$HOME/.gloo-mesh/bin:$PATH
```

The path `$HOME/.gloo-mesh/bin` includes a `meshctl-plugin` symlink to the plugin binary, and will also include symlinks to any plugins installed by meshctl. The actual binary for each plugin is at the path `$HOME/.gloo-mesh/store/<plugin_name>/<plugin_version>/`.

Now we can view all the plugin options by running:

```bash
meshctl plugin --help
```

```console
The interface for managing meshctl plugins

Usage:
  plugin [command]

Available Commands:
  help        Help about any command
  index       Manage the repositories used to install plugins from
  info        Show detailed information about a plugin
  install     Install one or more meshctl plugins
  list        List installed plugins and their versions
  search      Search for plugins where the name matches the pattern
  uninstall   Uninstall one or more meshctl plugins
  update      Update the local copy of plugin indexes
  upgrade     Upgrade installed plugins to newer versions.

Flags:
  -h, --help      help for plugin
  -v, --verbose   Show verbose logging information

Use "plugin [command] --help" for more information about a command.
```

Next we will find available plugins we can use from repositories included in our index.

## Search for available plugins

Meshctl plugins are available on repositories, and these repositories are held in an index. You can see the current list of repositories by using the `meshctl plugin index list` command.

```bash
meshctl plugin index list
```

```console
INDEX    URL
default  https://github.com/solo-io/meshctl-plugin-index.git
```

Right now we have a single repository available that was added during the plugin initialization process. Let's update the available plugins from that repository by running `meshctl plugin update`.

```bash
meshctl plugin update
```

```console
INDEX    STATUS
default  already up-to-date
```

With a fully up to date list of repositories, we can get all available plugins by running `meshctl plugin search` with no arguments:

```bash
meshctl plugin search
```

```console
NAME       DESCRIPTION
accesslog  View access logs collected by Gloo Mesh
plugin     The plugin management system for meshctl
wasm       Utility to manage Gloo Mesh wasm filter deployments
```

Let's take a closer look at the wasm plugin by using the `meshctl plugin info` command:

```bash
meshctl plugin info wasm
```

```console
NAME: wasm
DESCRIPTION: Utility to manage Gloo Mesh wasm filter deployments

VERSIONS:
v1.0.0-beta5 (latest)
\
 | REQUIREMENTS:
 | istio >= 1.8
/
v1.0.0-beta4
\
 | REQUIREMENTS:
 | istio >= 1.8
/
v1.0.0-beta3
\
 | REQUIREMENTS:
 | istio >= 1.8
/
v0.5.1
\
 | REQUIREMENTS:
 | istio >= 1.8
/
```

The output shows all available versions of the plugin. In the next section we will install and explore the latest version of the wasm plugin.

## Install the Wasm plugin for meshctl

In the previous section we explored the current list of plugins and viewed the versions of the wasm plugin. We can now install the wasm plugin using the `meshctl plugin install` command. The primary argument of the command will be the path to the plugin, an `@` symbol, and the version of the plugin to install. If you do not specify a version, the latest stable version will be installed. Let's try installing the latest stable wasm plugin:

```bash
meshctl plugin install wasm
```

```console
Installing plugin: wasm
Downloading binary 100% [========================================] (70/70 MB, 9.982 MB/s)             
Installed plugin: wasm to bin/meshctl-wasm
```

We can view the installed version by using the `list` command:

```bash
meshctl plugin list
```

```console
NAME    VERSION
plugin  v0.5.1
wasm    v0.5.1
```

It look like we have install version `v0.5.1`, which lines up with what we saw in the previous section. Installing a plugin makes it available as a command from the `meshctl` binary. We can invoke the wasm plugin by running `meshctl wasm`. To view the available subcommands for the wasm plugin, we can use the `--help` flag.

```bash
meshctl wasm --help
```

```console
The interface for managing Gloo Mesh WASM filters

Usage:
  wasm [command]

Available Commands:
  build       Build a wasm image from the filter source directory.
  deploy      Deploy an Envoy WASM Filter to Istio Sidecar Proxies (Envoy).
  help        Help about any command
  init        Initialize a project directory for a new Envoy WASM Filter.
  list        List Envoy WASM Filters stored locally or published to webassemblyhub.io.
  login       Log in so you can push images to the remote server.
  pull        Pull wasm filters from remote registry
  push        Push a wasm filter to remote registry
  tag         Create a tag TARGET_IMAGE that refers to SOURCE_IMAGE
  undeploy    Remove an Envoy WASM Filter deployment
  version     Display the version of meshctl wasm

Flags:
  -h, --help   help for wasm

Use "wasm [command] --help" for more information about a command.
```

At this point we are ready to use our wasm plugin to create wasm extensions.

## Upgrading or Deleting a Plugin

If we want to upgrade the version of a plugin, we can use the `upgrade` subcommand along with plugin path and version. For instance, let's try upgrading our wasm plugin to the beta version `v1.0.0-beta5` found in the available versions list:

```bash
meshctl plugin upgrade wasm@v1.0.0-beta5
```

```console
Upgrading plugin: wasm
Downloading binary 100% [========================================] (73/73 MB, 9.883 MB/s)             
Plugin upgraded: wasm
PLUGIN  STATUS
wasm    v0.5.1 -> v1.0.0-beta5
```

The upgrade process will delete the previous version of the plugin from `.gloo-mesh/store/` and update the symlink in `.gloo-mesh/bin/` to the new plugin binary.

If we are done using a plugin and want to remove it, we can use the `uninstall` subcommand along with the plugin name to remove. Here's an example of removing the wasm plugin. You may not want to run this command if you plan to follow the [Wasm filter development and deployment guide]({{% versioned_link_path fromRoot="/guides/wasm_meshctl" %}}) next.

```bash
meshctl plugin uninstall wasm
```

```console
Uninstalling wasm
Uninstalled plugin
```

The plugin has been uninstalled, the symlink removed from the `$HOME/.gloo-mesh/bin` directory, and the binary has been removed as well.

## Next Steps

Now that we've installed the Wasm plugin for meshctl, let's develop and deploy our [first Wasm filter]({{% versioned_link_path fromRoot="/guides/wasm_meshctl" %}})