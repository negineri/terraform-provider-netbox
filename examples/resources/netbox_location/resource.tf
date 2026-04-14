resource "netbox_site" "example" {
  name   = "Example Site"
  slug   = "example-site"
  status = "active"
}

# Basic location
resource "netbox_location" "example" {
  name        = "Example Location"
  slug        = "example-location"
  site_id     = netbox_site.example.id
  status      = "active"
  description = "Created via terraform-provider-netbox"
}

# Hierarchical location with parent
resource "netbox_location" "child" {
  name      = "Example Child Location"
  slug      = "example-child-location"
  site_id   = netbox_site.example.id
  parent_id = netbox_location.example.id
  status    = "active"
}
