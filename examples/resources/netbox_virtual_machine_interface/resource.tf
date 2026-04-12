resource "netbox_virtual_machine_interface" "example" {
  virtual_machine_id = netbox_virtual_machine.example.id
  name               = "eth0"
  enabled            = true
  description        = "Created via terraform-provider-netbox"
}
