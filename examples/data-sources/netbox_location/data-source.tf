# Fetch a single location by ID.
data "netbox_location" "example" {
  id = 1
}

output "location_name" {
  value = data.netbox_location.example.name
}

output "location_status" {
  value = data.netbox_location.example.status
}

output "location_site_id" {
  value = data.netbox_location.example.site_id
}
