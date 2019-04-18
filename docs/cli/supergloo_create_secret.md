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
  -i, --interactive        run in interactive mode
      --name string        name for the resource
      --namespace string   namespace for the resource (default "supergloo-system")
  -o, --output string      output format: (yaml, json, table)
```

### SEE ALSO

* [supergloo create](../supergloo_create)	 - commands for creating resources used by SuperGloo
* [supergloo create secret aws](../supergloo_create_secret_aws)	 - Create an AWS secret
* [supergloo create secret tls](../supergloo_create_secret_tls)	 - create a tls secret with cert

