
---

---

## Package : `udpa.annotations`



<a name="top"></a>

<a name="API Reference for migrate.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## migrate.proto


## Table of Contents
  - [FieldMigrateAnnotation](#udpa.annotations.FieldMigrateAnnotation)
  - [FileMigrateAnnotation](#udpa.annotations.FileMigrateAnnotation)
  - [MigrateAnnotation](#udpa.annotations.MigrateAnnotation)


  - [File-level Extensions](#migrate.proto-extensions)
  - [File-level Extensions](#migrate.proto-extensions)
  - [File-level Extensions](#migrate.proto-extensions)
  - [File-level Extensions](#migrate.proto-extensions)
  - [File-level Extensions](#migrate.proto-extensions)





<a name="udpa.annotations.FieldMigrateAnnotation"></a>

### FieldMigrateAnnotation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rename | string |  | Rename the field in next version. |
  | oneofPromotion | string |  | Add the field to a named oneof in next version. If this already exists, the field will join its siblings under the oneof, otherwise a new oneof will be created with the given name. |
  





<a name="udpa.annotations.FileMigrateAnnotation"></a>

### FileMigrateAnnotation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| moveToPackage | string |  | Move all types in the file to another package, this implies changing proto file path. |
  





<a name="udpa.annotations.MigrateAnnotation"></a>

### MigrateAnnotation



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| rename | string |  | Rename the message/enum/enum value in next version. |
  




 <!-- end messages -->

 <!-- end enums -->


<a name="migrate.proto-extensions"></a>

### File-level Extensions
| Extension | Type | Base | Number | Description |
| --------- | ---- | ---- | ------ | ----------- |
| enum_migrate | MigrateAnnotation | .google.protobuf.EnumOptions | 171962766 |  |
| enum_value_migrate | MigrateAnnotation | .google.protobuf.EnumValueOptions | 171962766 |  |
| field_migrate | FieldMigrateAnnotation | .google.protobuf.FieldOptions | 171962766 |  |
| file_migrate | FileMigrateAnnotation | .google.protobuf.FileOptions | 171962766 |  |
| message_migrate | MigrateAnnotation | .google.protobuf.MessageOptions | 171962766 |  |

 <!-- end HasExtensions -->

 <!-- end services -->

