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
* `--mesh` (string, optional) name of mesh which will be the target for this rule
* `--namespace` (string, optional, default "default") namespace this rule will be installed into
* `--destinations` (string, optional) destinations for this rule. Each entry consists of an upstream namespace and and upstream name, separated by a colon. eg `<namespace:name>`
* `--sources` (string, optional) Sources for this rule. Each entry consists of an upstream namespace and and upstream name, separated by a colon. eg `<namespace:name>`
* `-f --filename` (string, optional) filename containing info to construct route rule. Only yaml supported currently.

Creating routing rule functions as the sum total of all of the individual rule commands listed below. 
`create routing-rule` allows for the creation of multiple rules simultaneously, either via flags, or via yaml.

#####Traffic Shifting

```bash
supergloo traffic-shifting [flags] <resource-name>
```

Flags:
* `--mesh.name` (string, optional) name of mesh to update
* `--mesh.namespace` (string, optional) name of mesh to update
* `--serviceid` (string, optional) service to modify
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
* `--mesh.name` (string, optional) name of mesh to update
* `--mesh.namespace` (string, optional) name of mesh to update
* `--serviceid` (string, optional) service to modify


#####Timeout

```bash
supergloo timeout [flags] <resource-name>
```

Flags:
* `--mesh.name` (string, optional) name of mesh to update
* `--mesh.namespace` (string, optional) name of mesh to update
* `--serviceid` (string, optional) service to modify
* `--route.timeout.seconds` (int, required) timeout duration (seconds)
* `--rout.timeout.nanos` (int, required) timeout duration (nanoseconds)

#####Retries

```bash
supergloo retries [flags] <resource-name>
```

Flags:
* `--mesh.name` (string, optional) name of mesh to update
* `--mesh.namespace` (string, optional) name of mesh to update
* `--serviceid` (string, optional) service to modify
* `--route.retry.attempt` (int, required) number of retries to attempt
* `--route.retry.timeout.seconds` (int, required) timeout duration (seconds)
* `--route.retry.timeout.nanos` (int, required) timeout duration (nanoseconds)

#####Cors Policy

```bash
supergloo cors [flags] <resource-name>
```

#####Mirror

```bash
supergloo mirror [flags] <resource-name>
```

#####Header Manipulation

```bash
supergloo header-manipulation [flags] <resource-name>
```
