---
title: "supergloo apply securityrule"
weight: 5
---
## supergloo apply securityrule

Apply a security rule to one or more meshes.

### Synopsis


Each Security Rule applies an HTTP security policy to a mesh.

When no security rules are present in the system, SuperGloo will allow 
all traffic between services. As soon as a security rule is created, 
traffic policy is enforced, and only explicitly allowed traffic will be 
permitted.

Security rules implement the following semantics:

RULE:
  ALLOW HTTP Requests:
  - FROM these **source pods**
  - TO these **destination pods**
  - MATCHING these **allowed paths and methods**


```
supergloo apply securityrule [flags]
```

### Options

```
      --allowed-methods strings              HTTP methods that are allowed for this rule. Leave empty to allow all paths
      --allowed-paths strings                HTTP paths that are allowed for this rule. Leave empty to allow all paths
      --dest-labels MapStringStringValue     apply this rule to requests sent to pods with these labels. format must be KEY=VALUE (default [])
      --dest-namespaces strings              apply this rule to requests sent to pods in these namespaces
      --dest-upstreams ResourceRefsValue     apply this rule to requests sent to these upstreams. format must be <NAMESPACE>.<NAME>. (default [])
      --dryrun                               if true, this command will print the yaml used to create a kubernetes resource rather than directly trying to create/apply the resource
  -h, --help                                 help for securityrule
  -i, --interactive                          run in interactive mode
      --name string                          name for the resource
      --namespace string                     namespace for the resource (default "supergloo-system")
  -o, --output string                        output format: (yaml, json, table)
      --source-labels MapStringStringValue   apply this rule to requests originating from pods with these labels. format must be KEY=VALUE (default [])
      --source-namespaces strings            apply this rule to requests originating from pods in these namespaces
      --source-upstreams ResourceRefsValue   apply this rule to requests originating from these upstreams. format must be <NAMESPACE>.<NAME>. (default [])
      --target-mesh ResourceRefValue         select the target mesh or mesh group to which to apply this rule. format must be NAMESPACE.NAME (default { })
```

### SEE ALSO

* [supergloo apply](../supergloo_apply)	 - apply a rule to a mesh

