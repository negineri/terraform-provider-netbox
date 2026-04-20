resource "netbox_tenant" "example" {
  name        = "Example Tenant"
  slug        = "example-tenant"
  description = "Created via terraform-provider-netbox"
}
