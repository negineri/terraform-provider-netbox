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

var _ datasource.DataSource = &sitesDataSource{}
var _ datasource.DataSourceWithConfigure = &sitesDataSource{}

func NewSitesDataSource() datasource.DataSource {
	return &sitesDataSource{}
}

type sitesDataSource struct {
	client *client.NetboxClient
}

type sitesDataSourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Sites              []siteModel  `tfsdk:"sites"`
	CustomFieldFilters types.Map    `tfsdk:"custom_field_filters"`
	Name               types.String `tfsdk:"name"`
	Status             types.String `tfsdk:"status"`
	RegionId           types.Int64  `tfsdk:"region_id"`
	Tag                types.String `tfsdk:"tag"`
}

type siteModel struct {
	Id              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Slug            types.String `tfsdk:"slug"`
	Status          types.String `tfsdk:"status"`
	RegionId        types.Int64  `tfsdk:"region_id"`
	Description     types.String `tfsdk:"description"`
	Facility        types.String `tfsdk:"facility"`
	TimeZone        types.String `tfsdk:"time_zone"`
	PhysicalAddress types.String `tfsdk:"physical_address"`
}

func (d *sitesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sites"
}

func (d *sitesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of sites from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"custom_field_filters": schema.MapAttribute{
				MarkdownDescription: "Filter sites by custom field values. Keys are custom field names, values are the filter values.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Filter by site name.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Filter by status (e.g. active, planned, staging, decommissioning, retired).",
				Optional:            true,
			},
			"region_id": schema.Int64Attribute{
				MarkdownDescription: "Filter by region ID.",
				Optional:            true,
			},
			"tag": schema.StringAttribute{
				MarkdownDescription: "Filter by tag slug.",
				Optional:            true,
			},
			"sites": schema.ListNestedAttribute{
				MarkdownDescription: "List of sites.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the site.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the site.",
							Computed:            true,
						},
						"slug": schema.StringAttribute{
							MarkdownDescription: "URL-friendly unique shorthand for the site.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the site.",
							Computed:            true,
						},
						"region_id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the region this site belongs to.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the site.",
							Computed:            true,
						},
						"facility": schema.StringAttribute{
							MarkdownDescription: "Physical location of the site.",
							Computed:            true,
						},
						"time_zone": schema.StringAttribute{
							MarkdownDescription: "Time zone of the site.",
							Computed:            true,
						},
						"physical_address": schema.StringAttribute{
							MarkdownDescription: "Physical address of the site.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *sitesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *sitesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state sitesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Sites = []siteModel{}

	params := map[string]string{}
	stringFilterParam(params, "name", state.Name)
	stringFilterParam(params, "status", state.Status)
	int64FilterParam(params, "region_id", state.RegionId)
	stringFilterParam(params, "tag", state.Tag)

	apiPath := "api/dcim/sites/"
	if q := combineQueryStrings(buildFilterQuery(params), buildCustomFieldFilterQuery(ctx, state.CustomFieldFilters)); q != "" {
		apiPath += "?" + q
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch sites, got error: %s", err))
		return
	}

	type ApiSite struct {
		ID              int64          `json:"id"`
		Name            string         `json:"name"`
		Slug            string         `json:"slug"`
		Status          map[string]any `json:"status"`
		Region          map[string]any `json:"region"`
		Description     string         `json:"description"`
		Facility        string         `json:"facility"`
		TimeZone        string         `json:"time_zone"`
		PhysicalAddress string         `json:"physical_address"`
	}

	type ApiSitesResponse struct {
		Count   int       `json:"count"`
		Results []ApiSite `json:"results"`
	}

	var response ApiSitesResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		s := siteModel{
			Id:              types.Int64Value(result.ID),
			Name:            types.StringValue(result.Name),
			Slug:            types.StringValue(result.Slug),
			Description:     types.StringValue(result.Description),
			Facility:        types.StringValue(result.Facility),
			TimeZone:        types.StringValue(result.TimeZone),
			PhysicalAddress: types.StringValue(result.PhysicalAddress),
		}
		if val, ok := result.Status["value"].(string); ok {
			s.Status = types.StringValue(val)
		} else {
			s.Status = types.StringValue("")
		}
		if result.Region != nil {
			if regionID, ok := result.Region["id"].(float64); ok {
				s.RegionId = types.Int64Value(int64(regionID))
			} else {
				s.RegionId = types.Int64Null()
			}
		} else {
			s.RegionId = types.Int64Null()
		}
		state.Sites = append(state.Sites, s)
	}

	state.Id = types.StringValue("netbox_sites")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
