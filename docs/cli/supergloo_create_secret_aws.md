---
title: "supergloo create secret aws"
weight: 5
---
## supergloo create secret aws

Create an AWS secret

### Synopsis

Creates a secret holding AWS access credentials. You can provide the access-key-id and secret-access-key 
either directly via the correspondent flags, or by passing the location of an AWS credentials file.

```
supergloo create secret aws [flags]
```

### Options

```
      --access-key-id string       AWS Access Key ID
  -f, --file string                path to the AWS credentials file
  -h, --help                       help for aws
      --profile string             name of the desired AWS profile in the credentials file
      --secret-access-key string   AWS Secret Access Key
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

