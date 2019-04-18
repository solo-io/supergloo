---
title: "supergloo apply routingrule retries"
weight: 5
---
## supergloo apply routingrule retries

rt

### Synopsis

Retry rules are used to retry failed requests within a Mesh. 
The retries command contains subcommands for different types of retry policies. 
The retry policy you choose may only be compatible with a certain mesh type.
See documentation at https://supergloo.solo.io for more information.

### Options

```
  -h, --help   help for retries
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

* [supergloo apply routingrule](../supergloo_apply_routingrule)	 - Apply a routing rule to one or more meshes.
* [supergloo apply routingrule retries budget](../supergloo_apply_routingrule_retries_budget)	 - b
* [supergloo apply routingrule retries max](../supergloo_apply_routingrule_retries_max)	 - m

