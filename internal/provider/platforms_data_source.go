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

var _ datasource.DataSource = &platformsDataSource{}
var _ datasource.DataSourceWithConfigure = &platformsDataSource{}

func NewPlatformsDataSource() datasource.DataSource {
	return &platformsDataSource{}
}

type platformsDataSource struct {
	client *client.NetboxClient
}

type platformsDataSourceModel struct {
	Id                 types.String    `tfsdk:"id"`
	Platforms          []platformModel `tfsdk:"platforms"`
	CustomFieldFilters types.Map       `tfsdk:"custom_field_filters"`
	Name               types.String    `tfsdk:"name"`
	Slug               types.String    `tfsdk:"slug"`
	ManufacturerId     types.Int64     `tfsdk:"manufacturer_id"`
}

type platformModel struct {
	Id             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Slug           types.String `tfsdk:"slug"`
	Description    types.String `tfsdk:"description"`
	ManufacturerId types.Int64  `tfsdk:"manufacturer_id"`
}

func (d *platformsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_platforms"
}

func (d *platformsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of platforms from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"custom_field_filters": schema.MapAttribute{
				MarkdownDescription: "Filter platforms by custom field values. Keys are custom field names, values are the filter values.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Filter by platform name.",
				Optional:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Filter by platform slug.",
				Optional:            true,
			},
			"manufacturer_id": schema.Int64Attribute{
				MarkdownDescription: "Filter by manufacturer ID.",
				Optional:            true,
			},
			"platforms": schema.ListNestedAttribute{
				MarkdownDescription: "List of platforms.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the platform.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the platform.",
							Computed:            true,
						},
						"slug": schema.StringAttribute{
							MarkdownDescription: "URL-friendly unique shorthand for the platform.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the platform.",
							Computed:            true,
						},
						"manufacturer_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the manufacturer associated with this platform.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *platformsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *platformsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state platformsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Platforms = []platformModel{}

	params := map[string]string{}
	stringFilterParam(params, "name", state.Name)
	stringFilterParam(params, "slug", state.Slug)
	if !state.ManufacturerId.IsNull() && !state.ManufacturerId.IsUnknown() {
		params["manufacturer_id"] = fmt.Sprintf("%d", state.ManufacturerId.ValueInt64())
	}

	apiPath := "api/dcim/platforms/"
	if q := combineQueryStrings(buildFilterQuery(params), buildCustomFieldFilterQuery(ctx, state.CustomFieldFilters)); q != "" {
		apiPath += "?" + q
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch platforms, got error: %s", err))
		return
	}

	type ApiPlatformManufacturer struct {
		ID int64 `json:"id"`
	}

	type ApiPlatform struct {
		ID           int64                    `json:"id"`
		Name         string                   `json:"name"`
		Slug         string                   `json:"slug"`
		Description  string                   `json:"description"`
		Manufacturer *ApiPlatformManufacturer `json:"manufacturer"`
	}

	type ApiPlatformsResponse struct {
		Count   int           `json:"count"`
		Results []ApiPlatform `json:"results"`
	}

	var response ApiPlatformsResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		platform := platformModel{
			Id:             types.Int64Value(result.ID),
			Name:           types.StringValue(result.Name),
			Slug:           types.StringValue(result.Slug),
			Description:    types.StringValue(result.Description),
			ManufacturerId: types.Int64Null(),
		}
		if result.Manufacturer != nil {
			platform.ManufacturerId = types.Int64Value(result.Manufacturer.ID)
		}
		state.Platforms = append(state.Platforms, platform)
	}

	state.Id = types.StringValue("netbox_platforms")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
