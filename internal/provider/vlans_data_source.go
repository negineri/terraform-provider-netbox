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

var _ datasource.DataSource = &vlansDataSource{}
var _ datasource.DataSourceWithConfigure = &vlansDataSource{}

func NewVlansDataSource() datasource.DataSource {
	return &vlansDataSource{}
}

type vlansDataSource struct {
	client *client.NetboxClient
}

type vlansDataSourceModel struct {
	Id    types.String `tfsdk:"id"`
	Vlans []vlanModel  `tfsdk:"vlans"`
}

type vlanModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Vid         types.Int64  `tfsdk:"vid"`
	Status      types.String `tfsdk:"status"`
	Description types.String `tfsdk:"description"`
	GroupId     types.Int64  `tfsdk:"group_id"`
}

func (d *vlansDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlans"
}

func (d *vlansDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of VLANs from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"vlans": schema.ListNestedAttribute{
				MarkdownDescription: "List of VLANs.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the VLAN.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the VLAN.",
							Computed:            true,
						},
						"vid": schema.Int64Attribute{
							MarkdownDescription: "VLAN ID (1-4094).",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the VLAN.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the VLAN.",
							Computed:            true,
						},
						"group_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the VLAN group this VLAN belongs to.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *vlansDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *vlansDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state vlansDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.Vlans = []vlanModel{}

	bodyStr, err := d.client.Get(ctx, "api/ipam/vlans/")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch VLANs, got error: %s", err))
		return
	}

	type ApiVlan struct {
		ID          int64                  `json:"id"`
		Name        string                 `json:"name"`
		Vid         float64                `json:"vid"`
		Status      map[string]interface{} `json:"status"`
		Description string                 `json:"description"`
		Group       map[string]interface{} `json:"group"`
	}

	type ApiVlansResponse struct {
		Count   int       `json:"count"`
		Results []ApiVlan `json:"results"`
	}

	var response ApiVlansResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		v := vlanModel{
			Id:          types.Int64Value(result.ID),
			Name:        types.StringValue(result.Name),
			Vid:         types.Int64Value(int64(result.Vid)),
			Description: types.StringValue(result.Description),
		}
		if val, ok := result.Status["value"].(string); ok {
			v.Status = types.StringValue(val)
		} else {
			v.Status = types.StringValue("")
		}
		if result.Group != nil {
			if idFloat, ok := result.Group["id"].(float64); ok {
				v.GroupId = types.Int64Value(int64(idFloat))
			} else {
				v.GroupId = types.Int64Null()
			}
		} else {
			v.GroupId = types.Int64Null()
		}
		state.Vlans = append(state.Vlans, v)
	}

	state.Id = types.StringValue("netbox_vlans")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
