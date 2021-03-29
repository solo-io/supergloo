
---
title: "Enterprise Agent"
description: Reference for Helm values. 
weight: 2
---

|Option|Type|Default Value|Description|
|------|----|-----------|-------------|
|relay|struct| |options for connecting relay|
|relay.cluster|string||cluster identifier for the relay agent|
|relay.serverAddress|string||address of the relay server|
|relay.authority|string|enterprise-networking.gloo-mesh|set the authority/host header to this value when dialing the Relay gRPC Server|
|relay.insecure|bool|false|communicate with relay server over plain HTTP|
|relay.clientCertSecret|struct| |Reference to a Secret containing the Client TLS Certificates used to identify the Relay Agent to the Server. If the secret does not exist, a Token and Root cert secret are required.|
|relay.clientCertSecret.name|string|relay-client-tls-secret||
|relay.clientCertSecret.namespace|string|||
|relay.rootTlsSecret|struct| |Reference to a Secret containing a Root TLS Certificates used to verify the Relay Server Certificate. The secret can also optionally specify a 'tls.key' which will be used to generate the Agent Client Certificate.|
|relay.rootTlsSecret.name|string|relay-root-tls-secret||
|relay.rootTlsSecret.namespace|string|||
|relay.tokenSecret|struct| |Reference to a Secret containing a shared Token for authenticating to the Relay Server|
|relay.tokenSecret.name|string|relay-identity-token-secret|Name of the Kubernetes Secret|
|relay.tokenSecret.namespace|string||Namespace of the Kubernetes Secret|
|relay.tokenSecret.key|string|token|Key value of the data within the Kubernetes Secret|
|relay.maxGrpcMessageSize|string|4294967295|Specify to set a custom maximum message size for grpc messages sent to the Relay server|
|settingsRef|struct| |ref to the settings object that will be received from the networking server.|
|settingsRef.name|string|settings||
|settingsRef.namespace|string|gloo-mesh||
|verbose|bool|false||
