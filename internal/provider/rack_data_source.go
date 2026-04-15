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

var _ datasource.DataSource = &rackDataSource{}
var _ datasource.DataSourceWithConfigure = &rackDataSource{}

func NewRackDataSource() datasource.DataSource {
	return &rackDataSource{}
}

type rackDataSource struct {
	client *client.NetboxClient
}

type rackDataSourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	SiteId      types.Int64  `tfsdk:"site_id"`
	LocationId  types.Int64  `tfsdk:"location_id"`
	RoleId      types.Int64  `tfsdk:"role_id"`
	Status      types.String `tfsdk:"status"`
	FacilityId  types.String `tfsdk:"facility_id"`
	UHeight     types.Int64  `tfsdk:"u_height"`
	Description types.String `tfsdk:"description"`
}

func (d *rackDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rack"
}

func (d *rackDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single rack from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the rack.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the rack.",
				Computed:            true,
			},
			"site_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the site where the rack is located.",
				Computed:            true,
			},
			"location_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the location within the site.",
				Computed:            true,
			},
			"role_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the rack role.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the rack.",
				Computed:            true,
			},
			"facility_id": schema.StringAttribute{
				MarkdownDescription: "A field used to identify the rack by a facility-specific identifier.",
				Computed:            true,
			},
			"u_height": schema.Int64Attribute{
				MarkdownDescription: "Height in rack units.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the rack.",
				Computed:            true,
			},
		},
	}
}

func (d *rackDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *rackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state rackDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/racks/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch rack, got error: %s", err))
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
	if siteMap, ok := apiResponse["site"].(map[string]any); ok {
		if idFloat, ok := siteMap["id"].(float64); ok {
			state.SiteId = types.Int64Value(int64(idFloat))
		}
	}
	if locationMap, ok := apiResponse["location"].(map[string]any); ok {
		if idFloat, ok := locationMap["id"].(float64); ok {
			state.LocationId = types.Int64Value(int64(idFloat))
		}
	} else {
		state.LocationId = types.Int64Null()
	}
	if roleMap, ok := apiResponse["role"].(map[string]any); ok {
		if idFloat, ok := roleMap["id"].(float64); ok {
			state.RoleId = types.Int64Value(int64(idFloat))
		}
	} else {
		state.RoleId = types.Int64Null()
	}
	if statusMap, ok := apiResponse["status"].(map[string]any); ok {
		if val, ok := statusMap["value"].(string); ok {
			state.Status = types.StringValue(val)
		}
	}
	if facilityId, ok := apiResponse["facility_id"].(string); ok {
		state.FacilityId = types.StringValue(facilityId)
	} else {
		state.FacilityId = types.StringValue("")
	}
	if uHeight, ok := apiResponse["u_height"].(float64); ok {
		state.UHeight = types.Int64Value(int64(uHeight))
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
