
---

---

## Package : `udpa.annotations`



<a name="top"></a>

<a name="API Reference for status.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## status.proto


## Table of Contents
  - [StatusAnnotation](#udpa.annotations.StatusAnnotation)

  - [PackageVersionStatus](#udpa.annotations.PackageVersionStatus)

  - [File-level Extensions](#status.proto-extensions)





<a name="udpa.annotations.StatusAnnotation"></a>

### StatusAnnotation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| workInProgress | bool |  | The entity is work-in-progress and subject to breaking changes. |
  | packageVersionStatus | [udpa.annotations.PackageVersionStatus]({{< versioned_link_path fromRoot="/reference/api/github.com.cncf.udpa.udpa.annotations.status#udpa.annotations.PackageVersionStatus" >}}) |  | The entity belongs to a package with the given version status. |
  




 <!-- end messages -->


<a name="udpa.annotations.PackageVersionStatus"></a>

### PackageVersionStatus


| Name | Number | Description |
| ---- | ------ | ----------- |
| UNKNOWN | 0 | Unknown package version status. |
| FROZEN | 1 | This version of the package is frozen. |
| ACTIVE | 2 | This version of the package is the active development version. |
| NEXT_MAJOR_VERSION_CANDIDATE | 3 | This version of the package is the candidate for the next major version. It is typically machine generated from the active development version. |


 <!-- end enums -->


<a name="status.proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| file_status | StatusAnnotation | .google.protobuf.FileOptions | 222707719 |  |

 <!-- end HasExtensions -->

 <!-- end services -->

