# App Mesh demo application
The `bookinfo.yaml` file contains a manifest for the [Istio Bookinfo](https://istio.io/docs/examples/bookinfo/) 
application modified to work on EKS with Aws App Mesh.

## Deploy the application

---
**NOTE**: Currently you HAVE to register the App Mesh mesh CRD in the same namespace supergloo is deployed to.

---

To get the application running:
1. Run `supergloo init`.
2. Create an AWS secret with `supergloo create secret aws -i`. The secret must reference AWS credentials associated with 
a role that has the permissions to CRUD App Mesh resources.
3. Register an App Mesh instance using `supergloo register appmesh -i`. When prompted, enable auto-injection for the mesh 
and set the value of the `VirtualNodeLabel` to `vn-name`. Supergloo looks for a label with that key on auto-injected pods 
to determine which `VirtualNode` they belong to. Be sure to place the mesh in the same namespace supergloo is deployed to.
4. `kubectl -f apply` the `bookinfo.yaml` file.
5. Create a routing rule to split traffic for v1 of the `reviews` service between v1, v2, and v3.
6. To clean up, delete the routing rule(s) and the mesh CRD instance.
7 `port-forward` the `productpage` service to verify that the mesh is configured as intended.

## Cleanup AWS API resources
Run `./cleanup.sh MESH_NAME` to delete the Aws App Mesh control plane resources. You will probably need to run the 
script 2 times, as some resources will fail to delete on the first run due to referential constraints being violated.