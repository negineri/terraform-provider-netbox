# List all racks.
data "netbox_racks" "all" {}

# Output the name of the first rack, or a message if no racks are found.
output "first_rack_name" {
  value = length(data.netbox_racks.all.racks) > 0 ? data.netbox_racks.all.racks[0].name : "No racks found"
}

# Output the total number of racks found.
output "rack_count" {
  value = length(data.netbox_racks.all.racks)
}
