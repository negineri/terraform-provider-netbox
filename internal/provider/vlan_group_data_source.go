// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &vlanGroupDataSource{}
var _ datasource.DataSourceWithConfigure = &vlanGroupDataSource{}

func NewVlanGroupDataSource() datasource.DataSource {
	return &vlanGroupDataSource{}
}

type vlanGroupDataSource struct {
	client *client.NetboxClient
}

type vlanGroupDataSourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Description types.String `tfsdk:"description"`
	MinVid      types.Int64  `tfsdk:"min_vid"`
	MaxVid      types.Int64  `tfsdk:"max_vid"`
	ScopeType   types.String `tfsdk:"scope_type"`
	ScopeId     types.Int64  `tfsdk:"scope_id"`
}

func (d *vlanGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_group"
}

func (d *vlanGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single VLAN group from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the VLAN group.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the VLAN group.",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-friendly unique shorthand for the VLAN group.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the VLAN group.",
				Computed:            true,
			},
			"min_vid": schema.Int64Attribute{
				MarkdownDescription: "Minimum VLAN ID.",
				Computed:            true,
			},
			"max_vid": schema.Int64Attribute{
				MarkdownDescription: "Maximum VLAN ID.",
				Computed:            true,
			},
			"scope_type": schema.StringAttribute{
				MarkdownDescription: "The type of the scope object.",
				Computed:            true,
			},
			"scope_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the scope object.",
				Computed:            true,
			},
		},
	}
}

func (d *vlanGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vlanGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vlanGroupDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/ipam/vlan-groups/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch VLAN group, got error: %s", err))
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
	if slug, ok := apiResponse["slug"].(string); ok {
		state.Slug = types.StringValue(slug)
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}
	if minVid, ok := apiResponse["min_vid"].(float64); ok {
		state.MinVid = types.Int64Value(int64(minVid))
	} else {
		state.MinVid = types.Int64Null()
	}
	if maxVid, ok := apiResponse["max_vid"].(float64); ok {
		state.MaxVid = types.Int64Value(int64(maxVid))
	} else {
		state.MaxVid = types.Int64Null()
	}
	if scopeType, ok := apiResponse["scope_type"].(string); ok {
		state.ScopeType = types.StringValue(scopeType)
	} else {
		state.ScopeType = types.StringValue("")
	}
	if scopeId, ok := apiResponse["scope_id"].(float64); ok {
		state.ScopeId = types.Int64Value(int64(scopeId))
	} else {
		state.ScopeId = types.Int64Null()
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
