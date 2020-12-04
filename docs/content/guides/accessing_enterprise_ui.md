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

The Admin Dashboard runs on a pod in the Gloo Mesh Enterprise deployment and is exposed as a service. It does not have any authentication applied, so anyone with access to the Admin Dashboard can view the configuration and resources managed by the Gloo Mesh. That bears repeating:

{{< notice note >}}
Anyone who can reach the Admin Dashboard has unauthenticated access to view the configuration and resources managed by the Gloo Mesh.
{{< /notice >}}

Access to the Admin Dashboard should be restricted to only those who need to administer the Gloo Mesh. The `gloo-mesh-console` service is of the type ClusterIP, so it is not exposed outside of the cluster.

## Connecting to the Admin Dashboard

The Admin Dashboard is served from the gloo-mesh-console service on port 8090. You can connect using the `meshctl dashboard` command or by using the port-forward feature of kubectl. For this guide we will use port-forwarding. The following command assumes that you have deployed the Gloo Mesh to the namespace gloo-mesh. From a command prompt, run the following to set up port-forwarding for the gloo-mesh-console service.

kubectl port-forward -n gloo-mesh svc/gloo-mesh-console 8090:8090

Once the port-forwarding starts, you can open your browser and connect to http://localhost:8090. You will be taken to a webpage that looks similar to this:

![Gloo Mesh Admin Dashboard main page]({{% versioned_link_path fromRoot="/img/admin-main-page.png" %}})

Now that you’re connected, let’s explore the UI.

## Exploring the Admin Dashboard

The main page of the dashboard starts with an **Overview** of the resources under management of Gloo Mesh, such as *Clusters*, *Workloads*, and *Traffic Targets*.

![Gloo Mesh Admin Dashboard resources]({{% versioned_link_path fromRoot="/img/admin-resources.png" %}})

Across the top of the page is a navigation bar with five options.

![Gloo Mesh Admin Dashboard navigation]({{% versioned_link_path fromRoot="/img/admin-navigation.png" %}})

* **Overview**: Provides a high-level overview of Gloo Mesh
* **Meshes**: Displays service meshes being managed by Gloo Mesh
* **Policies**: Displays defined traffic and access policies for Gloo Mesh
* **Wasm**: Displays WASM deployments being managed by Gloo Mesh
* **Debug**: Displays full configurations for service meshes.

There is also a small gear to the right of the navigation elements, which will take you to the Admin area. 

![Gloo Mesh Admin Dashboard gear]({{% versioned_link_path fromRoot="/img/admin-gear.png" %}})

From there you are able to view clusters and Role-based access configurations.

![Gloo Mesh Admin area]({{% versioned_link_path fromRoot="/img/admin-view.png" %}})

The purpose of the Admin Dashboard is to view the status of Gloo Mesh and managed resources. It is not possible to make changes to resources or the configuration. Clicking on the **Register a Cluster** button simply provides you with directions on using meshctl to register a cluster.

### Meshes

The **Meshes** area provides a view of virtual meshes and each service mesh that is not part of a virtual mesh. You can view the health of each mesh, as well as information about traffic targets, workloads, failovers, and more.

![Gloo Mesh Admin Dashboard Meshes]({{% versioned_link_path fromRoot="/img/meshes-view.png" %}})

Clicking on a details link will provide in-depth information about each category associated with the mesh, and allows you to further drill down and view the configuration and associated resources for each element.

![Gloo Mesh Admin Dashboard Mesh Details]({{% versioned_link_path fromRoot="/img/meshes-details.png" %}})

By clicking on one of the traffic targets in the list, we can see more information about the target's configuration, policies, and any associated workloads.

![Gloo Mesh Admin Dashboard traffic target details]({{% versioned_link_path fromRoot="/img/traffic-target-details.png" %}})

### Policies

The **Policies** area allows us to explore the configured policy rules that have been created, and quickly assess what workloads, traffic targets, and meshes they are associated with.

![Gloo Mesh Admin Dashboard Policies Details]({{% versioned_link_path fromRoot="/img/policies-details.png" %}})

We can view additional detail about a policy rule by clicking on it.

![Gloo Mesh Admin Dashboard Rules Details]({{% versioned_link_path fromRoot="/img/rules-details.png" %}})

### Debug

The **Debug** section allows you to view and download the full configuration of any virtual meshes, as well as view details of each service mesh Gloo Mesh is aware of.

![Gloo Mesh Admin Dashboard Debug Details]({{% versioned_link_path fromRoot="/img/debug-details.png" %}})

You can use this view to quickly ascertain information about a particular mesh or to capture the current configuration of a virtual mesh.

## Next Steps

If your Admin Dashboard is looking a bit sparse, now might be a good time to walk through the [Istio installation]({{% versioned_link_path fromRoot="/guides/installing_istio/" %}}) or [traffic policy guides]({{% versioned_link_path fromRoot="/guides/traffic_policy/" %}}).
