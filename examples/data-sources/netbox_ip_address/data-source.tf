# Fetch a single IP address by ID.
data "netbox_ip_address" "example" {
  id = 1
}

output "ip_address" {
  value = data.netbox_ip_address.example.ip_address
}

output "ip_status" {
  value = data.netbox_ip_address.example.status
}
