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

var _ datasource.DataSource = &siteDataSource{}
var _ datasource.DataSourceWithConfigure = &siteDataSource{}

func NewSiteDataSource() datasource.DataSource {
	return &siteDataSource{}
}

type siteDataSource struct {
	client *client.NetboxClient
}

type siteDataSourceModel struct {
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

func (d *siteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site"
}

func (d *siteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single site from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the site.",
				Required:            true,
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
	}
}

func (d *siteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *siteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state siteDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/sites/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch site, got error: %s", err))
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
	if statusMap, ok := apiResponse["status"].(map[string]any); ok {
		if val, ok := statusMap["value"].(string); ok {
			state.Status = types.StringValue(val)
		}
	}
	if regionMap, ok := apiResponse["region"].(map[string]any); ok {
		if regionID, ok := regionMap["id"].(float64); ok {
			state.RegionId = types.Int64Value(int64(regionID))
		}
	} else {
		state.RegionId = types.Int64Null()
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}
	if facility, ok := apiResponse["facility"].(string); ok {
		state.Facility = types.StringValue(facility)
	}
	if tz, ok := apiResponse["time_zone"].(string); ok {
		state.TimeZone = types.StringValue(tz)
	}
	if addr, ok := apiResponse["physical_address"].(string); ok {
		state.PhysicalAddress = types.StringValue(addr)
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
