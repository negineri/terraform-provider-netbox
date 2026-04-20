# Fetch a single tenant by ID.
data "netbox_tenant" "example" {
  id = 1
}

output "tenant_name" {
  value = data.netbox_tenant.example.name
}

output "tenant_slug" {
  value = data.netbox_tenant.example.slug
}
