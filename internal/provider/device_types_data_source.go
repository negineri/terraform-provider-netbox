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

var _ datasource.DataSource = &deviceTypesDataSource{}
var _ datasource.DataSourceWithConfigure = &deviceTypesDataSource{}

func NewDeviceTypesDataSource() datasource.DataSource {
	return &deviceTypesDataSource{}
}

type deviceTypesDataSource struct {
	client *client.NetboxClient
}

type deviceTypesDataSourceModel struct {
	Id                 types.String      `tfsdk:"id"`
	DeviceTypes        []deviceTypeModel `tfsdk:"device_types"`
	CustomFieldFilters types.Map         `tfsdk:"custom_field_filters"`
}

type deviceTypeModel struct {
	Id             types.Int64   `tfsdk:"id"`
	ManufacturerId types.Int64   `tfsdk:"manufacturer_id"`
	Model          types.String  `tfsdk:"model"`
	Slug           types.String  `tfsdk:"slug"`
	PartNumber     types.String  `tfsdk:"part_number"`
	UHeight        types.Float64 `tfsdk:"u_height"`
	IsFullDepth    types.Bool    `tfsdk:"is_full_depth"`
	Description    types.String  `tfsdk:"description"`
}

func (d *deviceTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_types"
}

func (d *deviceTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of device types from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"custom_field_filters": schema.MapAttribute{
				MarkdownDescription: "Filter device types by custom field values. Keys are custom field names, values are the filter values.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"device_types": schema.ListNestedAttribute{
				MarkdownDescription: "List of device types.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the device type.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *deviceTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *deviceTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state deviceTypesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.DeviceTypes = []deviceTypeModel{}

	apiPath := "api/dcim/device-types/"
	if query := buildCustomFieldFilterQuery(ctx, state.CustomFieldFilters); query != "" {
		apiPath += "?" + query
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch device types, got error: %s", err))
		return
	}

	type ApiManufacturer struct {
		ID int64 `json:"id"`
	}

	type ApiDeviceType struct {
		ID           int64           `json:"id"`
		Manufacturer ApiManufacturer `json:"manufacturer"`
		Model        string          `json:"model"`
		Slug         string          `json:"slug"`
		PartNumber   string          `json:"part_number"`
		UHeight      float64         `json:"u_height"`
		IsFullDepth  bool            `json:"is_full_depth"`
		Description  string          `json:"description"`
	}

	type ApiDeviceTypesResponse struct {
		Count   int             `json:"count"`
		Results []ApiDeviceType `json:"results"`
	}

	var response ApiDeviceTypesResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		dt := deviceTypeModel{
			Id:             types.Int64Value(result.ID),
			ManufacturerId: types.Int64Value(result.Manufacturer.ID),
			Model:          types.StringValue(result.Model),
			Slug:           types.StringValue(result.Slug),
			PartNumber:     types.StringValue(result.PartNumber),
			UHeight:        types.Float64Value(result.UHeight),
			IsFullDepth:    types.BoolValue(result.IsFullDepth),
			Description:    types.StringValue(result.Description),
		}
		state.DeviceTypes = append(state.DeviceTypes, dt)
	}

	state.Id = types.StringValue("netbox_device_types")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
