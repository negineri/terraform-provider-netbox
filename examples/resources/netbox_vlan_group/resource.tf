resource "netbox_vlan_group" "example" {
  name        = "Example VLAN Group"
  slug        = "example-vlan-group"
  description = "Created via terraform-provider-netbox"
  min_vid     = 1
  max_vid     = 4094
  scope_type  = "dcim.site"
  scope_id    = 1
}

resource "netbox_vlan" "example" {
  name        = "Example VLAN"
  vid         = 100
  status      = "active"
  description = "Created via terraform-provider-netbox"
  group_id    = netbox_vlan_group.example.id
}
