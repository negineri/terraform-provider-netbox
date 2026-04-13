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

var _ datasource.DataSource = &ipAddressesDataSource{}
var _ datasource.DataSourceWithConfigure = &ipAddressesDataSource{}

func NewIpAddressesDataSource() datasource.DataSource {
	return &ipAddressesDataSource{}
}

type ipAddressesDataSource struct {
	client *client.NetboxClient
}

type ipAddressesDataSourceModel struct {
	Id          types.String     `tfsdk:"id"`
	IpAddresses []ipAddressModel `tfsdk:"ip_addresses"`
}

type ipAddressModel struct {
	Id            types.Int64  `tfsdk:"id"`
	IpAddress     types.String `tfsdk:"ip_address"`
	Status        types.String `tfsdk:"status"`
	Description   types.String `tfsdk:"description"`
	DnsName       types.String `tfsdk:"dns_name"`
	InterfaceId   types.Int64  `tfsdk:"interface_id"`
	InterfaceType types.String `tfsdk:"interface_type"`
}

func (d *ipAddressesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_addresses"
}

func (d *ipAddressesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of IP addresses from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"ip_addresses": schema.ListNestedAttribute{
				MarkdownDescription: "List of IP addresses.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the IP address.",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *ipAddressesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ipAddressesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ipAddressesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.IpAddresses = []ipAddressModel{}

	bodyStr, err := d.client.Get(ctx, "api/ipam/ip-addresses/")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch IP addresses, got error: %s", err))
		return
	}

	type ApiIPAddress struct {
		ID                 int64                  `json:"id"`
		Address            string                 `json:"address"`
		Status             map[string]interface{} `json:"status"`
		Description        string                 `json:"description"`
		DnsName            string                 `json:"dns_name"`
		AssignedObjectID   *float64               `json:"assigned_object_id"`
		AssignedObjectType string                 `json:"assigned_object_type"`
	}

	type ApiIPAddressesResponse struct {
		Count   int            `json:"count"`
		Results []ApiIPAddress `json:"results"`
	}

	var response ApiIPAddressesResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		m := ipAddressModel{
			Id:            types.Int64Value(result.ID),
			IpAddress:     types.StringValue(result.Address),
			Description:   types.StringValue(result.Description),
			DnsName:       types.StringValue(result.DnsName),
			InterfaceType: types.StringValue(result.AssignedObjectType),
		}
		if val, ok := result.Status["value"].(string); ok {
			m.Status = types.StringValue(val)
		} else {
			m.Status = types.StringValue("")
		}
		if result.AssignedObjectID != nil {
			m.InterfaceId = types.Int64Value(int64(*result.AssignedObjectID))
		} else {
			m.InterfaceId = types.Int64Null()
		}
		state.IpAddresses = append(state.IpAddresses, m)
	}

	state.Id = types.StringValue("netbox_ip_addresses")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
