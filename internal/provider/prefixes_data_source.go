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

var _ datasource.DataSource = &prefixesDataSource{}
var _ datasource.DataSourceWithConfigure = &prefixesDataSource{}

func NewPrefixesDataSource() datasource.DataSource {
	return &prefixesDataSource{}
}

type prefixesDataSource struct {
	client *client.NetboxClient
}

type prefixesDataSourceModel struct {
	Id       types.String  `tfsdk:"id"`
	Prefixes []prefixModel `tfsdk:"prefixes"`
}

type prefixModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Prefix      types.String `tfsdk:"prefix"`
	Status      types.String `tfsdk:"status"`
	Description types.String `tfsdk:"description"`
}

func (d *prefixesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_prefixes"
}

func (d *prefixesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of prefixes from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"prefixes": schema.ListNestedAttribute{
				MarkdownDescription: "List of prefixes.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the prefix.",
							Computed:            true,
						},
						"prefix": schema.StringAttribute{
							MarkdownDescription: "The subnet in CIDR format.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the prefix.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the prefix.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *prefixesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *prefixesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state prefixesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Prefixes = []prefixModel{}

	bodyStr, err := d.client.Get(ctx, "api/ipam/prefixes/")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch prefixes, got error: %s", err))
		return
	}

	type ApiPrefix struct {
		ID          int64                  `json:"id"`
		Prefix      string                 `json:"prefix"`
		Status      map[string]interface{} `json:"status"`
		Description string                 `json:"description"`
	}

	type ApiPrefixesResponse struct {
		Count   int         `json:"count"`
		Results []ApiPrefix `json:"results"`
	}

	var response ApiPrefixesResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		p := prefixModel{
			Id:          types.Int64Value(result.ID),
			Prefix:      types.StringValue(result.Prefix),
			Description: types.StringValue(result.Description),
		}
		if val, ok := result.Status["value"].(string); ok {
			p.Status = types.StringValue(val)
		} else {
			p.Status = types.StringValue("")
		}
		state.Prefixes = append(state.Prefixes, p)
	}

	state.Id = types.StringValue("netbox_prefixes")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
