resource "netbox_platform" "example" {
  name        = "Example Platform"
  slug        = "example-platform"
  description = "Created via terraform-provider-netbox"
}

# With manufacturer association
resource "netbox_manufacturer" "example" {
  name = "Example Manufacturer"
  slug = "example-manufacturer"
}

resource "netbox_platform" "with_manufacturer" {
  name            = "Example Platform with Manufacturer"
  slug            = "example-platform-with-manufacturer"
  manufacturer_id = netbox_manufacturer.example.id
}
