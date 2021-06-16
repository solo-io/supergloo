---
title: "meshctl debug report"
weight: 5
---
## meshctl debug report

meshctl debug report selectively captures cluster information and logs into an archive to help diagnose problems.

### Synopsis

meshctl debug report selectively captures cluster information and logs into an archive to help diagnose problems.
Proxy logs can be filtered using:
  --include|--exclude ns1,ns2.../dep1,dep2.../pod1,pod2.../cntr1,cntr.../lbl1=val1,lbl2=val2.../ann1=val1,ann2=val2...
where ns=namespace, dep=deployment, cntr=container, lbl=label, ann=annotation

For multiple clusters use commas to separate kube configs or kube contexts
--kubeconfig ~/.kube/config --context cluster-1,cluster-2
--kubeconfig ~/.kube/cluster1,~/.kube/cluster2 --context cluster-1,cluster-2

The filter spec is interpreted as 'must be in (ns1 OR ns2) AND (dep1 OR dep2) AND (cntr1 OR cntr2)...'
The log will be included only if the container matches at least one include filter and does not match any exclude filters.
All parts of the filter are optional and can be omitted e.g. ns1//pod1 filters only for namespace ns1 and pod1.
All names except label and annotation keys support '*' glob matching pattern.

e.g.
--include ns1,ns2 (only namespaces ns1 and ns2)
--include n*//p*/l=v* (pods with name beginning with 'p' in namespaces beginning with 'n' and having label 'l' with value beginning with 'v'.)

```
meshctl debug report [flags]
```

### Options

```
      --context string               Name of the kubeconfig Context. For multiple contexts use a comma (Ex. cluster1,cluster2)
      --critical-errs strings        List of comma separated glob patters to match against log error strings. If any pattern matches an error in the log, the logs is given the highest priority for archive inclusion.
      --dir string                   Set a specific directory for temporary artifact storage.
      --dry-run                      Only log commands that would be run, don't fetch or write.
      --duration duration            How far to go back in time from end-time for log entries to include in the archive. Default is infinity. If set, start-time must be unset.
      --end-time string              End time for the range of log entries to include in the archive. Default is now.
      --exclude strings              Spec for which pods' proxy logs to exclude from the archive, after the include spec is processed. See above for format and examples. (default ["kube-system, kube-public, kube-node-lease, local-path-storage"])
  -f, --filename string              Path to a file containing configuration in YAML format. The file contents are applied over the default values and flag settings, with lists being replaced per JSON merge semantics.
      --full-secrets                 If set, secret contents are included in output.
      --gloo-mesh-namespace string   Namespace where Gloo Mesh is installed. (default "gloo-mesh")
  -h, --help                         help for report
      --ignore-errs strings          List of comma separated glob patters to match against log error strings. Any error matching these patters is ignored when calculating the log importance heuristic.
      --include strings              Spec for which pods' proxy logs to include in the archive. See above for format and examples.
      --istio-namespace string       Namespace where Istio control plane is installed. (default "istio-system")
  -c, --kubeconfig string            Path to kube context. For multiple files use a comma (Ex. ~/.kube/cluster1,~/.kube/cluster2)
      --start-time string            Start time for the range of log entries to include in the archive. Default is the infinite past. If set, Since must be unset.
      --timeout duration             Maximum amount of time to spend fetching logs. When timeout is reached only the logs captured so far are saved to the archive. (default 30m0s)
```

### Options inherited from parent commands

```
  -v, --verbose   Enable verbose logging
```

### SEE ALSO

* [meshctl debug](../meshctl_debug)	 - Debug Gloo Mesh resources

