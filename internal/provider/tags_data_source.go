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

var _ datasource.DataSource = &tagsDataSource{}
var _ datasource.DataSourceWithConfigure = &tagsDataSource{}

func NewTagsDataSource() datasource.DataSource {
	return &tagsDataSource{}
}

type tagsDataSource struct {
	client *client.NetboxClient
}

type tagsDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Tags []tagModel   `tfsdk:"tags"`
}

type tagModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Color       types.String `tfsdk:"color"`
	Description types.String `tfsdk:"description"`
}

func (d *tagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}

func (d *tagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of tags from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"tags": schema.ListNestedAttribute{
				MarkdownDescription: "List of tags.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the tag.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the tag.",
							Computed:            true,
						},
						"slug": schema.StringAttribute{
							MarkdownDescription: "URL-friendly unique shorthand for the tag.",
							Computed:            true,
						},
						"color": schema.StringAttribute{
							MarkdownDescription: "The color of the tag as a 6-digit hex code.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the tag.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *tagsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tagsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Tags = []tagModel{}

	bodyStr, err := d.client.Get(ctx, "api/extras/tags/")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch tags, got error: %s", err))
		return
	}

	type ApiTag struct {
		ID          int64  `json:"id"`
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Color       string `json:"color"`
		Description string `json:"description"`
	}

	type ApiTagsResponse struct {
		Count   int      `json:"count"`
		Results []ApiTag `json:"results"`
	}

	var response ApiTagsResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		state.Tags = append(state.Tags, tagModel{
			Id:          types.Int64Value(result.ID),
			Name:        types.StringValue(result.Name),
			Slug:        types.StringValue(result.Slug),
			Color:       types.StringValue(result.Color),
			Description: types.StringValue(result.Description),
		})
	}

	state.Id = types.StringValue("netbox_tags")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
