terraform {
  required_providers {
    netbox = {
      source = "hashicorp.com/edu/netbox"
    }
  }
}

provider "netbox" {}

resource "netbox_ip_address" "test" {
  ip_address  = "10.10.10.10/24"
  status      = "active"
  description = "Created explicitly via terraform-provider-netbox"
}

output "created_ip" {
  value = netbox_ip_address.test.ip_address
}

output "created_ip_id" {
  value = netbox_ip_address.test.id
}
