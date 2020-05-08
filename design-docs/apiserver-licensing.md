# API Server Licensing

*Engineer:* Graham Goudeau

*Domain Area:* API Server Licensing

*Issue Link:* https://github.com/solo-io/service-mesh-hub/issues/659

## Problem Summary

Need to gate mutating API Server operations on whether or not a valid license for Service Mesh Hub exists in the cluster.

## Proposed Design

The design has three main pieces:

* An API Server endpoint to tell the UI whether or not a valid license exists
* gRPC middleware on the server for implementing a license check at request time, and allowing through a whitelist of read-only methods
* When a request is rejected because of a lack of a valid license, return gRPC status code `FAILED_PRECONDITION`
(status code `9`) to indicate that the precondition (a valid license) to executing that requested method was not satisfied 

## Expected Concerns and Mitigating Solutions

Reserving a status code always runs the risk of becoming a pain if we need that code later, but `FAILED_PRECONDITION` seems
sufficiently unlikely to be needed by us later.
