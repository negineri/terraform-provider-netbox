resource "netbox_device_interface" "example" {
  device_id   = netbox_device.example.id
  name        = "lo"
  type        = "virtual"
  enabled     = true
  description = "Created via terraform-provider-netbox"
}
