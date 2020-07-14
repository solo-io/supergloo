# API Server Licensing

*Engineer:* Graham Goudeau

*Domain Area:* Mesh Version Support

*Issue Link:* https://github.com/solo-io/service-mesh-hub/issues/719

## Problem Summary

As meshes release new versions, we need a way to cleanly handle the different requirements as things change. For example,
as Istio versions progress, any of the following breaking changes may occur from one version to the next:

* their configuration API may change
* the number/type of configuration resources we need to write to accomplish the same task may change
* other unforeseen idiosyncrasies of particular mesh versions

We need a way to sanely manage this complexity going into the future.

## Proposed Design

We propose explicitly treating different mesh versions as entirely separate meshes. This has the benefit of both
allowing us to easily handle behaviors of different mesh versions, but also formally state, in code, exactly which mesh versions
we support. 

To be precise about the proposed change, our discovered Mesh 
objects currently have the following mesh type on them:

```proto
    oneof mesh_type {
        IstioMesh istio = 1;
        AwsAppMesh aws_app_mesh = 2;
        LinkerdMesh linkerd = 3;
        ConsulConnectMesh consul_connect = 4;
    }
```

We propose the following API change:

```proto
    // I'm unsure what versions of anything other than Istio we support, so supplying bogus values; mainly pay attention to Istio
    oneof mesh_type {
        Istio15 istio_1_5 = 1;
        Istio16 istio_1_6 = 5;
        AwsAppMeshVersion aws_app_mesh = 2;
        LinkerdMesh20 linkerd = 3;
        ConsulConnectMeshVersion consul_connect = 4;
    }
```

Then in our code, we can have `switch` statements that look like:

```go
switch mesh.getType() {
case Istio15:
  // output Istio 1.5 resources
case Istio16:
  // output Istio 1.6 resources 
}
```

If there are no breaking changes from one mesh version to the next, we can easily reuse code:

```go
switch mesh.getType() {
case Istio15, Istio16:
  // output resources that work on both Istio 15 and 16
}
``` 

## Expected Concerns and Mitigating Solutions

Not sure there are too many risks with this approach. Since Go does not have complete pattern matching, there's always
a risk that we'll fail to update a switch statement somewhere when adding support for a new mesh type, but hopefully
thorough QA and e2e testing, as well as documentation, should alleviate that.
