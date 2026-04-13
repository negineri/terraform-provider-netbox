# List all tags.
data "netbox_tags" "all" {}

# Output the name of the first tag, or a message if no tags are found.
output "first_tag_name" {
  value = length(data.netbox_tags.all.tags) > 0 ? data.netbox_tags.all.tags[0].name : "No tags found"
}

# Output the total number of tags found.
output "tag_count" {
  value = length(data.netbox_tags.all.tags)
}
