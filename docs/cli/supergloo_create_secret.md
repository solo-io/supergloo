---
title: "supergloo create secret"
weight: 5
---
## supergloo create secret

create a secret for use with SuperGloo.

### Synopsis

SuperGloo uses secrets to authenticate to external APIs
and manage TLS certificates used for encryption in the mesh and ingress. 


### Options

```
  -h, --help               help for secret
      --name string        name for the resource
      --namespace string   namespace for the resource (default "supergloo-system")
  -o, --output string      output format: (yaml, json, table)
```

### Options inherited from parent commands

```
  -i, --interactive   use interactive mode
```

### SEE ALSO

* [supergloo create](../supergloo_create)	 - commands for creating managing resources used for SuperGloo
* [supergloo create secret tls](../supergloo_create_secret_tls)	 - create a tls secret with cert

