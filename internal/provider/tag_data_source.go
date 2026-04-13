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

var _ datasource.DataSource = &tagDataSource{}
var _ datasource.DataSourceWithConfigure = &tagDataSource{}

func NewTagDataSource() datasource.DataSource {
	return &tagDataSource{}
}

type tagDataSource struct {
	client *client.NetboxClient
}

type tagDataSourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Color       types.String `tfsdk:"color"`
	Description types.String `tfsdk:"description"`
}

func (d *tagDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (d *tagDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single tag from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the tag.",
				Required:            true,
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
	}
}

func (d *tagDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *tagDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tagDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/extras/tags/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch tag, got error: %s", err))
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
	if color, ok := apiResponse["color"].(string); ok {
		state.Color = types.StringValue(color)
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
