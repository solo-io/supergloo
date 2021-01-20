
---

---

## Package : `envoy.type.matcher.v3`



<a name="top"></a>

<a name="API Reference for regex.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## regex.proto


## Table of Contents
  - [RegexMatchAndSubstitute](#envoy.type.matcher.v3.RegexMatchAndSubstitute)
  - [RegexMatcher](#envoy.type.matcher.v3.RegexMatcher)
  - [RegexMatcher.GoogleRE2](#envoy.type.matcher.v3.RegexMatcher.GoogleRE2)







<a name="envoy.type.matcher.v3.RegexMatchAndSubstitute"></a>

### RegexMatchAndSubstitute



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| pattern | [envoy.type.matcher.v3.RegexMatcher]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.regex#envoy.type.matcher.v3.RegexMatcher" >}}) |  | The regular expression used to find portions of a string (hereafter called the "subject string") that should be replaced. When a new string is produced during the substitution operation, the new string is initially the same as the subject string, but then all matches in the subject string are replaced by the substitution string. If replacing all matches isn't desired, regular expression anchors can be used to ensure a single match, so as to replace just one occurrence of a pattern. Capture groups can be used in the pattern to extract portions of the subject string, and then referenced in the substitution string. |
  | substitution | string |  | The string that should be substituted into matching portions of the subject string during a substitution operation to produce a new string. Capture groups in the pattern can be referenced in the substitution string. Note, however, that the syntax for referring to capture groups is defined by the chosen regular expression engine. Google's `RE2 <https://github.com/google/re2>`_ regular expression engine uses a backslash followed by the capture group number to denote a numbered capture group. E.g., ``\1`` refers to capture group 1, and ``\2`` refers to capture group 2. |
  





<a name="envoy.type.matcher.v3.RegexMatcher"></a>

### RegexMatcher



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| googleRe2 | [envoy.type.matcher.v3.RegexMatcher.GoogleRE2]({{< versioned_link_path fromRoot="/reference/api/envoy.type.matcher.v3.regex#envoy.type.matcher.v3.RegexMatcher.GoogleRE2" >}}) |  | Google's RE2 regex engine. |
  | regex | string |  | The regex match string. The string must be supported by the configured engine. |
  





<a name="envoy.type.matcher.v3.RegexMatcher.GoogleRE2"></a>

### RegexMatcher.GoogleRE2



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| maxProgramSize | [google.protobuf.UInt32Value]({{< versioned_link_path fromRoot="/reference/api/github.com.solo-io.protoc-gen-ext.external.google.protobuf.wrappers#google.protobuf.UInt32Value" >}}) |  | This field controls the RE2 "program size" which is a rough estimate of how complex a compiled regex is to evaluate. A regex that has a program size greater than the configured value will fail to compile. In this case, the configured max program size can be increased or the regex can be simplified. If not specified, the default is 100.<br>This field is deprecated; regexp validation should be performed on the management server instead of being done by each individual client. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

