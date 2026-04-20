# List all tenants.
data "netbox_tenants" "all" {}

# Filter by name.
data "netbox_tenants" "filtered" {
  name = "Example Tenant"
}

# Output the name of the first tenant, or a message if no tenants are found.
output "first_tenant_name" {
  value = length(data.netbox_tenants.all.tenants) > 0 ? data.netbox_tenants.all.tenants[0].name : "No tenants found"
}

# Output the total number of tenants found.
output "tenant_count" {
  value = length(data.netbox_tenants.all.tenants)
}
