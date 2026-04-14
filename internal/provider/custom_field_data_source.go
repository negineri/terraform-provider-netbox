// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &customFieldDataSource{}
var _ datasource.DataSourceWithConfigure = &customFieldDataSource{}

func NewCustomFieldDataSource() datasource.DataSource {
	return &customFieldDataSource{}
}

type customFieldDataSource struct {
	client *client.NetboxClient
}

type customFieldDataSourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Label        types.String `tfsdk:"label"`
	Type         types.String `tfsdk:"type"`
	ContentTypes types.List   `tfsdk:"content_types"`
	Required     types.Bool   `tfsdk:"required"`
	Description  types.String `tfsdk:"description"`
	Default      types.String `tfsdk:"default"`
	Weight       types.Int64  `tfsdk:"weight"`
	FilterLogic  types.String `tfsdk:"filter_logic"`
	GroupName    types.String `tfsdk:"group_name"`
	Choices      types.List   `tfsdk:"choices"`
	UIVisible    types.String `tfsdk:"ui_visible"`
	UIEditable   types.String `tfsdk:"ui_editable"`
	IsCloneable  types.Bool   `tfsdk:"is_cloneable"`
	SearchWeight types.Int64  `tfsdk:"search_weight"`
}

func (d *customFieldDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_field"
}

func (d *customFieldDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single Custom Field from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the custom field.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Internal field name.",
				Computed:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Display name for the field in the UI.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the custom field.",
				Computed:            true,
			},
			"content_types": schema.ListAttribute{
				MarkdownDescription: "List of content types this custom field is assigned to.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"required": schema.BoolAttribute{
				MarkdownDescription: "Whether a value is required for this field.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the custom field.",
				Computed:            true,
			},
			"default": schema.StringAttribute{
				MarkdownDescription: "Default value for the custom field.",
				Computed:            true,
			},
			"weight": schema.Int64Attribute{
				MarkdownDescription: "Fields with higher weights appear lower in the form.",
				Computed:            true,
			},
			"filter_logic": schema.StringAttribute{
				MarkdownDescription: "Filtering logic for this field.",
				Computed:            true,
			},
			"group_name": schema.StringAttribute{
				MarkdownDescription: "Custom fields within the same group will be displayed together.",
				Computed:            true,
			},
			"choices": schema.ListAttribute{
				MarkdownDescription: "Available choices for select/multiselect fields.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"ui_visible": schema.StringAttribute{
				MarkdownDescription: "Controls visibility in the UI.",
				Computed:            true,
			},
			"ui_editable": schema.StringAttribute{
				MarkdownDescription: "Controls editability in the UI.",
				Computed:            true,
			},
			"is_cloneable": schema.BoolAttribute{
				MarkdownDescription: "Whether this field is copied when cloning an object.",
				Computed:            true,
			},
			"search_weight": schema.Int64Attribute{
				MarkdownDescription: "Weighting for search.",
				Computed:            true,
			},
		},
	}
}

func (d *customFieldDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.NetboxClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.NetboxClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *customFieldDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state customFieldDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/extras/custom-fields/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch custom field, got error: %s", err))
		return
	}

	var apiResponse map[string]any
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	if name, ok := apiResponse["name"].(string); ok {
		state.Name = types.StringValue(name)
	}
	if label, ok := apiResponse["label"].(string); ok {
		state.Label = types.StringValue(label)
	}
	if t, ok := apiResponse["type"].(map[string]interface{}); ok {
		if v, ok := t["value"].(string); ok {
			state.Type = types.StringValue(v)
		}
	}
	ctRaw, _ := apiResponse["object_types"].([]interface{})
	{
		elems := make([]attr.Value, 0, len(ctRaw))
		for _, v := range ctRaw {
			if s, ok := v.(string); ok {
				elems = append(elems, types.StringValue(s))
			}
		}
		state.ContentTypes = types.ListValueMust(types.StringType, elems)
	}
	if req, ok := apiResponse["required"].(bool); ok {
		state.Required = types.BoolValue(req)
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}
	if def, ok := apiResponse["default"]; ok && def != nil {
		switch v := def.(type) {
		case string:
			state.Default = types.StringValue(v)
		case float64:
			state.Default = types.StringValue(fmt.Sprintf("%v", v))
		case bool:
			if v {
				state.Default = types.StringValue("true")
			} else {
				state.Default = types.StringValue("false")
			}
		default:
			state.Default = types.StringNull()
		}
	} else {
		state.Default = types.StringNull()
	}
	if weight, ok := apiResponse["weight"].(float64); ok {
		state.Weight = types.Int64Value(int64(weight))
	}
	if fl, ok := apiResponse["filter_logic"].(map[string]interface{}); ok {
		if v, ok := fl["value"].(string); ok {
			state.FilterLogic = types.StringValue(v)
		}
	}
	if gn, ok := apiResponse["group_name"].(string); ok {
		state.GroupName = types.StringValue(gn)
	}
	if choicesRaw, ok := apiResponse["choices"].([]interface{}); ok {
		elems := make([]attr.Value, 0, len(choicesRaw))
		for _, v := range choicesRaw {
			if s, ok := v.(string); ok {
				elems = append(elems, types.StringValue(s))
			}
		}
		state.Choices = types.ListValueMust(types.StringType, elems)
	} else {
		state.Choices = types.ListValueMust(types.StringType, []attr.Value{})
	}
	if uiv, ok := apiResponse["ui_visible"].(map[string]interface{}); ok {
		if v, ok := uiv["value"].(string); ok {
			state.UIVisible = types.StringValue(v)
		}
	}
	if uie, ok := apiResponse["ui_editable"].(map[string]interface{}); ok {
		if v, ok := uie["value"].(string); ok {
			state.UIEditable = types.StringValue(v)
		}
	}
	if ic, ok := apiResponse["is_cloneable"].(bool); ok {
		state.IsCloneable = types.BoolValue(ic)
	}
	if sw, ok := apiResponse["search_weight"].(float64); ok {
		state.SearchWeight = types.Int64Value(int64(sw))
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
