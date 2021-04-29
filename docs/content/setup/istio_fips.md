---
title: "Install Gloo Mesh Istio FIPS"
menuTitle: Install Gloo Mesh Istio FIPS
description: Installing Gloo Mesh Istio FIPS distribution
weight: 15
---

{{% notice note %}}
Gloo Mesh Enterprise is required for this feature. Open source users and users who do not require the functionality
provided by Gloo Mesh Istio can use Gloo Mesh with an upstream Istio release and should refer to our guide on [installing Istio]({{% versioned_link_path fromRoot="/guides/installing_istio" %}}).
{{% /notice %}}

Some of our users run Istio in regulated environments like [FedRAMP](https://www.gsa.gov/technology/government-it-initiatives/fedramp) (Federal Risk and Authorization Management Program) or they run Istio for software that supports running in those environments. Gloo Mesh [offers FIPS builds of upstream Istio](https://www.solo.io/blog/distroless-fips-compliant-istio/) _without the need for any additional tooling or CLIs_. You can stick with the upstream-native Istio tooling (like [istioctl](https://istio.io/latest/docs/setup/install/istioctl/) or [IstioOperator](https://istio.io/latest/docs/setup/install/operator/)) and prefer to use Solo.io FIPS builds of Istio.

The quickest way to get started with FIPS Istio is to use one of our supported builds. Just run the following command:

```shell
istioctl install --set hub=gcr.io/istio-enterprise --set tag=1.8.5-fips
```

If using an `IstioOperator` file, it would look something like this:

{{< highlight yaml "hl_lines=7-10" >}}
apiVersion: install.istio.io/v1alpha1
kind: IstioOperator
metadata:
  name: gloo-mesh-istio
  namespace: istio-system
spec:
  # This value is required for Gloo Mesh Istio
  hub: gcr.io/istio-enterprise
  # This value can be any Gloo Mesh Istio tag
  tag: 1.8.5-fips
  
... additional config here ...  
{{< /highlight >}}

After installation, you should see something like the following in your `istio-system` namespace:

```shell
NAME                                   READY   STATUS    RESTARTS   AGE
istio-ingressgateway-d487ffdd9-5vsnd   1/1     Running   0          67s
istiod-944958c47-9wvjl                 1/1     Running   0          91s
```

For most auditors, both the control plane AND the data plane must be in FIPS compliant mode. We can verify that by checking the versions of Envoy and istiod. To verify the data plane, run the following command which checks the Envoy Proxy version:

```shell
kubectl exec -it -n istio-system deploy/istio-ingressgateway -- /usr/local/bin/envoy --version
```

An appropriate response should look like this:

```shell
/usr/local/bin/envoy  version: fa9fd362c488508a661d2ffa66e66976bb9104c3/1.15.1/Clean/RELEASE/BoringSSL-FIPS
```

To verify the control plane components, we will copy the pilot-discovery binary out of the istiod container and run [goversion](https://github.com/rsc/goversion) against it.

First, copy the binary out to local disk:

```shell
kubectl cp istio-system/istiod-85ff76b4b5-47pqg:/usr/local/bin/pilot-discovery /tmp/pilot-discovery && chmod +x /tmp/pilot-discovery
```

Then run the following:

```shell
goversion -crypto /tmp/pilot-discovery
```

You should see something similar to:

```shell
/tmp/pilot-discovery go1.14.12b4 (boring crypto)
```

### RSA 4096 key sizes

FIPS 140-2 originally didn't allow for key sizes 4096 but that has been revised in recent addendums. With Gloo Mesh Istio's FIPS builds, you can run the control plane in strict adherence to 140-2 or in a mode that supports 4096 key sizes. For strict mode, you can append `-fipsonly` to the build tag. 

If you run in strict mode (`-fipsonly`), when you verify the control plane you will see something like this:

```shell
/tmp/pilot-discovery go1.14.12b4 (boring crypto) +crypto/tls/fipsonly
```

This mode will not allow keys up to `4096`.

Otherwise you will see:

```shell
/tmp/pilot-discovery go1.14.12b4 (boring crypto)
```

Which will allow `4096` keys.

If you see:

```shell
/tmp/pilot-discovery go1.14.14 (standard crypto)
```

Then you are NOT running a FIPS build of Istio.


### Distroless builds of FIPS

Typically a container should just contain the application and its required dependencies but in reality it contains a lot of other things (shell, package manager, build tools, OS utilities, etc). All of these additional capabilities in the container image increase your exposure to CVS/vulnerabilities. Google opensourced an approach to stripping out any additional cruft from your images except for the necessary application and immediate dependencies called [distroless](https://istio.io/latest/docs/ops/configuration/security/harden-docker-images/).

For example, to install the distroless version of FIPS Istio:

```shell
istioctl install --set hub=gcr.io/istio-enterprise --set tag=1.8.5-fips-distroless
```

Keep in mind [there are some challenges around distroless builds](https://www.solo.io/blog/challenges-of-running-istio-distroless-images/), but luckily we have some solutions.

### Which version is right for you?

We have FIPS builds of various versions of Istio upstream and we can get you the exact version you're looking for. Our pipeline is automated, goes through all of the upstream tests, scanning, and validation steps as well as our own internal gates.

You can see some of the FIPS builds [here](https://console.cloud.google.com/gcr/images/istio-enterprise/GLOBAL/proxyv2?gcrImageListsize=30&gcrImageListquery=%255B%257B_22k_22_3A_22_22_2C_22t_22_3A10_2C_22v_22_3A_22_5C_22fips_5C_22_22%257D%255D) but [please reach out for the version you need](https://www.solo.io/company/contact/)



## Next steps

Now that we have Istio and Gloo Mesh installed ([and appropriate clusters registered]({{% versioned_link_path fromRoot="/setup/#register-a-cluster" %}})), we can continue to explore the [discovery capabilities]({{% versioned_link_path fromRoot="/guides/discovery_intro" %}}) of Gloo Mesh. 
