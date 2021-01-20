
---

---

## Package : `envoy.type.v3`



<a name="top"></a>

<a name="API Reference for percent.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## percent.proto


## Table of Contents
  - [FractionalPercent](#envoy.type.v3.FractionalPercent)
  - [Percent](#envoy.type.v3.Percent)

  - [FractionalPercent.DenominatorType](#envoy.type.v3.FractionalPercent.DenominatorType)






<a name="envoy.type.v3.FractionalPercent"></a>

### FractionalPercent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| numerator | uint32 |  | Specifies the numerator. Defaults to 0. |
  | denominator | [envoy.type.v3.FractionalPercent.DenominatorType]({{< versioned_link_path fromRoot="/reference/api/envoy.type.v3.percent#envoy.type.v3.FractionalPercent.DenominatorType" >}}) |  | Specifies the denominator. If the denominator specified is less than the numerator, the final fractional percentage is capped at 1 (100%). |
  





<a name="envoy.type.v3.Percent"></a>

### Percent



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| value | double |  |  |
  




 <!-- end messages -->


<a name="envoy.type.v3.FractionalPercent.DenominatorType"></a>

### FractionalPercent.DenominatorType


| Name | Number | Description |
| ---- | ------ | ----------- |
| HUNDRED | 0 | 100.<br>**Example**: 1/100 = 1%. |
| TEN_THOUSAND | 1 | 10,000.<br>**Example**: 1/10000 = 0.01%. |
| MILLION | 2 | 1,000,000.<br>**Example**: 1/1000000 = 0.0001%. |


 <!-- end enums -->

 <!-- end HasExtensions -->

 <!-- end services -->

