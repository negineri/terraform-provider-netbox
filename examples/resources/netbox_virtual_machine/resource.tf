resource "netbox_virtual_machine" "example" {
  name        = "vm-example"
  cluster_id  = 1
  status      = "active"
  vcpus       = 4
  memory      = 8192
  disk        = 100
  description = "Created via terraform-provider-netbox"
  tags        = [1]
}

# A virtual machine must be assigned to a site and/or cluster.
