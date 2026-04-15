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

var _ datasource.DataSource = &manufacturersDataSource{}
var _ datasource.DataSourceWithConfigure = &manufacturersDataSource{}

func NewManufacturersDataSource() datasource.DataSource {
	return &manufacturersDataSource{}
}

type manufacturersDataSource struct {
	client *client.NetboxClient
}

type manufacturersDataSourceModel struct {
	Id                 types.String        `tfsdk:"id"`
	Manufacturers      []manufacturerModel `tfsdk:"manufacturers"`
	CustomFieldFilters types.Map           `tfsdk:"custom_field_filters"`
	Name               types.String        `tfsdk:"name"`
	Slug               types.String        `tfsdk:"slug"`
}

type manufacturerModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Description types.String `tfsdk:"description"`
}

func (d *manufacturersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_manufacturers"
}

func (d *manufacturersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of manufacturers from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"custom_field_filters": schema.MapAttribute{
				MarkdownDescription: "Filter manufacturers by custom field values. Keys are custom field names, values are the filter values.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Filter by manufacturer name.",
				Optional:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "Filter by manufacturer slug.",
				Optional:            true,
			},
			"manufacturers": schema.ListNestedAttribute{
				MarkdownDescription: "List of manufacturers.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the manufacturer.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the manufacturer.",
							Computed:            true,
						},
						"slug": schema.StringAttribute{
							MarkdownDescription: "URL-friendly unique shorthand for the manufacturer.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the manufacturer.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *manufacturersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *manufacturersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state manufacturersDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Manufacturers = []manufacturerModel{}

	params := map[string]string{}
	stringFilterParam(params, "name", state.Name)
	stringFilterParam(params, "slug", state.Slug)

	apiPath := "api/dcim/manufacturers/"
	if q := combineQueryStrings(buildFilterQuery(params), buildCustomFieldFilterQuery(ctx, state.CustomFieldFilters)); q != "" {
		apiPath += "?" + q
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch manufacturers, got error: %s", err))
		return
	}

	type ApiManufacturer struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description"`
	}

	type ApiManufacturersResponse struct {
		Count   int               `json:"count"`
		Results []ApiManufacturer `json:"results"`
	}

	var response ApiManufacturersResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		m := manufacturerModel{
			Id:          types.Int64Value(result.ID),
			Name:        types.StringValue(result.Name),
			Slug:        types.StringValue(result.Slug),
			Description: types.StringValue(result.Description),
		}
		state.Manufacturers = append(state.Manufacturers, m)
	}

	state.Id = types.StringValue("netbox_manufacturers")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
