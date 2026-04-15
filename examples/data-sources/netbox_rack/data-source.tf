# Fetch a single rack by ID.
data "netbox_rack" "example" {
  id = 1
}

output "rack_name" {
  value = data.netbox_rack.example.name
}

output "rack_status" {
  value = data.netbox_rack.example.status
}

output "rack_site_id" {
  value = data.netbox_rack.example.site_id
}

output "rack_u_height" {
  value = data.netbox_rack.example.u_height
}
