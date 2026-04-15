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

var _ datasource.DataSource = &regionsDataSource{}
var _ datasource.DataSourceWithConfigure = &regionsDataSource{}

func NewRegionsDataSource() datasource.DataSource {
	return &regionsDataSource{}
}

type regionsDataSource struct {
	client *client.NetboxClient
}

type regionsDataSourceModel struct {
	Id                 types.String  `tfsdk:"id"`
	Regions            []regionModel `tfsdk:"regions"`
	CustomFieldFilters types.Map     `tfsdk:"custom_field_filters"`
}

type regionModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	ParentId    types.Int64  `tfsdk:"parent_id"`
	Description types.String `tfsdk:"description"`
}

func (d *regionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_regions"
}

func (d *regionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of regions from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"custom_field_filters": schema.MapAttribute{
				MarkdownDescription: "Filter regions by custom field values. Keys are custom field names, values are the filter values.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"regions": schema.ListNestedAttribute{
				MarkdownDescription: "List of regions.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the region.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the region.",
							Computed:            true,
						},
						"slug": schema.StringAttribute{
							MarkdownDescription: "URL-friendly unique shorthand for the region.",
							Computed:            true,
						},
						"parent_id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the parent region.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the region.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *regionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *regionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state regionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Regions = []regionModel{}

	apiPath := "api/dcim/regions/"
	if query := buildCustomFieldFilterQuery(ctx, state.CustomFieldFilters); query != "" {
		apiPath += "?" + query
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch regions, got error: %s", err))
		return
	}

	type ApiRegion struct {
		ID          int64          `json:"id"`
		Name        string         `json:"name"`
		Slug        string         `json:"slug"`
		Parent      map[string]any `json:"parent"`
		Description string         `json:"description"`
	}

	type ApiRegionsResponse struct {
		Count   int         `json:"count"`
		Results []ApiRegion `json:"results"`
	}

	var response ApiRegionsResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		reg := regionModel{
			Id:          types.Int64Value(result.ID),
			Name:        types.StringValue(result.Name),
			Slug:        types.StringValue(result.Slug),
			Description: types.StringValue(result.Description),
		}
		if result.Parent != nil {
			if parentID, ok := result.Parent["id"].(float64); ok {
				reg.ParentId = types.Int64Value(int64(parentID))
			} else {
				reg.ParentId = types.Int64Null()
			}
		} else {
			reg.ParentId = types.Int64Null()
		}
		state.Regions = append(state.Regions, reg)
	}

	state.Id = types.StringValue("netbox_regions")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
