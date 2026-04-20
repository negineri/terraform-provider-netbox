# List all platforms.
data "netbox_platforms" "all" {}

# Filter by name.
data "netbox_platforms" "filtered" {
  name = "Example Platform"
}

# Output the name of the first platform, or a message if no platforms are found.
output "first_platform_name" {
  value = length(data.netbox_platforms.all.platforms) > 0 ? data.netbox_platforms.all.platforms[0].name : "No platforms found"
}

# Output the total number of platforms found.
output "platform_count" {
  value = length(data.netbox_platforms.all.platforms)
}
