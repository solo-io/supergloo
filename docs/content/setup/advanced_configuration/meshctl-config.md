---
title: Meshctl Config File
weight: 60
description: Persistent configuration for `meshctl`
---

## Config File

When certain `meshctl` commands are invoked, they will attempt to read a configuration yaml file located at `$HOME/.gloo-mesh/meshctl-config.yaml`. The location of this file can be overridden by setting the `--config` value when invoking `meshctl`.

This file can be configured using the `meshctl cluster configure` command, which opens up an
interactive prompt to save information about your management and data plane clusters.

The resulting file should look similar to:
```yaml
apiVersion: v1
clusters:
  cluster1: # data plane cluster
    kubeConfig: <home_directory>/.kube/config
    kubeContext: cluster1_context
  cluster2: # data plane cluster
    kubeConfig: <home_directory>/.kube/config
    kubeContext: cluster2_context
  managementPlane: # management cluster
    kubeConfig: <home_directory>/.kube/config
    kubeContext: mgmt_cluster_context
```

You can also configure this file non-interactively by using the --disable-prompt flag in `meshctl cluster configure`.