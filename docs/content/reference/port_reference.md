---
title: "Gloo Mesh Port Reference"
description: Listing of ports used by Gloo Mesh and Gloo Mesh Enterprise
weight: 35
---

Gloo Mesh and Gloo Mesh Enterprise both deploy containers that listen on certain ports for incoming traffic. This document lists out the pods and services that make up Gloo Mesh and Gloo Mesh Enterprise, and the ports which these pods and services listen on. It is also possible to set up mutual TLS (mTLS) for communication between Gloo Mesh resources. The addition of mTLS changes the ports and traffic flows slightly, which is addressed in this document as well.

{{% notice note %}}
It is possible to customize some port settings by providing custom values to the Helm chart that installs Gloo Mesh open-source and Gloo Mesh Enterprise. The port reference below is for an installation of Gloo Mesh that uses the default settings in the Helm chart.
{{% /notice %}}

---

## Gloo Mesh Open-source

Gloo Mesh open-source software is the free, open-source version of Gloo Mesh. The installation process uses a Helm chart to create the necessary custom resource definitions (CRDs), deployments, services, pods, etc. The services and pods listen on specific ports to enable communication between the components that make up Gloo Mesh.

### What's included

A standard installation of Gloo Mesh includes four primary components:

* **Networking**
  * Translates gloo mesh discovery resources and user config to mesh config.
* **Discovery**
  * Discovers mesh related config on local and remote clusters.
  * Translate mesh related config to gloo mesh discovery resources.
* **Csr Agent**
  * Deployed on remote clusters.
  * Handles certitficate related tasks to ensure PKI doesn't leave the cluster.

### Pods and ports

The three primary components are instantiated using pods and services. The following table lists the deployed pods and ports in use by each pod.

| Pod | Port | Usage |
|-----|------|-------|
| gateway | 8443 | Validation |
| gloo | 9977 | xDS Server |
| gloo | 9988 | Validation |
| gloo | 9979 | WASM cache |
| gloo | 9966 | Metrics gRPC |
| gateway-proxy | 8080 | HTTP |
| gateway-proxy | 8443 | HTTPS |
| gateway-proxy | 19000 | Envoy admin |

### Services and ports

The following table lists the services backed by the deployed pods.

| Service | Port | Target | Target Port | Usage |
|---------|------|--------|-------------|-------|
| gateway | 443 | gateway | 8443 | Validation |
| gloo | 9977 | gloo | 9977 | xDS Server |
| gloo | 9988 | gloo | 9988 | Validation |
| gloo | 9979 | gloo | 9979 | WASM cache |
| gloo | 9966 | gloo | 9966 | Metrics gRPC |
| gateway-proxy | 80 | gateway-proxy | 8080 | HTTP |
| gateway-proxy | 443 | gateway-proxy | 8443 | HTTPS |
| access-log | 8083 | access-log | 8083 | Access logging |

---

## Gloo Mesh Enterprise

Gloo Mesh Enterprise changes the deployment model fairly signifigantly because of [relay](), as well as adding new deployments for additional functionality.

### What's included

At a high level, the following additional components are available in Gloo Mesh Enterprise.

* API and UI server
* Multi Cluster RBAC Webhook
* Prometheus metrics collection
* Prometheus server
* Grafana dashboard creation and presentation
* Redis

The Prometheus server and Grafana dashboard are optional components. If you have an existing instance of either, they can be used instead.

### Pods and ports

The Gloo Mesh Enterprise components are instantiated using pods and services. The following table lists the deployed pods and ports in use by each pod.

| Pod | Port | Usage |
|-----|------|-------|
| gloo-fed-console | 8090 | UI server |
| gloo-fed-console | 10101 | API Server |
| gloo-fed-console | 8081 | healthcheck |
| extauth | 8083 | External authentication |
| grafana | 80 | Grafana (unused) |
| grafana | 3000 | Grafana UI |
| prometheus-kube-state-metrics | 8080 | Kubernetes metric collection |
| prometheus-server | 9090 | Prometheus server |
| rate-limit | 18081 | Rate-limiting |
| redis | 6379 | Rate-limiting |

There is an `observability` pod that automatically configures dashboards on the Grafana instance. It does not accept inbound traffic, so it is not included in the table above.

### Services and ports

The following table lists the services backed by the deployed pods.

| Service | Port | Target | Target Port | Usage |
|---------|------|--------|-------------|-------|
| gloo-fed-console | 8090 | UI server |
| gloo-fed-console | 10101 | API Server |
| gloo-fed-console | 8081 | healthcheck |
| extauth | 8083 | extauth | 8083 | External authentication |
| glooe-grafana | 80 | grafana | 3000 | Grafana UI |
| glooe-prometheus-kube-state-metrics | 80 | prometheus-kube-state-metrics | 8080 | Kubernetes metric collection |
| glooe-prometheus-server | 80 | prometheus-server | 9090 | Prometheus server |
| rate-limit | 18081 | rate-limit | 18081 | Rate-limiting |
| redis | 6379 | redis | 6379 | Rate-limiting |
