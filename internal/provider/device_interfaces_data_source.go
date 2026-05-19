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

var _ datasource.DataSource = &deviceInterfacesDataSource{}
var _ datasource.DataSourceWithConfigure = &deviceInterfacesDataSource{}

func NewDeviceInterfacesDataSource() datasource.DataSource {
	return &deviceInterfacesDataSource{}
}

type deviceInterfacesDataSource struct {
	client *client.NetboxClient
}

type deviceInterfacesDataSourceModel struct {
	Id               types.String           `tfsdk:"id"`
	DeviceInterfaces []deviceInterfaceModel `tfsdk:"device_interfaces"`
	DeviceId         types.Int64            `tfsdk:"device_id"`
	Name             types.String           `tfsdk:"name"`
	Type             types.String           `tfsdk:"type"`
	Enabled          types.Bool             `tfsdk:"enabled"`
}

type deviceInterfaceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	DeviceId    types.Int64  `tfsdk:"device_id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	MacAddress  types.String `tfsdk:"mac_address"`
	Mtu         types.Int64  `tfsdk:"mtu"`
	Description types.String `tfsdk:"description"`
}

func (d *deviceInterfacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_interfaces"
}

func (d *deviceInterfacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of Device Interfaces from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"device_id": schema.Int64Attribute{
				MarkdownDescription: "Filter by device ID.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Filter by interface name.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Filter by interface type.",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Filter by enabled state.",
				Optional:            true,
			},
			"device_interfaces": schema.ListNestedAttribute{
				MarkdownDescription: "List of device interfaces.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the interface.",
							Computed:            true,
						},
						"device_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the device.",
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
				},
			},
		},
	}
}

func (d *deviceInterfacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *deviceInterfacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceInterfacesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.DeviceInterfaces = []deviceInterfaceModel{}

	params := map[string]string{}
	int64FilterParam(params, "device_id", state.DeviceId)
	stringFilterParam(params, "name", state.Name)
	stringFilterParam(params, "type", state.Type)
	if !state.Enabled.IsNull() && !state.Enabled.IsUnknown() {
		if state.Enabled.ValueBool() {
			params["enabled"] = "true"
		} else {
			params["enabled"] = "false"
		}
	}

	apiPath := "api/dcim/interfaces/"
	if q := buildFilterQuery(params); q != "" {
		apiPath += "?" + q
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch device interfaces, got error: %s", err))
		return
	}

	type ApiDeviceInterface struct {
		ID          int64                  `json:"id"`
		Device      map[string]interface{} `json:"device"`
		Name        string                 `json:"name"`
		Type        map[string]interface{} `json:"type"`
		Enabled     bool                   `json:"enabled"`
		MacAddress  *string                `json:"mac_address"`
		Mtu         *float64               `json:"mtu"`
		Description string                 `json:"description"`
	}

	type ApiResponse struct {
		Count   int                  `json:"count"`
		Results []ApiDeviceInterface `json:"results"`
	}

	var response ApiResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		m := deviceInterfaceModel{
			Id:          types.Int64Value(result.ID),
			Name:        types.StringValue(result.Name),
			Enabled:     types.BoolValue(result.Enabled),
			Description: types.StringValue(result.Description),
		}
		if idFloat, ok := result.Device["id"].(float64); ok {
			m.DeviceId = types.Int64Value(int64(idFloat))
		}
		if val, ok := result.Type["value"].(string); ok {
			m.Type = types.StringValue(val)
		} else {
			m.Type = types.StringValue("")
		}
		if result.MacAddress != nil {
			m.MacAddress = types.StringValue(*result.MacAddress)
		} else {
			m.MacAddress = types.StringValue("")
		}
		if result.Mtu != nil {
			m.Mtu = types.Int64Value(int64(*result.Mtu))
		} else {
			m.Mtu = types.Int64Null()
		}
		state.DeviceInterfaces = append(state.DeviceInterfaces, m)
	}

	state.Id = types.StringValue("netbox_device_interfaces")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
