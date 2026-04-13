# Fetch a single device by ID.
data "netbox_device" "example" {
  id = 141
}

output "device_name" {
  value = data.netbox_device.example.name
}

output "device_status" {
  value = data.netbox_device.example.status
}
