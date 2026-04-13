# Fetch a single site by ID.
data "netbox_site" "example" {
  id = 1
}

output "site_name" {
  value = data.netbox_site.example.name
}

output "site_status" {
  value = data.netbox_site.example.status
}
