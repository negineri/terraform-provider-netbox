terraform {
  required_providers {
    netbox = {
      source = "hashicorp.com/edu/netbox"
    }
  }
}

provider "netbox" {
}

data "netbox_devices" "example" {}
