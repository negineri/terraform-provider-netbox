resource "netbox_device" "example" {
  name           = "router-01"
  device_type_id = 1
  role_id        = 1
  site_id        = 1
  status         = "active"
  description    = "Created via terraform-provider-netbox"
}
