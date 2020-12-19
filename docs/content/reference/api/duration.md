
---
title: "duration.proto"
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
A Duration represents a signed, fixed-length span of time represented as a count of seconds and fractions of seconds at nanosecond resolution. It is independent of any calendar and concepts like "day" or "month". It is related to Timestamp in that the difference between two Timestamp values is a Duration and it can be added or subtracted from a Timestamp. Range is approximately +-10,000 years.<br># Examples<br>Example 1: Compute Duration from two Timestamps in pseudo code.<br>    Timestamp start = ...;     Timestamp end = ...;     Duration duration = ...;<br>    duration.seconds = end.seconds - start.seconds;     duration.nanos = end.nanos - start.nanos;<br>    if (duration.seconds < 0 && duration.nanos > 0) {       duration.seconds += 1;       duration.nanos -= 1000000000;     } else if (durations.seconds > 0 && duration.nanos < 0) {       duration.seconds -= 1;       duration.nanos += 1000000000;     }<br>Example 2: Compute Timestamp from Timestamp + Duration in pseudo code.<br>    Timestamp start = ...;     Duration duration = ...;     Timestamp end = ...;<br>    end.seconds = start.seconds + duration.seconds;     end.nanos = start.nanos + duration.nanos;<br>    if (end.nanos < 0) {       end.seconds -= 1;       end.nanos += 1000000000;     } else if (end.nanos >= 1000000000) {       end.seconds += 1;       end.nanos -= 1000000000;     }<br>Example 3: Compute Duration from datetime.timedelta in Python.<br>    td = datetime.timedelta(days=3, minutes=10)     duration = Duration()     duration.FromTimedelta(td)<br># JSON Mapping<br>In JSON format, the Duration type is encoded as a string rather than an object, where the string ends in the suffix "s" (indicating seconds) and is preceded by the number of seconds, with nanoseconds expressed as fractional seconds. For example, 3 seconds with 0 nanoseconds should be encoded in JSON format as "3s", while 3 seconds and 1 nanosecond should be expressed in JSON format as "3.000000001s", and 3 seconds and 1 microsecond should be expressed in JSON format as "3.000001s".


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| seconds | int64 |  | Signed seconds of the span of time. Must be from -315,576,000,000 to +315,576,000,000 inclusive. Note: these bounds are computed from: 60 sec/min * 60 min/hr * 24 hr/day * 365.25 days/year * 10000 years |
  | nanos | int32 |  | Signed fractions of a second at nanosecond resolution of the span of time. Durations less than one second are represented with a 0 `seconds` field and a positive or negative `nanos` field. For durations of one second or more, a non-zero value for the `nanos` field must be of the same sign as the `seconds` field. Must be from -999,999,999 to +999,999,999 inclusive. |
  




 <!-- end messages -->

 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

