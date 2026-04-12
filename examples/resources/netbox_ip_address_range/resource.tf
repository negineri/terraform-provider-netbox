# This example demonstrates how to create an IP range in Netbox using the terraform-provider-netbox.
resource "netbox_ip_address_range" "example" {
  ip_range    = "10.18.48.[224-239]/24"
  status      = "active"
  description = "Created via terraform-provider-netbox"
}

output "ip_range_id" {
  value = netbox_ip_address_range.example.id
}

output "start_address" {
  value = netbox_ip_address_range.example.start_address
}

output "end_address" {
  value = netbox_ip_address_range.example.end_address
}
