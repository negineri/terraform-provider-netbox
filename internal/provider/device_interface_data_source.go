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

var _ datasource.DataSource = &deviceInterfaceDataSource{}
var _ datasource.DataSourceWithConfigure = &deviceInterfaceDataSource{}

func NewDeviceInterfaceDataSource() datasource.DataSource {
	return &deviceInterfaceDataSource{}
}

type deviceInterfaceDataSource struct {
	client *client.NetboxClient
}

type deviceInterfaceDataSourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	DeviceId    types.Int64  `tfsdk:"device_id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	MacAddress  types.String `tfsdk:"mac_address"`
	Mtu         types.Int64  `tfsdk:"mtu"`
	Description types.String `tfsdk:"description"`
}

func (d *deviceInterfaceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_interface"
}

func (d *deviceInterfaceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single Device Interface from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the interface.",
				Required:            true,
			},
			"device_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the device this interface belongs to.",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the interface.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The interface type.",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the interface is enabled.",
				Computed:            true,
			},
			"mac_address": schema.StringAttribute{
				MarkdownDescription: "The MAC address of the interface.",
				Computed:            true,
			},
			"mtu": schema.Int64Attribute{
				MarkdownDescription: "The MTU of the interface.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the interface.",
				Computed:            true,
			},
		},
	}
}

func (d *deviceInterfaceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *deviceInterfaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceInterfaceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/interfaces/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch device interface, got error: %s", err))
		return
	}

	var apiResponse map[string]any
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	if deviceMap, ok := apiResponse["device"].(map[string]any); ok {
		if idFloat, ok := deviceMap["id"].(float64); ok {
			state.DeviceId = types.Int64Value(int64(idFloat))
		}
	}
	if name, ok := apiResponse["name"].(string); ok {
		state.Name = types.StringValue(name)
	}
	if typeMap, ok := apiResponse["type"].(map[string]any); ok {
		if val, ok := typeMap["value"].(string); ok {
			state.Type = types.StringValue(val)
		}
	}
	if enabled, ok := apiResponse["enabled"].(bool); ok {
		state.Enabled = types.BoolValue(enabled)
	}
	if macAddr, ok := apiResponse["mac_address"].(string); ok {
		state.MacAddress = types.StringValue(macAddr)
	} else {
		state.MacAddress = types.StringValue("")
	}
	if mtu, ok := apiResponse["mtu"].(float64); ok {
		state.Mtu = types.Int64Value(int64(mtu))
	} else {
		state.Mtu = types.Int64Null()
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
