terraform {
  required_providers {
    netbox = {
      source = "hashicorp.com/edu/netbox"
    }
  }
}

provider "netbox" {}

# Before running this, please create an IP range in your Netbox server
# and replace the `ip_range_id` below with the actual ID.
# Assuming range ID 1 for now.
resource "netbox_available_ip" "test_range_ip" {
  ip_range_id = 1
  status      = "active"
  description = "Allocated from terraform-created ip range"
}

output "allocated_ip_from_range" {
  value = netbox_available_ip.test_range_ip.ip_address
}
