# List all sites.
data "netbox_sites" "all" {}

# Output the name of the first site, or a message if no sites are found.
output "first_site_name" {
  value = length(data.netbox_sites.all.sites) > 0 ? data.netbox_sites.all.sites[0].name : "No sites found"
}

# Output the total number of sites found.
output "site_count" {
  value = length(data.netbox_sites.all.sites)
}
