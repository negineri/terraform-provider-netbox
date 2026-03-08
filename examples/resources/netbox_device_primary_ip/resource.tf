resource "netbox_device_primary_ip" "example" {
  device_id       = netbox_device.example.id
  primary_ipv4_id = netbox_ip_address.example_v4.id
  primary_ipv6_id = netbox_ip_address.example_v6.id
}
