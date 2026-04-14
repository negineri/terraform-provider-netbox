# 全カスタムフィールドを一覧取得する例。
data "netbox_custom_fields" "all" {}

# カスタムフィールドの件数を出力する。
output "custom_field_count" {
  value = length(data.netbox_custom_fields.all.custom_fields)
}

# 最初のカスタムフィールド名を出力する。
output "first_custom_field_name" {
  value = length(data.netbox_custom_fields.all.custom_fields) > 0 ? data.netbox_custom_fields.all.custom_fields[0].name : "No custom fields found"
}

# dcim.device に割り当てられたカスタムフィールドのみをフィルタリングする例。
locals {
  device_custom_fields = [
    for cf in data.netbox_custom_fields.all.custom_fields :
    cf if contains(cf.content_types, "dcim.device")
  ]
}

output "device_custom_field_names" {
  value = [for cf in local.device_custom_fields : cf.name]
}
