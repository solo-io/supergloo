---
title: Extending the Mesh with Wasm
menuTitle: Extending the Mesh with Wasm
weight: 85
---

{{% notice note %}}
Gloo Mesh Enterprise is required for this feature.
{{% /notice %}}

[WebAssembly](https://webassembly.org/) (Wasm) is the future of cloud-native infrastructure extensibility. Wasm provides a safe, secure, and 
dynamic way of extending infrastructure with the programming language of your choice. While feature-rich service 
meshes like Istio provide myriad features to help solve problems related to microservice communication, organizations
inevitably find themselves needing custom solutions tailored specifically to their individual constraints. Gloo Mesh Enterprise 
packages all the tooling you need to develop, publish, and deploy the Wasm extensions your business needs to meet its unique service mesh requirements.

With the Gloo Mesh Enterprise CLI, you can initialize, build, and push proprietary Wasm filters. Choose your preferred 
programming language (currently C++, Rust, Go, and AssemblyScript are supported) and we’ll generate all the source code 
you need to get started implementing custom mesh behavior. To publish your work use the `build` and `push` commands. 
These will compile your Wasm module and make it available via webassemblyhub.io or the OCI registry of your choice.

To add your new Wasm filter to the mesh, all you need is a `WasmDeployment` Kubernetes custom resource. Specify which 
Workloads should be configured and with which Wasm filters, then let Gloo Mesh handle the rest. A Gloo Mesh Enterprise 
extension server will watch for WasmDeployments and manage the lifecycle of all your Wasm filters accordingly.

To learn more about Wasm and Gloo Mesh Enterprise, request a demo at [solo.io](https://www.solo.io/).