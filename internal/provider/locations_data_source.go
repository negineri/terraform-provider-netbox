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

var _ datasource.DataSource = &locationsDataSource{}
var _ datasource.DataSourceWithConfigure = &locationsDataSource{}

func NewLocationsDataSource() datasource.DataSource {
	return &locationsDataSource{}
}

type locationsDataSource struct {
	client *client.NetboxClient
}

type locationsDataSourceModel struct {
	Id                 types.String    `tfsdk:"id"`
	Locations          []locationModel `tfsdk:"locations"`
	CustomFieldFilters types.Map       `tfsdk:"custom_field_filters"`
	Name               types.String    `tfsdk:"name"`
	SiteId             types.Int64     `tfsdk:"site_id"`
	Status             types.String    `tfsdk:"status"`
}

type locationModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	SiteId      types.Int64  `tfsdk:"site_id"`
	ParentId    types.Int64  `tfsdk:"parent_id"`
	Status      types.String `tfsdk:"status"`
	Description types.String `tfsdk:"description"`
}

func (d *locationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_locations"
}

func (d *locationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of locations from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"custom_field_filters": schema.MapAttribute{
				MarkdownDescription: "Filter locations by custom field values. Keys are custom field names, values are the filter values.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Filter by location name.",
				Optional:            true,
			},
			"site_id": schema.Int64Attribute{
				MarkdownDescription: "Filter by site ID.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Filter by status (e.g. active, planned, staging, decommissioning, retired).",
				Optional:            true,
			},
			"locations": schema.ListNestedAttribute{
				MarkdownDescription: "List of locations.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the location.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the location.",
							Computed:            true,
						},
						"slug": schema.StringAttribute{
							MarkdownDescription: "URL-friendly unique shorthand for the location.",
							Computed:            true,
						},
						"site_id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the site this location belongs to.",
							Computed:            true,
						},
						"parent_id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the parent location.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the location.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the location.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *locationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *locationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state locationsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Locations = []locationModel{}

	params := map[string]string{}
	stringFilterParam(params, "name", state.Name)
	int64FilterParam(params, "site_id", state.SiteId)
	stringFilterParam(params, "status", state.Status)

	apiPath := "api/dcim/locations/"
	if q := combineQueryStrings(buildFilterQuery(params), buildCustomFieldFilterQuery(ctx, state.CustomFieldFilters)); q != "" {
		apiPath += "?" + q
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch locations, got error: %s", err))
		return
	}

	type ApiLocation struct {
		ID          int64          `json:"id"`
		Name        string         `json:"name"`
		Slug        string         `json:"slug"`
		Site        map[string]any `json:"site"`
		Parent      map[string]any `json:"parent"`
		Status      map[string]any `json:"status"`
		Description string         `json:"description"`
	}

	type ApiLocationsResponse struct {
		Count   int           `json:"count"`
		Results []ApiLocation `json:"results"`
	}

	var response ApiLocationsResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		loc := locationModel{
			Id:          types.Int64Value(result.ID),
			Name:        types.StringValue(result.Name),
			Slug:        types.StringValue(result.Slug),
			Description: types.StringValue(result.Description),
		}
		if siteID, ok := result.Site["id"].(float64); ok {
			loc.SiteId = types.Int64Value(int64(siteID))
		}
		if result.Parent != nil {
			if parentID, ok := result.Parent["id"].(float64); ok {
				loc.ParentId = types.Int64Value(int64(parentID))
			} else {
				loc.ParentId = types.Int64Null()
			}
		} else {
			loc.ParentId = types.Int64Null()
		}
		if val, ok := result.Status["value"].(string); ok {
			loc.Status = types.StringValue(val)
		} else {
			loc.Status = types.StringValue("")
		}
		state.Locations = append(state.Locations, loc)
	}

	state.Id = types.StringValue("netbox_locations")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
