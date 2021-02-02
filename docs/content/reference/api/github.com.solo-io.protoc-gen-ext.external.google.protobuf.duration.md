
---

---

## Package : `google.protobuf`



<a name="top"></a>

<a name="API Reference for duration.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## duration.proto


## Table of Contents
  - [Duration](#google.protobuf.Duration)







<a name="google.protobuf.Duration"></a>

### Duration



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| seconds | int64 |  | Signed seconds of the span of time. Must be from -315,576,000,000 to +315,576,000,000 inclusive. Note: these bounds are computed from: 60 sec/min * 60 min/hr * 24 hr/day * 365.25 days/year * 10000 years |
  | nanos | int32 |  | Signed fractions of a second at nanosecond resolution of the span of time. Durations less than one second are represented with a 0 `seconds` field and a positive or negative `nanos` field. For durations of one second or more, a non-zero value for the `nanos` field must be of the same sign as the `seconds` field. Must be from -999,999,999 to +999,999,999 inclusive. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

