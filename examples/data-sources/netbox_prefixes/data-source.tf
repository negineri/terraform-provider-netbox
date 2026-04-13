# List all prefixes.
data "netbox_prefixes" "all" {}

# Output the first prefix in CIDR notation, or a message if no prefixes are found.
output "first_prefix" {
  value = length(data.netbox_prefixes.all.prefixes) > 0 ? data.netbox_prefixes.all.prefixes[0].prefix : "No prefixes found"
}

# Output the total number of prefixes found.
output "prefix_count" {
  value = length(data.netbox_prefixes.all.prefixes)
}
