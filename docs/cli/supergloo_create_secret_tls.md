---
title: "supergloo create secret tls"
weight: 5
---
## supergloo create secret tls

create a tls secret with cert

### Synopsis

Create a secret with the given name

```
supergloo create secret tls [flags]
```

### Options

```
      --cacert string      path to ca-cert file
      --cakey string       path to ca-key file
      --certchain string   path to cert-chain file
  -h, --help               help for tls
      --rootcert string    path to root-cert file
```

### Options inherited from parent commands

```
  -i, --interactive        run in interactive mode
      --name string        name for the resource
      --namespace string   namespace for the resource (default "supergloo-system")
  -o, --output string      output format: (yaml, json, table)
```

### SEE ALSO

* [supergloo create secret](../supergloo_create_secret)	 - create a secret for use with SuperGloo.

