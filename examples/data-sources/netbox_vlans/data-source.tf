# List all VLANs.
data "netbox_vlans" "all" {}

# Output the name of the first VLAN, or a message if no VLANs are found.
output "first_vlan_name" {
  value = length(data.netbox_vlans.all.vlans) > 0 ? data.netbox_vlans.all.vlans[0].name : "No VLANs found"
}

# Output the total number of VLANs found.
output "vlan_count" {
  value = length(data.netbox_vlans.all.vlans)
}
