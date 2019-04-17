---
title: "supergloo apply routingrule retries max"
weight: 5
---
## supergloo apply routingrule retries max

m

### Synopsis

m

```
supergloo apply routingrule retries max [flags]
```

### Options

```
  -a, --attempts uint32            REQUIRED. Number of retries for a given request. The interval between retries will be determined automatically (25ms+). Actual number of retries attempted depends on the httpReqTimeout.
  -h, --help                       help for max
  -t, --per-try-timeout duration   Timeout per retry attempt for a given request. format: 1h/1m/1s/1ms. MUST BE >=1ms
  -r, --retry-on string            Specifies the conditions under which retry takes place. One or more policies can be specified using a ‘,’ delimited list. The supported policies can be found in <https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#x-envoy-retry-on> and <https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#x-envoy-retry-grpc-on>
```

### Options inherited from parent commands

```
      --dest-labels MapStringStringValue       apply this rule to requests sent to pods with these labels. format must be KEY=VALUE (default [])
      --dest-namespaces strings                apply this rule to requests sent to pods in these namespaces
      --dest-upstreams ResourceRefsValue       apply this rule to requests sent to these upstreams. format must be <NAMESPACE>.<NAME>. (default [])
      --dryrun                                 if true, this command will print the yaml used to create a kubernetes resource rather than directly trying to create/apply the resource
  -i, --interactive                            run in interactive mode
      --name string                            name for the resource
      --namespace string                       namespace for the resource (default "supergloo-system")
  -o, --output string                          output format: (yaml, json, table)
      --request-matcher RequestMatchersValue   json-formatted string which can be parsed as a RequestMatcher type, e.g. {"path_prefix":"/users","path_exact":"","path_regex":"","methods":["GET"],"header_matchers":{"x-custom-header":"bar"}} (default [])
      --source-labels MapStringStringValue     apply this rule to requests originating from pods with these labels. format must be KEY=VALUE (default [])
      --source-namespaces strings              apply this rule to requests originating from pods in these namespaces
      --source-upstreams ResourceRefsValue     apply this rule to requests originating from these upstreams. format must be <NAMESPACE>.<NAME>. (default [])
      --target-mesh ResourceRefValue           select the target mesh or mesh group to which to apply this rule. format must be NAMESPACE.NAME (default { })
```

### SEE ALSO

* [supergloo apply routingrule retries](../supergloo_apply_routingrule_retries)	 - rt

