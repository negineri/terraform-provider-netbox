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

var _ datasource.DataSource = &ipAddressDataSource{}
var _ datasource.DataSourceWithConfigure = &ipAddressDataSource{}

func NewIpAddressDataSource() datasource.DataSource {
	return &ipAddressDataSource{}
}

type ipAddressDataSource struct {
	client *client.NetboxClient
}

type ipAddressDataSourceModel struct {
	Id            types.Int64  `tfsdk:"id"`
	IpAddress     types.String `tfsdk:"ip_address"`
	Status        types.String `tfsdk:"status"`
	Description   types.String `tfsdk:"description"`
	DnsName       types.String `tfsdk:"dns_name"`
	InterfaceId   types.Int64  `tfsdk:"interface_id"`
	InterfaceType types.String `tfsdk:"interface_type"`
}

func (d *ipAddressDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address"
}

func (d *ipAddressDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single IP address from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the IP address.",
				Required:            true,
			},
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "The IP address in CIDR notation.",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the IP address.",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the IP address.",
				Computed:            true,
			},
			"dns_name": schema.StringAttribute{
				MarkdownDescription: "DNS name associated with the IP address.",
				Computed:            true,
			},
			"interface_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the assigned interface.",
				Computed:            true,
			},
			"interface_type": schema.StringAttribute{
				MarkdownDescription: "The type of the assigned interface object.",
				Computed:            true,
			},
		},
	}
}

func (d *ipAddressDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ipAddressDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ipAddressDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/ipam/ip-addresses/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch IP address, got error: %s", err))
		return
	}

	var apiResponse map[string]any
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	if address, ok := apiResponse["address"].(string); ok {
		state.IpAddress = types.StringValue(address)
	}
	if statusMap, ok := apiResponse["status"].(map[string]any); ok {
		if val, ok := statusMap["value"].(string); ok {
			state.Status = types.StringValue(val)
		}
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}
	if dnsName, ok := apiResponse["dns_name"].(string); ok {
		state.DnsName = types.StringValue(dnsName)
	}
	if ifType, ok := apiResponse["assigned_object_type"].(string); ok {
		state.InterfaceType = types.StringValue(ifType)
	} else {
		state.InterfaceType = types.StringValue("")
	}
	if ifIdFloat, ok := apiResponse["assigned_object_id"].(float64); ok {
		state.InterfaceId = types.Int64Value(int64(ifIdFloat))
	} else {
		state.InterfaceId = types.Int64Null()
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
