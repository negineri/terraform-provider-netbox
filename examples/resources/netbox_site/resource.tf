resource "netbox_site" "example" {
  name             = "Example Site"
  slug             = "example-site"
  status           = "active"
  description      = "Created via terraform-provider-netbox"
  facility         = "Example Datacenter"
  time_zone        = "Asia/Tokyo"
  physical_address = "1-1-1 Example Street, Tokyo"
}

resource "netbox_vlan_group" "example" {
  name       = "Example VLAN Group"
  slug       = "example-vlan-group"
  scope_type = "dcim.site"
  scope_id   = netbox_site.example.id
}
