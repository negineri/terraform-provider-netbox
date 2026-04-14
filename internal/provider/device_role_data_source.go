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

var _ datasource.DataSource = &deviceRoleDataSource{}
var _ datasource.DataSourceWithConfigure = &deviceRoleDataSource{}

func NewDeviceRoleDataSource() datasource.DataSource {
	return &deviceRoleDataSource{}
}

type deviceRoleDataSource struct {
	client *client.NetboxClient
}

type deviceRoleDataSourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Color       types.String `tfsdk:"color"`
	VmRole      types.Bool   `tfsdk:"vm_role"`
	Description types.String `tfsdk:"description"`
}

func (d *deviceRoleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_role"
}

func (d *deviceRoleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single device role from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the device role.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the device role.",
				Computed:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-friendly unique shorthand for the device role.",
				Computed:            true,
			},
			"color": schema.StringAttribute{
				MarkdownDescription: "Color for the device role as a 6-digit hex string.",
				Computed:            true,
			},
			"vm_role": schema.BoolAttribute{
				MarkdownDescription: "Whether this role is used for virtual machines.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the device role.",
				Computed:            true,
			},
		},
	}
}

func (d *deviceRoleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *deviceRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceRoleDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/device-roles/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch device role, got error: %s", err))
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
	if color, ok := apiResponse["color"].(string); ok {
		state.Color = types.StringValue(color)
	}
	if vmRole, ok := apiResponse["vm_role"].(bool); ok {
		state.VmRole = types.BoolValue(vmRole)
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
