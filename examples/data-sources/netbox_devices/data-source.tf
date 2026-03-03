# List all devices.
data "netbox_devices" "all" {}

# Output the name of the first device, or a message if no devices are found.
output "first_device_name" {
  value = length(data.netbox_devices.all.devices) > 0 ? data.netbox_devices.all.devices[0].name : "No devices found"
}

# Output the total number of devices found.
output "device_count" {
  value = length(data.netbox_devices.all.devices)
}
