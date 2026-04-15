resource "netbox_site" "example" {
  name   = "Example Site"
  slug   = "example-site"
  status = "active"
}

# Basic rack
resource "netbox_rack" "example" {
  name        = "Example Rack"
  site_id     = netbox_site.example.id
  status      = "active"
  u_height    = 42
  description = "Created via terraform-provider-netbox"
}

# Rack with facility ID and location
resource "netbox_location" "example" {
  name    = "Example Location"
  slug    = "example-location"
  site_id = netbox_site.example.id
  status  = "active"
}

resource "netbox_rack" "with_location" {
  name        = "Example Rack with Location"
  site_id     = netbox_site.example.id
  location_id = netbox_location.example.id
  status      = "active"
  facility_id = "RACK-001"
  u_height    = 48
  description = "Rack in specific location"
}
