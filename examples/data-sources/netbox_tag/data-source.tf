# Fetch a single tag by ID.
data "netbox_tag" "example" {
  id = 1
}

output "tag_name" {
  value = data.netbox_tag.example.name
}

output "tag_color" {
  value = data.netbox_tag.example.color
}
