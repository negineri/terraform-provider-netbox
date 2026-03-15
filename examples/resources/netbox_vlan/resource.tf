resource "netbox_vlan" "example" {
  name        = "Example VLAN"
  vid         = 100
  status      = "active"
  description = "Created via terraform-provider-netbox"
}
