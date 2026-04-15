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

var _ datasource.DataSource = &deviceTypeDataSource{}
var _ datasource.DataSourceWithConfigure = &deviceTypeDataSource{}

func NewDeviceTypeDataSource() datasource.DataSource {
	return &deviceTypeDataSource{}
}

type deviceTypeDataSource struct {
	client *client.NetboxClient
}

type deviceTypeDataSourceModel struct {
	Id             types.Int64   `tfsdk:"id"`
	ManufacturerId types.Int64   `tfsdk:"manufacturer_id"`
	Model          types.String  `tfsdk:"model"`
	Slug           types.String  `tfsdk:"slug"`
	PartNumber     types.String  `tfsdk:"part_number"`
	UHeight        types.Float64 `tfsdk:"u_height"`
	IsFullDepth    types.Bool    `tfsdk:"is_full_depth"`
	Description    types.String  `tfsdk:"description"`
}

func (d *deviceTypeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_type"
}

func (d *deviceTypeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single device type from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the device type.",
				Required:            true,
			},
			"manufacturer_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the manufacturer.",
				Computed:            true,
			},
			"model": schema.StringAttribute{
				MarkdownDescription: "The model name of the device type.",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-friendly unique shorthand for the device type.",
				Computed:            true,
			},
			"part_number": schema.StringAttribute{
				MarkdownDescription: "Discrete part number for the device type.",
				Computed:            true,
			},
			"u_height": schema.Float64Attribute{
				MarkdownDescription: "Device height in rack units.",
				Computed:            true,
			},
			"is_full_depth": schema.BoolAttribute{
				MarkdownDescription: "Whether the device type occupies the full depth of a rack.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the device type.",
				Computed:            true,
			},
		},
	}
}

func (d *deviceTypeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *deviceTypeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceTypeDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/device-types/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch device type, got error: %s", err))
		return
	}

	var apiResponse map[string]any
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	if manufacturerMap, ok := apiResponse["manufacturer"].(map[string]any); ok {
		if idFloat, ok := manufacturerMap["id"].(float64); ok {
			state.ManufacturerId = types.Int64Value(int64(idFloat))
		}
	}
	if model, ok := apiResponse["model"].(string); ok {
		state.Model = types.StringValue(model)
	}
	if slugVal, ok := apiResponse["slug"].(string); ok {
		state.Slug = types.StringValue(slugVal)
	}
	if partNumber, ok := apiResponse["part_number"].(string); ok {
		state.PartNumber = types.StringValue(partNumber)
	} else {
		state.PartNumber = types.StringValue("")
	}
	if uHeight, ok := apiResponse["u_height"].(float64); ok {
		state.UHeight = types.Float64Value(uHeight)
	}
	if isFullDepth, ok := apiResponse["is_full_depth"].(bool); ok {
		state.IsFullDepth = types.BoolValue(isFullDepth)
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
