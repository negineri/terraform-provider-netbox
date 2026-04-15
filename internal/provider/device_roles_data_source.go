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

var _ datasource.DataSource = &deviceRolesDataSource{}
var _ datasource.DataSourceWithConfigure = &deviceRolesDataSource{}

func NewDeviceRolesDataSource() datasource.DataSource {
	return &deviceRolesDataSource{}
}

type deviceRolesDataSource struct {
	client *client.NetboxClient
}

type deviceRolesDataSourceModel struct {
	Id                 types.String      `tfsdk:"id"`
	DeviceRoles        []deviceRoleModel `tfsdk:"device_roles"`
	CustomFieldFilters types.Map         `tfsdk:"custom_field_filters"`
}

type deviceRoleModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Color       types.String `tfsdk:"color"`
	VmRole      types.Bool   `tfsdk:"vm_role"`
	Description types.String `tfsdk:"description"`
}

func (d *deviceRolesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_roles"
}

func (d *deviceRolesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of device roles from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"custom_field_filters": schema.MapAttribute{
				MarkdownDescription: "Filter device roles by custom field values. Keys are custom field names, values are the filter values.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"device_roles": schema.ListNestedAttribute{
				MarkdownDescription: "List of device roles.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the device role.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *deviceRolesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *deviceRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceRolesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.DeviceRoles = []deviceRoleModel{}

	apiPath := "api/dcim/device-roles/"
	if query := buildCustomFieldFilterQuery(ctx, state.CustomFieldFilters); query != "" {
		apiPath += "?" + query
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch device roles, got error: %s", err))
		return
	}

	type ApiDeviceRole struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Color       string `json:"color"`
		VmRole      bool   `json:"vm_role"`
		Description string `json:"description"`
	}

	type ApiDeviceRolesResponse struct {
		Count   int             `json:"count"`
		Results []ApiDeviceRole `json:"results"`
	}

	var response ApiDeviceRolesResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		role := deviceRoleModel{
			Id:          types.Int64Value(result.ID),
			Name:        types.StringValue(result.Name),
			Slug:        types.StringValue(result.Slug),
			Color:       types.StringValue(result.Color),
			VmRole:      types.BoolValue(result.VmRole),
			Description: types.StringValue(result.Description),
		}
		state.DeviceRoles = append(state.DeviceRoles, role)
	}

	state.Id = types.StringValue("netbox_device_roles")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
