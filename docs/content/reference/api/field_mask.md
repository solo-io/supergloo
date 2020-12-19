
---
title: "field_mask.proto"
---

## Package : `google.protobuf`



<a name="top"></a>

<a name="API Reference for field_mask.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## field_mask.proto


## Table of Contents
  - [FieldMask](#google.protobuf.FieldMask)







<a name="google.protobuf.FieldMask"></a>

### FieldMask
`FieldMask` represents a set of symbolic field paths, for example:<br>    paths: "f.a"     paths: "f.b.d"<br>Here `f` represents a field in some root message, `a` and `b` fields in the message found in `f`, and `d` a field found in the message in `f.b`.<br>Field masks are used to specify a subset of fields that should be returned by a get operation or modified by an update operation. Field masks also have a custom JSON encoding (see below).<br># Field Masks in Projections<br>When used in the context of a projection, a response message or sub-message is filtered by the API to only contain those fields as specified in the mask. For example, if the mask in the previous example is applied to a response message as follows:<br>    f {       a : 22       b {         d : 1         x : 2       }       y : 13     }     z: 8<br>The result will not contain specific values for fields x,y and z (their value will be set to the default, and omitted in proto text output):<br>     f {       a : 22       b {         d : 1       }     }<br>A repeated field is not allowed except at the last position of a paths string.<br>If a FieldMask object is not present in a get operation, the operation applies to all fields (as if a FieldMask of all fields had been specified).<br>Note that a field mask does not necessarily apply to the top-level response message. In case of a REST get operation, the field mask applies directly to the response, but in case of a REST list operation, the mask instead applies to each individual message in the returned resource list. In case of a REST custom method, other definitions may be used. Where the mask applies will be clearly documented together with its declaration in the API.  In any case, the effect on the returned resource/resources is required behavior for APIs.<br># Field Masks in Update Operations<br>A field mask in update operations specifies which fields of the targeted resource are going to be updated. The API is required to only change the values of the fields as specified in the mask and leave the others untouched. If a resource is passed in to describe the updated values, the API ignores the values of all fields not covered by the mask.<br>If a repeated field is specified for an update operation, the existing repeated values in the target resource will be overwritten by the new values. Note that a repeated field is only allowed in the last position of a `paths` string.<br>If a sub-message is specified in the last position of the field mask for an update operation, then the existing sub-message in the target resource is overwritten. Given the target message:<br>    f {       b {         d : 1         x : 2       }       c : 1     }<br>And an update message:<br>    f {       b {         d : 10       }     }<br>then if the field mask is:<br> paths: "f.b"<br>then the result will be:<br>    f {       b {         d : 10       }       c : 1     }<br>However, if the update mask was:<br> paths: "f.b.d"<br>then the result would be:<br>    f {       b {         d : 10         x : 2       }       c : 1     }<br>In order to reset a field's value to the default, the field must be in the mask and set to the default value in the provided resource. Hence, in order to reset all fields of a resource, provide a default instance of the resource and set all fields in the mask, or do not provide a mask as described below.<br>If a field mask is not present on update, the operation applies to all fields (as if a field mask of all fields has been specified). Note that in the presence of schema evolution, this may mean that fields the client does not know and has therefore not filled into the request will be reset to their default. If this is unwanted behavior, a specific service may require a client to always specify a field mask, producing an error if not.<br>As with get operations, the location of the resource which describes the updated values in the request message depends on the operation kind. In any case, the effect of the field mask is required to be honored by the API.<br>## Considerations for HTTP REST<br>The HTTP kind of an update operation which uses a field mask must be set to PATCH instead of PUT in order to satisfy HTTP semantics (PUT must only be used for full updates).<br># JSON Encoding of Field Masks<br>In JSON, a field mask is encoded as a single string where paths are separated by a comma. Fields name in each path are converted to/from lower-camel naming conventions.<br>As an example, consider the following message declarations:<br>    message Profile {       User user = 1;       Photo photo = 2;     }     message User {       string display_name = 1;       string address = 2;     }<br>In proto a field mask for `Profile` may look as such:<br>    mask {       paths: "user.display_name"       paths: "photo"     }<br>In JSON, the same mask is represented as below:<br>    {       mask: "user.displayName,photo"     }<br># Field Masks and Oneof Fields<br>Field masks treat fields in oneofs just as regular fields. Consider the following message:<br>    message SampleMessage {       oneof test_oneof {         string name = 4;         SubMessage sub_message = 9;       }     }<br>The field mask can be:<br>    mask {       paths: "name"     }<br>Or:<br>    mask {       paths: "sub_message"     }<br>Note that oneof type names ("test_oneof" in this case) cannot be used in paths.<br>## Field Mask Verification<br>The implementation of any API method which has a FieldMask type field in the request should verify the included field paths, and return an `INVALID_ARGUMENT` error if any path is duplicated or unmappable.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| paths | []string | repeated | The set of field mask paths. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

