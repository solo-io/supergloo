# Running Locally

Running locally should "just work"™ (run `RUN_E2E=1 ginkgo`).
If you want to use an existing cluster, you can run the tests like so:

```
RUN_E2E=1 USE_EXISTING=mgmt_cluster_ctx,remote_cluster_ctx ginkgo
```

Where `mgmt_cluster_ctx` and `remote_cluster_ctx` are the kube contexts for the management and remote clusters. This assumes that the clusters are already setup with service mesh hub.