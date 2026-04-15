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

var _ datasource.DataSource = &racksDataSource{}
var _ datasource.DataSourceWithConfigure = &racksDataSource{}

func NewRacksDataSource() datasource.DataSource {
	return &racksDataSource{}
}

type racksDataSource struct {
	client *client.NetboxClient
}

type racksDataSourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Racks              []rackModel  `tfsdk:"racks"`
	CustomFieldFilters types.Map    `tfsdk:"custom_field_filters"`
}

type rackModel struct {
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

func (d *racksDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_racks"
}

func (d *racksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of racks from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"custom_field_filters": schema.MapAttribute{
				MarkdownDescription: "Filter racks by custom field values. Keys are custom field names, values are the filter values.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"racks": schema.ListNestedAttribute{
				MarkdownDescription: "List of racks.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the rack.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *racksDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *racksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state racksDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Racks = []rackModel{}

	apiPath := "api/dcim/racks/"
	if query := buildCustomFieldFilterQuery(ctx, state.CustomFieldFilters); query != "" {
		apiPath += "?" + query
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch racks, got error: %s", err))
		return
	}

	type ApiRack struct {
		ID          int64          `json:"id"`
		Name        string         `json:"name"`
		Site        map[string]any `json:"site"`
		Location    map[string]any `json:"location"`
		Role        map[string]any `json:"role"`
		Status      map[string]any `json:"status"`
		FacilityID  *string        `json:"facility_id"`
		UHeight     int64          `json:"u_height"`
		Description string         `json:"description"`
	}

	type ApiRacksResponse struct {
		Count   int       `json:"count"`
		Results []ApiRack `json:"results"`
	}

	var response ApiRacksResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		rack := rackModel{
			Id:          types.Int64Value(result.ID),
			Name:        types.StringValue(result.Name),
			UHeight:     types.Int64Value(result.UHeight),
			Description: types.StringValue(result.Description),
		}
		if siteID, ok := result.Site["id"].(float64); ok {
			rack.SiteId = types.Int64Value(int64(siteID))
		}
		if result.Location != nil {
			if locationID, ok := result.Location["id"].(float64); ok {
				rack.LocationId = types.Int64Value(int64(locationID))
			} else {
				rack.LocationId = types.Int64Null()
			}
		} else {
			rack.LocationId = types.Int64Null()
		}
		if result.Role != nil {
			if roleID, ok := result.Role["id"].(float64); ok {
				rack.RoleId = types.Int64Value(int64(roleID))
			} else {
				rack.RoleId = types.Int64Null()
			}
		} else {
			rack.RoleId = types.Int64Null()
		}
		if val, ok := result.Status["value"].(string); ok {
			rack.Status = types.StringValue(val)
		} else {
			rack.Status = types.StringValue("")
		}
		if result.FacilityID != nil {
			rack.FacilityId = types.StringValue(*result.FacilityID)
		} else {
			rack.FacilityId = types.StringValue("")
		}
		state.Racks = append(state.Racks, rack)
	}

	state.Id = types.StringValue("netbox_racks")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
