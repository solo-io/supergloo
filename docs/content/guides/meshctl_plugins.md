---
title: Extending Meshctl
menuTitle: Extending Meshctl
weight: 120
---

{{% notice note %}}
Gloo Mesh Enterprise is required for this feature.
{{% /notice %}}

In this guide we will walk you through how to get up and running with meshctl plugins. Meshctl plugins extend the core features of meshctl for advanced Gloo Mesh Enterprise use cases such as building and deploying Wasm filters (TODO Link) to Istio workloads and fetching access logs from services in your mesh. To utilize these plugins, we'll:

1. Initialize the meshctl plugin manager
1. Search for available plugins
1. Install the meshctl-wasm plugin

## Before you begin

All you need to get started is meshctl 1.0.0 or greater. If you haven't already, you can install or upgrade meshctl with the following:

```shell
curl -sL http://run.solo.io/meshctl/install | sh -
```

## Initialize the plugin manager

TODO `meshctl init-plugin-manager`

TODO `meshctl plugin --help` 

## Search for available plugins

`meshctl plugin update` updates the plugin index with all available plugins

`meshctl search` with no args will show all available

`meshctl plugin info wasm` will show some details about the wasm plugin

## Install the Wasm plugin for meshctl

`meshctl plugin install wasm` will install the latest stable

`meshctl wasm --help` will show the available commands

`meshctl plugin upgrade wasm@v1.0.0-beta5` will upgrade to a specific tag

`meshctl plugin uninstall wasm` will remove the plugin

## Next Steps

TODO now that we've installed the Wasm plugin for meshctl, let's develop and deploy our first Wasm filter (link)