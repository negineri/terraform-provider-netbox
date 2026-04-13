# Fetch a single virtual machine by ID.
data "netbox_virtual_machine" "example" {
  id = 1
}

output "virtual_machine_name" {
  value = data.netbox_virtual_machine.example.name
}

output "virtual_machine_status" {
  value = data.netbox_virtual_machine.example.status
}
