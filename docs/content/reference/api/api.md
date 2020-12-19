
---
title: "api.proto"
---

## Package : `google.protobuf`



<a name="top"></a>

<a name="API Reference for api.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## api.proto


## Table of Contents
  - [Api](#google.protobuf.Api)
  - [Method](#google.protobuf.Method)
  - [Mixin](#google.protobuf.Mixin)







<a name="google.protobuf.Api"></a>

### Api
Api is a light-weight descriptor for an API Interface.<br>Interfaces are also described as "protocol buffer services" in some contexts, such as by the "service" keyword in a .proto file, but they are different from API Services, which represent a concrete implementation of an interface as opposed to simply a description of methods and bindings. They are also sometimes simply referred to as "APIs" in other contexts, such as the name of this message itself. See https://cloud.google.com/apis/design/glossary for detailed terminology.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The fully qualified name of this interface, including package name followed by the interface's simple name. |
  | methods | [][google.protobuf.Method]({{< ref "api.md#google.protobuf.Method" >}}) | repeated | The methods of this interface, in unspecified order. |
  | options | [][google.protobuf.Option]({{< ref "type.md#google.protobuf.Option" >}}) | repeated | Any metadata attached to the interface. |
  | version | string |  | A version string for this interface. If specified, must have the form `major-version.minor-version`, as in `1.10`. If the minor version is omitted, it defaults to zero. If the entire version field is empty, the major version is derived from the package name, as outlined below. If the field is not empty, the version in the package name will be verified to be consistent with what is provided here.<br>The versioning schema uses [semantic versioning](http://semver.org) where the major version number indicates a breaking change and the minor version an additive, non-breaking change. Both version numbers are signals to users what to expect from different versions, and should be carefully chosen based on the product plan.<br>The major version is also reflected in the package name of the interface, which must end in `v<major-version>`, as in `google.feature.v1`. For major versions 0 and 1, the suffix can be omitted. Zero major versions must only be used for experimental, non-GA interfaces. |
  | sourceContext | [google.protobuf.SourceContext]({{< ref "source_context.md#google.protobuf.SourceContext" >}}) |  | Source context for the protocol buffer service represented by this message. |
  | mixins | [][google.protobuf.Mixin]({{< ref "api.md#google.protobuf.Mixin" >}}) | repeated | Included interfaces. See [Mixin][]. |
  | syntax | [google.protobuf.Syntax]({{< ref "type.md#google.protobuf.Syntax" >}}) |  | The source syntax of the service. |
  





<a name="google.protobuf.Method"></a>

### Method
Method represents a method of an API interface.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The simple name of this method. |
  | requestTypeUrl | string |  | A URL of the input message type. |
  | requestStreaming | bool |  | If true, the request is streamed. |
  | responseTypeUrl | string |  | The URL of the output message type. |
  | responseStreaming | bool |  | If true, the response is streamed. |
  | options | [][google.protobuf.Option]({{< ref "type.md#google.protobuf.Option" >}}) | repeated | Any metadata attached to the method. |
  | syntax | [google.protobuf.Syntax]({{< ref "type.md#google.protobuf.Syntax" >}}) |  | The source syntax of this method. |
  





<a name="google.protobuf.Mixin"></a>

### Mixin
Declares an API Interface to be included in this interface. The including interface must redeclare all the methods from the included interface, but documentation and options are inherited as follows:<br>- If after comment and whitespace stripping, the documentation   string of the redeclared method is empty, it will be inherited   from the original method.<br>- Each annotation belonging to the service config (http,   visibility) which is not set in the redeclared method will be   inherited.<br>- If an http annotation is inherited, the path pattern will be   modified as follows. Any version prefix will be replaced by the   version of the including interface plus the [root][] path if   specified.<br>Example of a simple mixin:<br>    package google.acl.v1;     service AccessControl {       // Get the underlying ACL object.       rpc GetAcl(GetAclRequest) returns (Acl) {         option (google.api.http).get = "/v1/{resource=**}:getAcl";       }     }<br>    package google.storage.v2;     service Storage {       rpc GetAcl(GetAclRequest) returns (Acl);<br>      // Get a data record.       rpc GetData(GetDataRequest) returns (Data) {         option (google.api.http).get = "/v2/{resource=**}";       }     }<br>Example of a mixin configuration:<br>    apis:     - name: google.storage.v2.Storage       mixins:       - name: google.acl.v1.AccessControl<br>The mixin construct implies that all methods in `AccessControl` are also declared with same name and request/response types in `Storage`. A documentation generator or annotation processor will see the effective `Storage.GetAcl` method after inherting documentation and annotations as follows:<br>    service Storage {       // Get the underlying ACL object.       rpc GetAcl(GetAclRequest) returns (Acl) {         option (google.api.http).get = "/v2/{resource=**}:getAcl";       }       ...     }<br>Note how the version in the path pattern changed from `v1` to `v2`.<br>If the `root` field in the mixin is specified, it should be a relative path under which inherited HTTP paths are placed. Example:<br>    apis:     - name: google.storage.v2.Storage       mixins:       - name: google.acl.v1.AccessControl         root: acls<br>This implies the following inherited HTTP annotation:<br>    service Storage {       // Get the underlying ACL object.       rpc GetAcl(GetAclRequest) returns (Acl) {         option (google.api.http).get = "/v2/acls/{resource=**}:getAcl";       }       ...     }


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | string |  | The fully qualified name of the interface which is included. |
  | root | string |  | If non-empty specifies a path under which inherited HTTP paths are rooted. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

