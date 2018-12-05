Command-Line Interface 
----
The supergloo cli is a command line wrapper for supergloo's [REST API](https://supergloo.solo.io/v1/supergloo.solo.io.project.sk/).

 In order to create a seamless experience, the supergloo cli is interactive by default.
However, if you would not like to take advantage of these features, feel free to use the `static -s` flag and turn them off.


* [`Initializing supergloo`](cli.md#initializing-supergloo)
* [`Installing Meshes`](cli.md#installing-meshes)
* [`Uninstalling Meshes`](cli.md#uninstalling-meshes)
* [`Routing Rules`](cli.md#routing-rules)
* [`Security`](cli.md#security)
* [`Policies`](cli.md#policies)
* [`Configuration`](cli.md#configuration)
* [`Get`](cli.md#get)



### Initializing supergloo

The first step with the CLI is Initializing supergloo in your desired cluster.

```bash
supergloo init
```

The above command will create supergloo-system namespace in your cluster, as well as all of it's related resources.

---

### Installing Meshes


```bash
supergloo install [flags]
```

Flags:
 
* `--aws-region` (string, optional)
* `--awssecret.name` (string, optional)
* `--awssecret.namespace` (string, optional)
* `--meshtype` (string, required, options: "istio", "consul", "linkerd2", "" ) name of mesh to install
* `--mtls` (boolean, optional, default: false) enable mtls for mesh
* `--secret.name` (string, optional)
* `--secret.namespace` (string, optional)

Example usage: 
```bash
supergloo install --meshtype istio --namespace istio-system
```
The above command will install istio into the istio-system namespace


---

### Uninstalling Meshes

```bash
supergloo uninstall [flags]
```


Flags:
* `--all` (boolean, optional)
* `--meshname` (string, required)
* `--meshtype` (string, optional)

Example usage:
```bash
supergloo uninstall --meshname <name-of-mesh>
```

---

### Routing Rules

* [`Traffic Shifting`](cli.md#traffic-shifting)
* [`Fault Injection`](cli.md#fault-injection)
* [`Timeout`](cli.md#timeout)
* [`Retries`](cli.md#retries)
* [`Cors Policy`](cli.md#cors-policy)
* [`Mirror`](cli.md#mirror)
* [`Header Manipulation`](cli.md#header-manipulation)

```bash
supergloo create routing-rule [flags] <resource-name>
```


Persistent flags:
* `--matchers` (string, optional) list of comma seperated matchers. eg: ("*")
* `-m --mesh` (string, required) name of mesh which will be the target for this rule
* `-n --namespace` (string, required, default: "default") namespace this rule will be installed into
* `-d --destinations` (string, required) destinations for this rule. Each entry consists of an upstream namespace and and upstream name, separated by a colon. eg `<namespace:name>`
* `--sources` (string, required) Sources for this rule. Each entry consists of an upstream namespace and and upstream name, separated by a colon. eg `<namespace:name>`
* `-f --filename` (string, optional) filename containing info to construct route rule. Only yaml supported currently. note: currently this feature supports multiple data formats.
Either use the form defined in the `routing_rule.proto` or one that conforms to the CLI.
* `-s --static` (bool, optional, default: false) As stated above. the CLI by default is interactive unless specified via this flag


Notes:
* `upstreams` are found via our discovery system and can be found easily using the following command. It is important to check all namespaces because the user is given the freedom to place these wherever he/she pleases.
```bash
kubectl get routingrules.supergloo.solo.io --all-namespaces
```

Create routing rule functions as the sum total of all of the individual rule commands listed below. 
`create routing-rule` allows for the creation of multiple rules simultaneously, either via flags, or via yaml.

#### Traffic Shifting

```bash
supergloo traffic-shifting [flags] <resource-name>
```

Flags:
* `--traffic.upstreams` (string, optional) upstreams for this rule. Each entry consists of an upstream namespace and and upstream name, separated by a colon. eg `<namespace:name>`
* `--traffic.weights` (string, optional) Comma-separated list of integer weights corresponding to the associated upstream's traffic sharing percentage. Must be the same length as the # of upstreams


Example: 
```bash
supergloo traffic-shifting ts-rule -s -m <mesh-name> -n <namespace> \
    --destinations <namespace:name> --sources <namespace:name> \
    --traffic.upstreams <namespace:name>, <namespace:name> --traffic.weights 10, 20
```

#### Fault Injection

```bash
supergloo fault-injection [flags] <resource-name>
```

FLags:
* `--fault.abort.message` (string, required) Error message (int for type=http errors, string otherwise).
* `--fault.abort.percent` (int, required) Percentage of requests on which the abort will be injected (0-100)
* `--fault.abort.type` (string, required) Type of error (http, http2, or grpc)
* `--fault.delay.percent` (int , required) Percentage of requests on which the delay will be injected (0-100)
* `--fault.delay.type` (string, required) Type of delay (fixed or exponential)
* `--fault.delay.value.seconds` (int, required) delay duration (seconds)
* `--fault.delay.value.nanos` (int, required) delay duration (nanoseconds)


#### Timeout

```bash
supergloo timeout [flags] <resource-name>
```

Flags:
* `--route.timeout.seconds` (int, required) timeout duration (seconds)
* `--rout.timeout.nanos` (int, required) timeout duration (nanoseconds)

#### Retries

```bash
supergloo retries [flags] <resource-name>
```

Flags:
* `--route.retry.attempt` (int, required) number of retries to attempt
* `--route.retry.timeout.seconds` (int, required) timeout duration (seconds)
* `--route.retry.timeout.nanos` (int, required) timeout duration (nanoseconds)

#### Cors Policy

```bash
supergloo cors [flags] <resource-name>
```

Flags:
* `--cors.allow.credentials` (bool, required) Indicates whether the caller is allowed to send the actual request (not the preflight) using credentials. Translates to Access-Control-Allow-Credentials header.
* `--cors.allow.headers` (string, required) List of HTTP headers that can be used when requesting the resource. Serialized to Access-Control-Allow-Methods header.
* `--cors.allow.methods` (string, required) List of HTTP methods allowed to access the resource. The content will be serialized into the Access-Control-Allow-Methods header.
* `--cors.allow.origin` (string, required) The list of origins that are allowed to perform CORS requests. The content will be serialized into the Access-Control-Allow-Origin header. Wildcard * will allow all origins.
* `--cors.expose.headers` (string, required) A white list of HTTP headers that the browsers are allowed to access. Serialized into Access-Control-Expose-Headers header.
* `--cors.maxage.seconds` (int, required) Max age time in seconds. Specifies how long the the results of a preflight request can be cached. Translates to the Access-Control-Max-Age header.
* `--cors.maxage.nanos` (int, required) Max age time in nanoseconds. Specifies how long the the results of a preflight request can be cached. Translates to the Access-Control-Max-Age header.

#### Mirror

```bash
supergloo mirror [flags] <resource-name>
```

Flags:
* `--mirror` (string, required) Destination upstream (ex: upstream_namespace:upstream_name).

Example:
```bash
supergloo mirror m-rule -s --mesh <mesh-name> --namespace <mesh-namespace> \
    --destinations <namespace:name> --sources <namespace:name> \
    --mirror <namespace:name>
```
The above command creates a new mirror rule

#### Header Manipulation

```bash
supergloo header-manipulation [flags] <resource-name>
```


Flags:
* `--header.request.append` (string, optional) Headers to append to request (ex: h1,v1,h2,v2).
* `--header.request.remove` (string, optional) Headers to remove from request (ex: h1,h2).
* `--header.response.append` (string, optional) Headers to append to response (ex: h1,v1,h2,v2).
* `--header.response.remove` (string, optional) Headers to remove from response (ex: h1,h2).

Example:
```bash
supergloo header-manipulation hm-rule -s --mesh <mesh-name> --namespace <mesh-namespace> \
    --destinations <namespace:name> --sources <namespace:name> \
    --header.request.append header-request1,text --header.request.remove header-request2 \
    --header.response.append header-response1,text --header.request.remove header-response2
```
The above command creates a new header-manipulation rule for the mesh (`mesh`) in the namespace (`namespace`).
This rule will append the request header `[header-request1, text]` and remvove the request header `header-request2`.
This rule will then append the response header `[header-response1, text]` and remove the response header `header-response2`.

---


### Security

security features: 
* [`policies`](cli.md#policies)
* [`mtls`](cli.md#mtls)
* [`ingress`](cli.md#fortify-ingress)

#### Policies

TODO: Brief description here

```bash
supergloo policy [sub-command] [flags] 
```

Sub-Commands:
* `add` apply a policy
* `claer` clear all policies
* `remove` remove a single policy

Persistent-Flags:
* `--mesh.name` (string, required) name of mesh to update
* `--mesh.namespace` (string, required) namespace of mesh to update
* `--destination.name` (string, required) name of policy destination upstream
* `--destination.namespace` (string, required) namespace of policy destination upstream
* `--source.name` (string, required) name of policy source upstream
* `--source.namespace` (string, required) namespace of policy source upstream


Add:

Adds a single policy to a given mesh with the specified source and destination
```bash
supergloo policy add -s --mesh.name <mesh-name> --mesh.namespace <namespace> \
    --destination.name <upstream-name> --destination.namespace <upstream-namespace> \
    --source.name <upstream-name> --source.namespace <upstream-namespace>
```

Remove: 

Removes a single policy from a given mesh with the specified source and destination.
```bash
supergloo policy remove -s --mesh.name <mesh-name> --mesh.namespace <namespace> \
    --destination.name <upstream-name> --destination.namespace <upstream-namespace> \
    --source.name <upstream-name> --source.namespace <upstream-namespace>
```

Clear:

Clears all policies from a given mesh
```bash
supergloo policy remove -s --mesh.name <mesh-name> --mesh.namespace <namespace>
```
#### mTLS

control mTLS for a given mesh

```bash
supergloo mtls [sub-command] [flags] 
```

Sub-Commands:
* `disable` disable mTLS
* `enable` enable mTLS
* `toggle` toggle mTLS

Persistent-Flags:
* `--mesh.name` (string, required) name of mesh to update
* `--mesh.namespace` (string, required) namespace of mesh to update


Example:
```bash
supergloo mtls enable -s --mesh.namespace <namespace> --mesh.name <mesh-name>
```

#### Fortify Ingress

Configure ingress security parameters

This feature will be available in 2019.

---

### Configuration

---


### Get

Fetch supergloo data/CRDs in a readable format

```bash
supergloo get <RESOURCE-TYPE> <resource-name> [flags]
```

This command functions similarly to `kubectl get`. The first arg is the resource type, and the second (optional) arg is the name of the individual resource.
 
Sub-Commands:
*  `-r --resources` prints a list of available resource types and their corresponding shortcuts

Flags:
* `-n --namespace` (string, optional, default: "gloo-system) namespace to search for the corresponding CRD
* `-o --output` (string, optional) output format for the data, small table by default (yaml, wide...)
