# List all VLAN groups.
data "netbox_vlan_groups" "all" {}

# Output the name of the first VLAN group, or a message if no VLAN groups are found.
output "first_vlan_group_name" {
  value = length(data.netbox_vlan_groups.all.vlan_groups) > 0 ? data.netbox_vlan_groups.all.vlan_groups[0].name : "No VLAN groups found"
}

# Output the total number of VLAN groups found.
output "vlan_group_count" {
  value = length(data.netbox_vlan_groups.all.vlan_groups)
}
