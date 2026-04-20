# Fetch a single platform by ID.
data "netbox_platform" "example" {
  id = 1
}

output "platform_name" {
  value = data.netbox_platform.example.name
}

output "platform_slug" {
  value = data.netbox_platform.example.slug
}
