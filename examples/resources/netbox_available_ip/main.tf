terraform {
  required_providers {
    netbox = {
      source = "hashicorp.com/edu/netbox"
    }
  }
}

provider "netbox" {}

# Using prefix 1 for testing purposes. If it fails, I will change to another ID.
resource "netbox_available_ip" "test" {
  prefix_id   = 1
  description = "terraform test IP updated"
  status      = "active"
}

output "ip_address_output" {
  value = netbox_available_ip.test.ip_address
}
