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

var _ datasource.DataSource = &macAddressDataSource{}
var _ datasource.DataSourceWithConfigure = &macAddressDataSource{}

func NewMacAddressDataSource() datasource.DataSource {
	return &macAddressDataSource{}
}

type macAddressDataSource struct {
	client *client.NetboxClient
}

type macAddressDataSourceModel struct {
	Id                 types.Int64  `tfsdk:"id"`
	MacAddress         types.String `tfsdk:"mac_address"`
	AssignedObjectType types.String `tfsdk:"assigned_object_type"`
	AssignedObjectId   types.Int64  `tfsdk:"assigned_object_id"`
	Description        types.String `tfsdk:"description"`
	Comments           types.String `tfsdk:"comments"`
}

func (d *macAddressDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mac_address"
}

func (d *macAddressDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single MAC address from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the MAC address.",
				Required:            true,
			},
			"mac_address": schema.StringAttribute{
				MarkdownDescription: "The MAC address.",
				Computed:            true,
			},
			"assigned_object_type": schema.StringAttribute{
				MarkdownDescription: "The type of the assigned object (e.g., dcim.interface, virtualization.vminterface).",
				Computed:            true,
			},
			"assigned_object_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the assigned object.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the MAC address.",
				Computed:            true,
			},
			"comments": schema.StringAttribute{
				MarkdownDescription: "Comments for the MAC address.",
				Computed:            true,
			},
		},
	}
}

func (d *macAddressDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *macAddressDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state macAddressDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/mac-addresses/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch MAC address, got error: %s", err))
		return
	}

	var apiResponse map[string]any
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	if macAddr, ok := apiResponse["mac_address"].(string); ok {
		state.MacAddress = types.StringValue(macAddr)
	}
	if objType, ok := apiResponse["assigned_object_type"].(string); ok {
		state.AssignedObjectType = types.StringValue(objType)
	} else {
		state.AssignedObjectType = types.StringValue("")
	}
	if objIdFloat, ok := apiResponse["assigned_object_id"].(float64); ok {
		state.AssignedObjectId = types.Int64Value(int64(objIdFloat))
	} else {
		state.AssignedObjectId = types.Int64Null()
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}
	if comments, ok := apiResponse["comments"].(string); ok {
		state.Comments = types.StringValue(comments)
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
