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

var _ datasource.DataSource = &vlanDataSource{}
var _ datasource.DataSourceWithConfigure = &vlanDataSource{}

func NewVlanDataSource() datasource.DataSource {
	return &vlanDataSource{}
}

type vlanDataSource struct {
	client *client.NetboxClient
}

type vlanDataSourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Vid         types.Int64  `tfsdk:"vid"`
	Status      types.String `tfsdk:"status"`
	Description types.String `tfsdk:"description"`
	GroupId     types.Int64  `tfsdk:"group_id"`
}

func (d *vlanDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan"
}

func (d *vlanDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single VLAN from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the VLAN.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the VLAN.",
				Computed:            true,
			},
			"vid": schema.Int64Attribute{
				MarkdownDescription: "VLAN ID (1-4094).",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the VLAN.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the VLAN.",
				Computed:            true,
			},
			"group_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the VLAN group this VLAN belongs to.",
				Computed:            true,
			},
		},
	}
}

func (d *vlanDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vlanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vlanDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/ipam/vlans/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch VLAN, got error: %s", err))
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
	if vid, ok := apiResponse["vid"].(float64); ok {
		state.Vid = types.Int64Value(int64(vid))
	}
	if statusMap, ok := apiResponse["status"].(map[string]any); ok {
		if val, ok := statusMap["value"].(string); ok {
			state.Status = types.StringValue(val)
		}
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}
	if groupMap, ok := apiResponse["group"].(map[string]any); ok {
		if idFloat, ok := groupMap["id"].(float64); ok {
			state.GroupId = types.Int64Value(int64(idFloat))
		} else {
			state.GroupId = types.Int64Null()
		}
	} else {
		state.GroupId = types.Int64Null()
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
