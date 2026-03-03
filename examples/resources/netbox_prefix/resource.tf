# This example demonstrates how to create a prefix and allocate an available IP address from it in Netbox using the terraform-provider-netbox.
resource "netbox_prefix" "test" {
  prefix      = "10.0.0.0/24"
  status      = "active"
  description = "Created via terraform-provider-netbox"
}

resource "netbox_available_ip" "test_ip" {
  prefix_id   = netbox_prefix.test.id
  status      = "active"
  description = "Allocated from terraform-created prefix"
}

output "prefix_id" {
  value = netbox_prefix.test.id
}

output "allocated_ip" {
  value = netbox_available_ip.test_ip.ip_address
}
