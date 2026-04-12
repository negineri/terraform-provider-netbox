resource "netbox_virtual_machine_primary_ip" "example" {
  virtual_machine_id = netbox_virtual_machine.example.id
  primary_ipv4_id    = netbox_ip_address.example_v4.id
  primary_ipv6_id    = netbox_ip_address.example_v6.id
}
