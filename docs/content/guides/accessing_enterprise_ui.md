---
title: Using the Admin Dashboard
menuTitle: Using the Admin Dashboard
description: "How to access and use the Gloo Mesh Admin Dashboard."
weight: 110
---

{{< notice note >}}
This feature is available in Gloo Mesh Enterprise only. If you are using the open source version of Gloo Mesh, this tutorial will not work.
{{< /notice >}}

When you install Gloo Mesh Enterprise, it includes the Admin Dashboard service by default. The service provides a visual dashboard into the health and configuration of Gloo Mesh and registered clusters.

In this guide, you will learn how to connect to the Admin Dashboard and the basic layout of the portal’s contents.

## About the Admin Dashboard

The Admin Dashboard runs on a pod in the Gloo Mesh Enterprise deployment and is exposed as a service. It does not have any authentication applied, so anyone with access to the Admin Dashboard can make changes to the resources managed by the Gloo Mesh. That bears repeating:

{{< notice note >}}
Anyone who can reach the Admin Dashboard has unauthenticated access to make changes to resources managed by the Gloo Mesh.
{{< /notice >}}

Access to the Admin Dashboard should be restricted to only those who need to administer the Gloo Mesh. While the functions available in the dashboard are limited in nature today, more functionality will be added in the future. The `gloo-mesh-console` service is of the type ClusterIP, so it is not exposed outside of the cluster.

## Connecting to the Admin Dashboard

The Admin Dashboard is served from the gloo-mesh-console service on port 8090. For this guide we are going to connect using the port-forward feature of kubectl. The following command assumes that you have deployed the Gloo Mesh to the namespace gloo-mesh. From a command prompt, run the following to set up port-forwarding for the gloo-mesh-console service.

kubectl port-forward -n gloo-mesh svc/gloo-mesh-console 8090:8090

Once the port-forwarding starts, you can open your browser and connect to http://localhost:8090. You will be taken to a webpage that looks similar to this:

![Gloo Mesh Admin Dashboard main page]({{% versioned_link_path fromRoot="/img/admin-main-page.png" %}})

Now that you’re connected, let’s explore the UI.

## Exploring the Admin Dashboard

The main page of the dashboard starts with an **Overview** of the resources under management of Gloo Mesh, such as *Clusters*, *Workloads*, and *Traffic Targets*.

![Gloo Mesh Admin Dashboard resources]({{% versioned_link_path fromRoot="/img/admin-resources.png" %}})