# List all IP addresses.
data "netbox_ip_addresses" "all" {}

# Output the first IP address, or a message if no IP addresses are found.
output "first_ip_address" {
  value = length(data.netbox_ip_addresses.all.ip_addresses) > 0 ? data.netbox_ip_addresses.all.ip_addresses[0].ip_address : "No IP addresses found"
}

# Output the total number of IP addresses found.
output "ip_address_count" {
  value = length(data.netbox_ip_addresses.all.ip_addresses)
}
