---
title: Using Meshctl for Wasm
menuTitle: Meshctl for Wasm
weight: 120
---

{{% notice note %}}
Gloo Mesh Enterprise is required for this feature.
{{% /notice %}}

In this guide we will walk through how you can use the wasm plugin for the `meshctl` command line tool to create and apply a Wasm-based Envoy filter to an Istio service mesh managed by Gloo Mesh Enterprise. To accomplish our goal, we will walk through the following steps:

1. Initialize a Wasm filter
1. Build a Wasm filter
1. Push the Wasm filter to WebAssembly Hub
1. Deploy the Wasm filter to an Istio managed instance of Envoy

## Before you begin
To illustrate these concepts, we will assume that you have already followed the [Wasm Extension Guide for Gloo Mesh Enterprise]({{% versioned_link_path fromRoot="/guides/wasm_extension/" %}}) to get all the necessary components ready for the deploy step.

You will also need the wasm plugin for `meshctl`. It can be installed by running the following one-liner:

```shell
curl solo.io/meshctl-wasm/foo/bar/install | sh
```

We will be pushing our filter to the publicly hosted WebAssembly Hub, so you will need to sign up for a user account following [this guide](https://docs.solo.io/web-assembly-hub/latest/tutorial_code/push_tutorials/basic_push/#create-a-user-on-webassemblyhub-io-https-webassemblyhub-io).

{{% notice note %}}
Be sure to review the assumptions and satisfy the pre-requisites from the [Guides]({{% versioned_link_path fromRoot="/guides" %}}) top-level document.
{{% /notice %}}

## Initialize a Wasm filter

`meshctl` is a CLI tool that helps bootstrap Gloo Mesh, register clusters, describe configured resources, and more. The `wasm` plugin allows `meshctl` to perform the necessary actions to initialize, push, and deploy Wasm filters. Our first step is to initialize a directory where we will build a basic filter in AssemblyScript.

Navigate to the location where you would like to create the directory and then run the following command:

```shell
meshctl wasm init assemblyscript-filter
```

You will be asked which language platform you are building for.

```shell
Use the arrow keys to navigate: ↓ ↑ → ←
? What language do you wish to use for the filter:
    cpp
    rust
  ▸ assemblyscript
    tinygo
```

Use the arrow keys to select `assemblyscript` and press **enter**.

You will next be prompted to select which platforms the filter will be used on.

```shell
✔ assemblyscript
Use the arrow keys to navigate: ↓ ↑ → ←
? With which platforms do you wish to use the filter?:
  ▸ istio:1.8.x
```

Currently the only supported platform is Istio 1.8. Press **enter** to complete the process.

```shell
INFO[0188] extracting 1798 bytes to /home/gloo/wasm/assemblyscript-filter
```

If we run `tree` on the created directory, you should see a layout similar to what is shown below.

```shell
assemblyscript-filter/
├── assembly
│   ├── index.ts
│   └── tsconfig.json
├── package-lock.json
├── package.json
└── runtime-config.json
```

With our project successfully initialized, open it up in your favorite IDE. We are going to customize our new filter by making a few changes.

## Build a Wasm filter

The new directory contains all files necessary to build and deploy a Wasm filter with `meshctl`. A brief description of each file is found below:

| File | Description |
| ---- | ----------- |
| assembly/index.ts | The source code for the filter, written in AssemblyScript. |
| assembly/tsconfig.json |	Typescript config file (AssemblyScript is a subset of Typescript). |
| package.json | Used by to import npm modules during build time. |
| package-lock.json | Locked npm modules. |
| runtime-config.json | Config stored with the filter image used to load the filter at runtime. |

Open `assembly/index.ts` in your favorite text editor. The source code is AssemblyScript and we’ll make some changes to customize our new filter.

Navigate to the `onResponseHeaders` method defined in the file. This method will add a header to the response from a service behind an Envoy proxy. We are going to customize the header by updating the highlighted lines below with new values:

{{< highlight javascript "hl_lines=10-12" >}}
class AddHeader extends Context {
  root_context : AddHeaderRoot;
  constructor(root_context:AddHeaderRoot){
    super();
    this.root_context = root_context;
  }
  onResponseHeaders(a: u32): FilterHeadersStatusValues {
    const root_context = this.root_context;
    if (root_context.configuration == "") {
      stream_context.headers.response.add("wasm", "built-with-meshctl!");
    } else {
      stream_context.headers.response.add("wasm", root_context.configuration);
    }
    return FilterHeadersStatusValues.Continue;
  }
}
{{< /highlight >}}

Now save the file and we will build our updated filter with `meshctl`. The filter will be tagged and stored in a local registry, similar to how Docker stores images.

Images tagged with `meshctl` have the following format:

```shell
<registry address>/<registry username|org>/<image name>:<version tag>
```

* `<registry address>` specifies the address of the remote OCI registry where the image will be pushed by the `meshctl wasm push` command. The project authors maintain a free public registry at webassemblyhub.io.
* `<registry username|org>` either your username for the remote OCI registry, or a valid org name with which you are registered.

In this example we’ll include the registry address `webassemblyhub.io` so our image can be pushed to the remote registry, along with the WebAssembly Hub username which will be used to authenticate to the registry. As we mentioned at the beginning of this guide, you will need to [sign up for an account on WebAssembly Hub](https://docs.solo.io/web-assembly-hub/latest/tutorial_code/push_tutorials/basic_push/#create-a-user-on-webassemblyhub-io-https-webassemblyhub-io).

Build and tag your image by running the following commands from the `assemblyscript-filter/` directory:

```shell
HUB_USERNAME=<your WebAssembly Hub username>

meshctl wasm build assemblyscript -t webassemblyhub.io/$HUB_USERNAME/add-header:v0.1 .
```

{{< notice note >}}
`meshctl wasm build` runs a build container inside of Docker which may run into issues due to SELinux (on Linux environments). To disable, run `sudo setenforce 0`
{{< /notice >}}

The module will take up to a few minutes to build. In the background, `meshctl wasm` has launched a Docker container to run the necessary build steps.

```shell
INFO[0014] adding image to cache...                      filter file=/tmp/gloo-mesh-wasm339603922/filter.wasm tag="webassemblyhub.io/gloo/add-header:v0.1"
INFO[0014] tagged image                                  digest="sha256:a515a5d244b021c753f2e36c744e03a109cff6f5988e34714dbe725c904fa917" image="webassemblyhub.io/gloo/add-header:v0.1"
```

When the build has finished, you’ll be able to see the image with `meshctl wasm list`:

```shell
NAME                              TAG  SIZE    SHA      UPDATED
webassemblyhub.io/gloo/add-header v0.1 12.6 kB a515a5d2 15 Dec 20 11:00 EST
```

Now we can push the image up to WebAssembly hub.

## Push the Wasm filter to WebAssembly Hub

In this step, we'll take the image of the Wasm filter we just created and push it up to WebAssembly Hub. 

We've already tagged our image with with `webassemblyhub.io` and our username, so the next step is to log into WebAssembly Hub:

```shell
meshctl wasm login -u $HUB_USERNAME
Enter password: 
```

You will be prompted to enter your password. Once your login is complete, the credentials will be stored for future reference.

```shell
INFO[0018] Successfully logged in as gloo (Gloo Mesh) 
INFO[0018] stored credentials in /home/gloo/.meshctl/wasm/credentials.json
```

With a successful authentication, we are now ready to push our Wasm image. 

Pushing the image is similar to pushing a Docker image, except we are using the `meshctl wasm` plugin.

```shell
meshctl wasm push webassemblyhub.io/$HUB_USERNAME/add-header:v0.1
```

You should see output similar to what's shown below.

```shell
INFO[0000] Pushing image webassemblyhub.io/gloo/add-header:v0.1 
INFO[0001] Pushed webassemblyhub.io/gloo/add-header:v0.1 
INFO[0001] Digest: sha256:d696cba6bd95e6f0e45d87fe5698da7c27ff036477a57c7fe3d1f0708042d92c
```

We can verify the image has been pushed by using the `list` subcommand with the `--search` flag.

```shell
meshctl wasm list --search $HUB_USERNAME
```

```shell
NAME                              TAG  SIZE    SHA      UPDATED
webassemblyhub.io/gloo/add-header v0.1 13.9 kB d696cba6 15 Dec 20 18:56 UTC
```

With our image pushed to a repository, we are now able to use it in an Istio configuration.

## Deploy the Wasm filter

We are going to deploy our Wasm filter on Istio 1.8 to the bookinfo application running in the `meshctl` namespace on the management cluster. We will be using Gloo Mesh Enterprise's Wasm extension to do so. As mentioned at the beginning of the guide, we will assume that you have already followed the [Wasm Extension Guide for Gloo Mesh Enterprise]({{% versioned_link_path fromRoot="/guides/wasm_extension/" %}}). That means you have two Kuberenetes clusters, with Istio 1.8 and Gloo Mesh Enterprise installed. 

We are going to deploy just the `reviews` components of the bookinfo application to a new namespace in the management cluster, so we don't have a collision with the bookinfo components you deployed in previous guides.

Run these commands to create the namespace and deploy the components:

```shell
# Set the management cluster context variable
MGMT_CONTEXT=kind-mgmt-cluster

# Select the management context for use
kubectl config use-context $MGMT_CONTEXT

# Create the namespace and label it for Istio
kubectl create ns meshctl
kubectl label namespace meshctl istio-injection=enabled

# Deploy the reviews service account, pod, and service
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.8/samples/bookinfo/platform/kube/bookinfo.yaml -n meshctl -l account=reviews
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.8/samples/bookinfo/platform/kube/bookinfo.yaml -n meshctl -l app=reviews,version=v1
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.8/samples/bookinfo/platform/kube/bookinfo.yaml -n meshctl -l app=reviews,service=reviews
```

In the previous guide, we altered the Envoy filters for v3 of the reviews service in the remote cluster. Included were the steps to apply a ConfigMap, patch a deployment, and write a WasmDeployment CRD. The `meshctl wasm deploy` command takes care of all those steps for you. 

Before we deploy the filter, let's test the reviews service in the meshctl namespace. We'll spin up a temporary pod with and use curl to check.

```bash
kubectl run -it -n meshctl --context $MGMT_CONTEXT curl \
  --image=curlimages/curl:7.73.0 --rm  -- sh

# From the new terminal run the following
curl http://reviews:9080/reviews/1 -v
```

We should see the following output:

```shell
*   Trying 10.96.168.205:9080...
* Connected to reviews (10.96.168.205) port 9080 (#0)
> GET /reviews/1 HTTP/1.1
> Host: reviews:9080
> User-Agent: curl/7.73.0-DEV
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< x-powered-by: Servlet/3.1
< content-type: application/json
< date: Tue, 15 Dec 2020 19:46:31 GMT
< content-language: en-US
< content-length: 295
< x-envoy-upstream-service-time: 573
< server: envoy
<
* Connection #0 to host reviews left intact
{"id": "1","reviews": [{  "reviewer": "Reviewer1",  "text": "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!"},{  "reviewer": "Reviewer2", 
 "text": "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare."}]}/
```

Note that our custom header is not present, but we can see there is an `envoy` value for the `server` header.

Now we can run the `deploy` command to get our Wasm filter in place.

```shell
meshctl wasm deploy \
  --cluster mgmt-cluster \
  --namespace meshctl \
  --labels "app=reviews" \
  --filter-name meshctl-filter \
  --image webassemblyhub.io/$HUB_USERNAME/add-header:v0.1 \
  --mgmt-kubecontext $MGMT_CONTEXT \
  --remote-kubecontext $MGMT_CONTEXT
```

You should see the following output.

```shell
creating envoy bootstrap configmap...
deploying wasm filter
```

Now let's try using `curl` again to check for our custom header.

```bash
kubectl run -it -n meshctl --context $MGMT_CONTEXT curl \
  --image=curlimages/curl:7.73.0 --rm  -- sh

# From the new terminal run the following
curl http://reviews:9080/reviews/1 -v
```

You should see the following output:

{{< highlight shell "hl_lines=16" >}
*   Trying 10.96.177.171:9080...
* Connected to reviews (10.96.177.171) port 9080 (#0)
> GET /reviews/1 HTTP/1.1
> Host: reviews:9080
> User-Agent: curl/7.73.0-DEV
> Accept: */*
>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< x-powered-by: Servlet/3.1
< content-type: application/json
< date: Wed, 16 Dec 2020 14:45:07 GMT
< content-language: en-US
< content-length: 295
< x-envoy-upstream-service-time: 547
< wasm: built-with-meshctl!
< server: envoy
<
* Connection #0 to host reviews left intact
{"id": "1","reviews": [{  "reviewer": "Reviewer1",  "text": "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!"},{  "reviewer": "Reviewer2",  "text": "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare."}]}/
{{< /highlight >}}

We should see the `< wasm: built-with-meshctl!` header in our response if the filter was deployed successfully.

## Summary and Next Steps

In this guide you used `meshctl wasm` to push a Wasm filter to a service managed by Gloo Mesh.

This is a simple example of a Wasm filter to illustrate the concept. The flexibility of Wasm filters coupled with Envoy provides a platform for incredible innovation. Check out our docs on [Web Assembly Hub](https://docs.solo.io/web-assembly-hub/latest) for more information.