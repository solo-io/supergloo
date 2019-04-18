---
title: "supergloo apply routingrule faultinjection"
weight: 5
---
## supergloo apply routingrule faultinjection

fi

### Synopsis

Fault injection rules are used to inject faults into requests in order to test for tolerance.

```
supergloo apply routingrule faultinjection [flags]
```

### Options

```
  -h, --help            help for faultinjection
  -p, --percent float   percentage of traffic to fault-inject
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
* [supergloo apply routingrule faultinjection abort](../supergloo_apply_routingrule_faultinjection_abort)	 - a
* [supergloo apply routingrule faultinjection delay](../supergloo_apply_routingrule_faultinjection_delay)	 - d

