terraform {
  required_providers {
    netbox = {
      source = "hashicorp.com/edu/netbox"
    }
  }
}

provider "netbox" {
}

data "netbox_devices" "all" {}

output "first_device_name" {
  value = length(data.netbox_devices.all.devices) > 0 ? data.netbox_devices.all.devices[0].name : "No devices found"
}

output "device_count" {
  value = length(data.netbox_devices.all.devices)
}
