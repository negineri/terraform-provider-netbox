# ID を指定してカスタムフィールドを1件取得する例。
data "netbox_custom_field" "example" {
  id = 1
}

output "custom_field_name" {
  value = data.netbox_custom_field.example.name
}

output "custom_field_type" {
  value = data.netbox_custom_field.example.type
}

output "custom_field_content_types" {
  value = data.netbox_custom_field.example.content_types
}
