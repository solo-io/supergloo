# Istio E2e Tests

The Istio End-to-End tests provide the most robust single test suite in the Gloo Mesh. 

These tests use `kind` and the `ci/setup-kind.sh` script to run tests against two KinD clusters which are configured 
with a VirtualMesh for federation between two Istio meshes.

A local invocation of these tests can be performed with:

```
make run-tests TEST_PKG=test/e2e/istio RUN_E2E=1
```

Note, if you plan on re-running the tests, it is recommended to run with `NO_CLEANUP=1` (on the first run) and `USE_EXISTING=1` for subsequent test runs.

The Full set of options for the tests can be found here:

|  env | description  |
|---|---|
| `RUN_E2E=1`  | Must be set or tests will not run. |
| `NO_CLEANUP=1`  | Spin up new clusters but do not tear them down when tests finish. |
| `USE_EXISTING=1`  | Use the existing `kind-mgmt-cluster` and `kind-remote-cluster` clusters. |
| `GINKGOFLAGS=<ginkgo flags>`  | Provide an optional set of flags for the `ginkgo` invocation used to run these tests. |

