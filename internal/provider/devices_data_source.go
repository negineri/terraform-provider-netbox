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

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &devicesDataSource{}
var _ datasource.DataSourceWithConfigure = &devicesDataSource{}

func NewDevicesDataSource() datasource.DataSource {
	return &devicesDataSource{}
}

type devicesDataSource struct {
	client *client.NetboxClient
}

type devicesDataSourceModel struct {
	Id                 types.String  `tfsdk:"id"`
	Devices            []deviceModel `tfsdk:"devices"`
	CustomFieldFilters types.Map     `tfsdk:"custom_field_filters"`
	Name               types.String  `tfsdk:"name"`
	Status             types.String  `tfsdk:"status"`
	SiteId             types.Int64   `tfsdk:"site_id"`
	RoleId             types.Int64   `tfsdk:"role_id"`
	Tag                types.String  `tfsdk:"tag"`
}

type deviceModel struct {
	Id            types.Int64  `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	AssetTag      types.String `tfsdk:"asset_tag"`
	PrimaryIPv4Id types.Int64  `tfsdk:"primary_ipv4_id"`
	PrimaryIPv4   types.String `tfsdk:"primary_ipv4"`
	PrimaryIPv6Id types.Int64  `tfsdk:"primary_ipv6_id"`
	PrimaryIPv6   types.String `tfsdk:"primary_ipv6"`
}

func (d *devicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_devices"
}

func (d *devicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of devices from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"custom_field_filters": schema.MapAttribute{
				MarkdownDescription: "Filter devices by custom field values. Keys are custom field names, values are the filter values.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Filter by device name.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Filter by status (e.g. active, planned, staged, failed, decommissioning, inventory, offline).",
				Optional:            true,
			},
			"site_id": schema.Int64Attribute{
				MarkdownDescription: "Filter by site ID.",
				Optional:            true,
			},
			"role_id": schema.Int64Attribute{
				MarkdownDescription: "Filter by device role ID.",
				Optional:            true,
			},
			"tag": schema.StringAttribute{
				MarkdownDescription: "Filter by tag slug.",
				Optional:            true,
			},
			"devices": schema.ListNestedAttribute{
				MarkdownDescription: "List of devices.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the device.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the device.",
							Computed:            true,
						},
						"asset_tag": schema.StringAttribute{
							MarkdownDescription: "A unique tag used to identify the device.",
							Computed:            true,
						},
						"primary_ipv4_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the primary IPv4 address.",
							Computed:            true,
						},
						"primary_ipv4": schema.StringAttribute{
							MarkdownDescription: "The primary IPv4 address in CIDR notation.",
							Computed:            true,
						},
						"primary_ipv6_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the primary IPv6 address.",
							Computed:            true,
						},
						"primary_ipv6": schema.StringAttribute{
							MarkdownDescription: "The primary IPv6 address in CIDR notation.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *devicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *devicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state devicesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Devices = []deviceModel{}

	params := map[string]string{}
	stringFilterParam(params, "name", state.Name)
	stringFilterParam(params, "status", state.Status)
	int64FilterParam(params, "site_id", state.SiteId)
	int64FilterParam(params, "role_id", state.RoleId)
	stringFilterParam(params, "tag", state.Tag)

	apiPath := "api/dcim/devices/"
	if q := combineQueryStrings(buildFilterQuery(params), buildCustomFieldFilterQuery(ctx, state.CustomFieldFilters)); q != "" {
		apiPath += "?" + q
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch devices, got error: %s", err))
		return
	}

	type ApiIPRef struct {
		ID      int64  `json:"id"`
		Address string `json:"address"`
	}

	type ApiDevice struct {
		ID          int64     `json:"id"`
		Name        string    `json:"name"`
		AssetTag    *string   `json:"asset_tag"`
		PrimaryIPv4 *ApiIPRef `json:"primary_ip4"`
		PrimaryIPv6 *ApiIPRef `json:"primary_ip6"`
	}

	type ApiDevicesResponse struct {
		Count   int         `json:"count"`
		Results []ApiDevice `json:"results"`
	}

	var response ApiDevicesResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		deviceState := deviceModel{
			Id:   types.Int64Value(result.ID),
			Name: types.StringValue(result.Name),
		}
		if result.AssetTag != nil {
			deviceState.AssetTag = types.StringValue(*result.AssetTag)
		} else {
			deviceState.AssetTag = types.StringNull()
		}
		if result.PrimaryIPv4 != nil {
			deviceState.PrimaryIPv4Id = types.Int64Value(result.PrimaryIPv4.ID)
			deviceState.PrimaryIPv4 = types.StringValue(result.PrimaryIPv4.Address)
		} else {
			deviceState.PrimaryIPv4Id = types.Int64Null()
			deviceState.PrimaryIPv4 = types.StringNull()
		}
		if result.PrimaryIPv6 != nil {
			deviceState.PrimaryIPv6Id = types.Int64Value(result.PrimaryIPv6.ID)
			deviceState.PrimaryIPv6 = types.StringValue(result.PrimaryIPv6.Address)
		} else {
			deviceState.PrimaryIPv6Id = types.Int64Null()
			deviceState.PrimaryIPv6 = types.StringNull()
		}
		state.Devices = append(state.Devices, deviceState)
	}

	state.Id = types.StringValue("netbox_devices")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
