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

var _ datasource.DataSource = &macAddressesDataSource{}
var _ datasource.DataSourceWithConfigure = &macAddressesDataSource{}

func NewMacAddressesDataSource() datasource.DataSource {
	return &macAddressesDataSource{}
}

type macAddressesDataSource struct {
	client *client.NetboxClient
}

type macAddressesDataSourceModel struct {
	Id                 types.String      `tfsdk:"id"`
	MacAddresses       []macAddressModel `tfsdk:"mac_addresses"`
	MacAddressFilter   types.String      `tfsdk:"mac_address"`
	AssignedObjectType types.String      `tfsdk:"assigned_object_type"`
	AssignedObjectId   types.Int64       `tfsdk:"assigned_object_id"`
	Description        types.String      `tfsdk:"description"`
}

type macAddressModel struct {
	Id                 types.Int64  `tfsdk:"id"`
	MacAddress         types.String `tfsdk:"mac_address"`
	AssignedObjectType types.String `tfsdk:"assigned_object_type"`
	AssignedObjectId   types.Int64  `tfsdk:"assigned_object_id"`
	Description        types.String `tfsdk:"description"`
	Comments           types.String `tfsdk:"comments"`
}

func (d *macAddressesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mac_addresses"
}

func (d *macAddressesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of MAC addresses from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"mac_address": schema.StringAttribute{
				MarkdownDescription: "Filter by MAC address.",
				Optional:            true,
			},
			"assigned_object_type": schema.StringAttribute{
				MarkdownDescription: "Filter by assigned object type (e.g., dcim.interface).",
				Optional:            true,
			},
			"assigned_object_id": schema.Int64Attribute{
				MarkdownDescription: "Filter by assigned object ID.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Filter by description.",
				Optional:            true,
			},
			"mac_addresses": schema.ListNestedAttribute{
				MarkdownDescription: "List of MAC addresses.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the MAC address.",
							Computed:            true,
						},
						"mac_address": schema.StringAttribute{
							MarkdownDescription: "The MAC address.",
							Computed:            true,
						},
						"assigned_object_type": schema.StringAttribute{
							MarkdownDescription: "The type of the assigned object.",
							Computed:            true,
						},
						"assigned_object_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the assigned object.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the MAC address.",
							Computed:            true,
						},
						"comments": schema.StringAttribute{
							MarkdownDescription: "Comments for the MAC address.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *macAddressesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *macAddressesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state macAddressesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.MacAddresses = []macAddressModel{}

	params := map[string]string{}
	stringFilterParam(params, "mac_address", state.MacAddressFilter)
	stringFilterParam(params, "assigned_object_type", state.AssignedObjectType)
	int64FilterParam(params, "assigned_object_id", state.AssignedObjectId)
	stringFilterParam(params, "description", state.Description)

	apiPath := "api/dcim/mac-addresses/"
	if q := buildFilterQuery(params); q != "" {
		apiPath += "?" + q
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch MAC addresses, got error: %s", err))
		return
	}

	type ApiMacAddress struct {
		ID                 int64    `json:"id"`
		MacAddress         string   `json:"mac_address"`
		AssignedObjectType *string  `json:"assigned_object_type"`
		AssignedObjectID   *float64 `json:"assigned_object_id"`
		Description        string   `json:"description"`
		Comments           string   `json:"comments"`
	}

	type ApiMacAddressesResponse struct {
		Count   int             `json:"count"`
		Results []ApiMacAddress `json:"results"`
	}

	var response ApiMacAddressesResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		m := macAddressModel{
			Id:          types.Int64Value(result.ID),
			MacAddress:  types.StringValue(result.MacAddress),
			Description: types.StringValue(result.Description),
			Comments:    types.StringValue(result.Comments),
		}
		if result.AssignedObjectType != nil {
			m.AssignedObjectType = types.StringValue(*result.AssignedObjectType)
		} else {
			m.AssignedObjectType = types.StringValue("")
		}
		if result.AssignedObjectID != nil {
			m.AssignedObjectId = types.Int64Value(int64(*result.AssignedObjectID))
		} else {
			m.AssignedObjectId = types.Int64Null()
		}
		state.MacAddresses = append(state.MacAddresses, m)
	}

	state.Id = types.StringValue("netbox_mac_addresses")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
