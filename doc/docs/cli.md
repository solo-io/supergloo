Command-Line Interface 
----
The supergloo cli is a command line wrapper for supergloo's [REST API](https://supergloo.solo.io/v1/supergloo.solo.io.project.sk/).

* [`Initializing supergloo`](cli.md#initializing-supergloo)
* [`Installing Meshes`](cli.md#installing-meshes)
* [`Uninstalling Meshes`](cli.md#uninstalling-meshes)
* [`Routing Rules`](cli.md#routing-rules)
* Getting
* Policies

###supergloo-cli conventions
In order to create a seamless experience, the supergloo cli is interactive by default.
However, if you would not like to take advantage of these features, feel free to use the `static -s` flag and turn them off.


####Initializing supergloo

The first step with the CLI is Initializing supergloo in your desired cluster.

```bash
supergloo init
```

The above command will create supergloo-system namespace in your cluster, as well as all of it's related resources.

---

####Installing Meshes


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
WIP: DO THIS LATER
```

---

####Uninstalling Meshes

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

####Routing Rules

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
* `--mesh` (string, required) name of mesh which will be the target for this rule
* `--namespace` (string, required, default "default") namespace this rule will be installed into
* `--destinations` (string, required) destinations for this rule. Each entry consists of an upstream namespace and and upstream name, separated by a colon. eg `<namespace:name>`
* `--sources` (string, required) Sources for this rule. Each entry consists of an upstream namespace and and upstream name, separated by a colon. eg `<namespace:name>`
* `-f --filename` (string, optional) filename containing info to construct route rule. Only yaml supported currently. note: currently this feature supports multiple data formats.
Either use the form defined in the `routing_rule.proto` or one that conforms to the CLI.

Creating routing rule functions as the sum total of all of the individual rule commands listed below. 
`create routing-rule` allows for the creation of multiple rules simultaneously, either via flags, or via yaml.

#####Traffic Shifting

```bash
supergloo traffic-shifting [flags] <resource-name>
```

Flags:
* `--traffic.upstreams` (string, optional) upstreams for this rule. Each entry consists of an upstream namespace and and upstream name, separated by a colon. eg `<namespace:name>`
* `--traffic.weights` (string, optional) Comma-separated list of integer weights corresponding to the associated upstream's traffic sharing percentage. Must be the same length as the # of upstreams


Example: 
```bash
supergloo traffic-shifting [flags] ts-rule
```

#####Fault Injection

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


#####Timeout

```bash
supergloo timeout [flags] <resource-name>
```

Flags:
* `--route.timeout.seconds` (int, required) timeout duration (seconds)
* `--rout.timeout.nanos` (int, required) timeout duration (nanoseconds)

#####Retries

```bash
supergloo retries [flags] <resource-name>
```

Flags:
* `--route.retry.attempt` (int, required) number of retries to attempt
* `--route.retry.timeout.seconds` (int, required) timeout duration (seconds)
* `--route.retry.timeout.nanos` (int, required) timeout duration (nanoseconds)

#####Cors Policy

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

#####Mirror

```bash
supergloo mirror [flags] <resource-name>
```

Flags:
* `--mirror` (string, required) Destination upstream (ex: upstream_namespace:upstream_name).

#####Header Manipulation

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
supergloo header-manipulation hm-rule --mesh.name mesh --mesh.namespace mesh-namespace \
    --header.request.append header-request1,text --header.request.remove header-request2 \
    --header.response.append header-response1,text --header.request.remove header-response2
```
The above command creates a new header-manipulation rule for the mesh (`mesh`) in the namespace (`namespace`).
This rule will append the request header `[header-request1, text]` and remvove the request header `header-request2`.
This rule will then append the response header `[]`
