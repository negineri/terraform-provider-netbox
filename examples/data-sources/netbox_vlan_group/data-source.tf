# Fetch a single VLAN group by ID.
data "netbox_vlan_group" "example" {
  id = 1
}

output "vlan_group_name" {
  value = data.netbox_vlan_group.example.name
}

output "vlan_group_slug" {
  value = data.netbox_vlan_group.example.slug
}
