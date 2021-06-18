
---

title: "string_match.proto"

---

## Package : `networking.mesh.gloo.solo.io`



<a name="top"></a>

<a name="API Reference for string_match.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## string_match.proto


## Table of Contents
  - [StringMatch](#networking.mesh.gloo.solo.io.StringMatch)







<a name="networking.mesh.gloo.solo.io.StringMatch"></a>

### StringMatch
Describes how to match a given string in HTTP headers. Match is case-sensitive.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| exact | string |  | Exact string match. |
  | prefix | string |  | Prefix-based match. |
  | regex | string |  | ECMAscript style regex-based match. |
  | suffix | string |  | Suffix-based match. |
  | ignoreCase | bool |  | If true, indicates the exact/prefix/suffix matching should be case insensitive. This has no effect for the regex match. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

