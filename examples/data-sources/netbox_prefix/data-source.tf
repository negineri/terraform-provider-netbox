# Fetch a single prefix by ID.
data "netbox_prefix" "example" {
  id = 1
}

output "prefix_cidr" {
  value = data.netbox_prefix.example.prefix
}

output "prefix_status" {
  value = data.netbox_prefix.example.status
}
