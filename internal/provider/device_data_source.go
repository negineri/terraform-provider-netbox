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

var _ datasource.DataSource = &deviceDataSource{}
var _ datasource.DataSourceWithConfigure = &deviceDataSource{}

func NewDeviceDataSource() datasource.DataSource {
	return &deviceDataSource{}
}

type deviceDataSource struct {
	client *client.NetboxClient
}

type deviceDataSourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	DeviceTypeId types.Int64  `tfsdk:"device_type_id"`
	RoleId       types.Int64  `tfsdk:"role_id"`
	SiteId       types.Int64  `tfsdk:"site_id"`
	Status       types.String `tfsdk:"status"`
	Description  types.String `tfsdk:"description"`
	Serial       types.String `tfsdk:"serial"`
}

func (d *deviceDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (d *deviceDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single device from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the device.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the device.",
				Computed:            true,
			},
			"device_type_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the device type.",
				Computed:            true,
			},
			"role_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the device role.",
				Computed:            true,
			},
			"site_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the site where the device is located.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the device.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the device.",
				Computed:            true,
			},
			"serial": schema.StringAttribute{
				MarkdownDescription: "Serial number of the device.",
				Computed:            true,
			},
		},
	}
}

func (d *deviceDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *deviceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/devices/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch device, got error: %s", err))
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
	if deviceTypeMap, ok := apiResponse["device_type"].(map[string]any); ok {
		if idFloat, ok := deviceTypeMap["id"].(float64); ok {
			state.DeviceTypeId = types.Int64Value(int64(idFloat))
		}
	}
	if roleMap, ok := apiResponse["role"].(map[string]any); ok {
		if idFloat, ok := roleMap["id"].(float64); ok {
			state.RoleId = types.Int64Value(int64(idFloat))
		}
	}
	if siteMap, ok := apiResponse["site"].(map[string]any); ok {
		if idFloat, ok := siteMap["id"].(float64); ok {
			state.SiteId = types.Int64Value(int64(idFloat))
		}
	}
	if statusMap, ok := apiResponse["status"].(map[string]any); ok {
		if val, ok := statusMap["value"].(string); ok {
			state.Status = types.StringValue(val)
		}
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}
	if serial, ok := apiResponse["serial"].(string); ok {
		state.Serial = types.StringValue(serial)
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
