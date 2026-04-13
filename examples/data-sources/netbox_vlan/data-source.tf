# Fetch a single VLAN by ID.
data "netbox_vlan" "example" {
  id = 1
}

output "vlan_name" {
  value = data.netbox_vlan.example.name
}

output "vlan_vid" {
  value = data.netbox_vlan.example.vid
}
