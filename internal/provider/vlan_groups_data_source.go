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

var _ datasource.DataSource = &vlanGroupsDataSource{}
var _ datasource.DataSourceWithConfigure = &vlanGroupsDataSource{}

func NewVlanGroupsDataSource() datasource.DataSource {
	return &vlanGroupsDataSource{}
}

type vlanGroupsDataSource struct {
	client *client.NetboxClient
}

type vlanGroupsDataSourceModel struct {
	Id         types.String     `tfsdk:"id"`
	VlanGroups []vlanGroupModel `tfsdk:"vlan_groups"`
}

type vlanGroupModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Description types.String `tfsdk:"description"`
	MinVid      types.Int64  `tfsdk:"min_vid"`
	MaxVid      types.Int64  `tfsdk:"max_vid"`
	ScopeType   types.String `tfsdk:"scope_type"`
	ScopeId     types.Int64  `tfsdk:"scope_id"`
}

func (d *vlanGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_groups"
}

func (d *vlanGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of VLAN groups from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"vlan_groups": schema.ListNestedAttribute{
				MarkdownDescription: "List of VLAN groups.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the VLAN group.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the VLAN group.",
							Computed:            true,
						},
						"slug": schema.StringAttribute{
							MarkdownDescription: "URL-friendly unique shorthand for the VLAN group.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the VLAN group.",
							Computed:            true,
						},
						"min_vid": schema.Int64Attribute{
							MarkdownDescription: "Minimum VLAN ID.",
							Computed:            true,
						},
						"max_vid": schema.Int64Attribute{
							MarkdownDescription: "Maximum VLAN ID.",
							Computed:            true,
						},
						"scope_type": schema.StringAttribute{
							MarkdownDescription: "The type of the scope object.",
							Computed:            true,
						},
						"scope_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the scope object.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *vlanGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vlanGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vlanGroupsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.VlanGroups = []vlanGroupModel{}

	bodyStr, err := d.client.Get(ctx, "api/ipam/vlan-groups/")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch VLAN groups, got error: %s", err))
		return
	}

	type ApiVlanGroup struct {
		ID          int64    `json:"id"`
		Name        string   `json:"name"`
		Slug        string   `json:"slug"`
		Description string   `json:"description"`
		MinVid      *float64 `json:"min_vid"`
		MaxVid      *float64 `json:"max_vid"`
		ScopeType   string   `json:"scope_type"`
		ScopeID     *float64 `json:"scope_id"`
	}

	type ApiVlanGroupsResponse struct {
		Count   int            `json:"count"`
		Results []ApiVlanGroup `json:"results"`
	}

	var response ApiVlanGroupsResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		g := vlanGroupModel{
			Id:          types.Int64Value(result.ID),
			Name:        types.StringValue(result.Name),
			Slug:        types.StringValue(result.Slug),
			Description: types.StringValue(result.Description),
			ScopeType:   types.StringValue(result.ScopeType),
		}
		if result.MinVid != nil {
			g.MinVid = types.Int64Value(int64(*result.MinVid))
		} else {
			g.MinVid = types.Int64Null()
		}
		if result.MaxVid != nil {
			g.MaxVid = types.Int64Value(int64(*result.MaxVid))
		} else {
			g.MaxVid = types.Int64Null()
		}
		if result.ScopeID != nil {
			g.ScopeId = types.Int64Value(int64(*result.ScopeID))
		} else {
			g.ScopeId = types.Int64Null()
		}
		state.VlanGroups = append(state.VlanGroups, g)
	}

	state.Id = types.StringValue("netbox_vlan_groups")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
