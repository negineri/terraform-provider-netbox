# List all virtual machines.
data "netbox_virtual_machines" "all" {}

# Output the name of the first virtual machine, or a message if no virtual machines are found.
output "first_virtual_machine_name" {
  value = length(data.netbox_virtual_machines.all.virtual_machines) > 0 ? data.netbox_virtual_machines.all.virtual_machines[0].name : "No virtual machines found"
}

# Output the total number of virtual machines found.
output "virtual_machine_count" {
  value = length(data.netbox_virtual_machines.all.virtual_machines)
}
