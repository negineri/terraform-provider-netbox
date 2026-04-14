# テキスト型のカスタムフィールドを dcim.device に追加する例。
resource "netbox_custom_field" "asset_tag" {
  name          = "asset_tag"
  label         = "Asset Tag"
  type          = "text"
  content_types = ["dcim.device"]
  description   = "Internal asset tag for inventory tracking"
  required      = false
  weight        = 100
  filter_logic  = "loose"
}

# セレクト型のカスタムフィールドを複数のオブジェクトに追加する例。
resource "netbox_custom_field" "environment" {
  name          = "environment"
  label         = "Environment"
  type          = "select"
  content_types = ["dcim.device", "virtualization.virtualmachine"]
  description   = "Deployment environment"
  required      = true
  choices       = ["production", "staging", "development"]
  default       = "development"
  filter_logic  = "exact"
  group_name    = "Deployment"
  ui_visible    = "always"
  ui_editable   = "yes"
  is_cloneable  = true
}

# 整数型のカスタムフィールドの例。
resource "netbox_custom_field" "rack_units" {
  name          = "rack_units"
  label         = "Rack Units"
  type          = "integer"
  content_types = ["dcim.device"]
  description   = "Number of rack units occupied"
  required      = false
  weight        = 200
}
