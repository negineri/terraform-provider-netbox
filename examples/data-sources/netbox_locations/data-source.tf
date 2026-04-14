# List all locations.
data "netbox_locations" "all" {}

# Output the name of the first location, or a message if no locations are found.
output "first_location_name" {
  value = length(data.netbox_locations.all.locations) > 0 ? data.netbox_locations.all.locations[0].name : "No locations found"
}

# Output the total number of locations found.
output "location_count" {
  value = length(data.netbox_locations.all.locations)
}
