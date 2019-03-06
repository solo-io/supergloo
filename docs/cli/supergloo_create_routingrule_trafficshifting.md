---
title: "supergloo create routingrule trafficshifting"
weight: 5
---
## supergloo create routingrule trafficshifting

ts

### Synopsis

Traffic Shifting rules are used to divert HTTP requests sent within the mesh from their original destinations. 
This can be used to force traffic to be sent to a specific subset of a service, a different service entirely, and/or 
be load-balanced by weight across a variety of destinations

```
supergloo create routingrule trafficshifting [flags]
```

### Options

```
      --destination TrafficShiftingValue   append a traffic shifting destination. format must be <NAMESPACE>.<NAME>:<WEIGHT>
  -h, --help                               help for trafficshifting
```

### Options inherited from parent commands

```
      --dest-labels MapStringStringValue       apply this rule to requests sent to pods with these labels. format must be KEY=VALUE (default [])
      --dest-namespaces strings                apply this rule to requests sent to pods in these namespaces
      --dest-upstreams ResourceRefsValue       apply this rule to requests sent to these upstreams. format must be <NAMESPACE>.<NAME>. (default [])
  -i, --interactive                            run in interactive mode
      --name string                            name for the resource
      --namespace string                       namespace for the resource (default "supergloo-system")
  -o, --output string                          output format: (yaml, json, table)
      --request-matcher RequestMatchersValue   json-formatted string which can be parsed as a RequestMatcher type, e.g. {"path_prefix":"/users","path_exact":"","path_regex":"","methods":["GET"],"header_matchers":{"x-custom-header":"bar"}} (default [])
      --source-labels MapStringStringValue     apply this rule to requests originating from pods with these labels. format must be KEY=VALUE (default [])
      --source-namespaces strings              apply this rule to requests originating from pods in these namespaces
      --source-upstreams ResourceRefsValue     apply this rule to requests originating from these upstreams. format must be <NAMESPACE>.<NAME>. (default [])
```

### SEE ALSO

* [supergloo create routingrule](../supergloo_create_routingrule)	 - Create a routing rule to apply to one or more meshes.

