
---

---

## Package : `udpa.annotations`



<a name="top"></a>

<a name="API Reference for versioning.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## versioning.proto


## Table of Contents
  - [VersioningAnnotation](#udpa.annotations.VersioningAnnotation)


  - [File-level Extensions](#versioning.proto-extensions)





<a name="udpa.annotations.VersioningAnnotation"></a>

### VersioningAnnotation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| previousMessageType | string |  | Track the previous message type. E.g. this message might be udpa.foo.v3alpha.Foo and it was previously udpa.bar.v2.Bar. This information is consumed by UDPA via proto descriptors. |
  




 <!-- end messages -->

 <!-- end enums -->


<a name="versioning.proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| versioning | VersioningAnnotation | .google.protobuf.MessageOptions | 7881811 |  |

 <!-- end HasExtensions -->

 <!-- end services -->

