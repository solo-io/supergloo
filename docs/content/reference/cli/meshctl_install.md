---
title: "meshctl install"
weight: 5
---
## meshctl install

Install Service Mesh Hub

### Synopsis

Install Service Mesh Hub

```
meshctl install [flags]
```

### Options

```
      --cluster-name string   Name by which to register the management-plane cluster in Service Mesh Hub. This flag will only be considered if --register is set. (default "management-plane")
      --create-namespace      Create the namespace to install Service Mesh Hub into (default true)
  -d, --dry-run               Send the raw installation yaml to stdout instead of applying it to kubernetes
  -f, --file string           Install Service Mesh Hub from this Helm chart archive file rather than from a release
  -h, --help                  help for install
  -r, --register              Register the management plane cluster. This would be the same as running the meshctl cluster register command on the management plane cluster after installing.
      --release-name string   Helm release name (default "service-mesh-hub")
      --values strings        List of files with value overrides for the Service Mesh Hub Helm chart, (e.g. --values file1,file2 or --values file1 --values file2)
      --version string        Version to install (e.g. v1.2.0, defaults to latest)
```

### Options inherited from parent commands

```
      --context string          Specify which context from the kubeconfig should be used; uses current context if none is specified
      --kube-timeout duration   Specify the default timeout for requests to kubernetes API servers (default 5s)
      --kubeconfig string       Specify the kubeconfig for the current command
  -n, --namespace string        Specify the namespace where Service Mesh Hub resources should be written (default "service-mesh-hub")
  -v, --verbose                 Enable verbose mode, which outputs additional execution details that may be helpful for debugging
```

### SEE ALSO

* [meshctl](../meshctl)	 - CLI for Service Mesh Hub

