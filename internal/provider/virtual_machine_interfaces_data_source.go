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

var _ datasource.DataSource = &virtualMachineInterfacesDataSource{}
var _ datasource.DataSourceWithConfigure = &virtualMachineInterfacesDataSource{}

func NewVirtualMachineInterfacesDataSource() datasource.DataSource {
	return &virtualMachineInterfacesDataSource{}
}

type virtualMachineInterfacesDataSource struct {
	client *client.NetboxClient
}

type virtualMachineInterfacesDataSourceModel struct {
	Id                       types.String                   `tfsdk:"id"`
	VirtualMachineInterfaces []virtualMachineInterfaceModel `tfsdk:"virtual_machine_interfaces"`
	VirtualMachineId         types.Int64                    `tfsdk:"virtual_machine_id"`
	Name                     types.String                   `tfsdk:"name"`
	Enabled                  types.Bool                     `tfsdk:"enabled"`
}

type virtualMachineInterfaceModel struct {
	Id               types.Int64  `tfsdk:"id"`
	VirtualMachineId types.Int64  `tfsdk:"virtual_machine_id"`
	Name             types.String `tfsdk:"name"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	MacAddress       types.String `tfsdk:"mac_address"`
	Mtu              types.Int64  `tfsdk:"mtu"`
	Description      types.String `tfsdk:"description"`
}

func (d *virtualMachineInterfacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine_interfaces"
}

func (d *virtualMachineInterfacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of Virtual Machine Interfaces from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"virtual_machine_id": schema.Int64Attribute{
				MarkdownDescription: "Filter by virtual machine ID.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Filter by interface name.",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Filter by enabled state.",
				Optional:            true,
			},
			"virtual_machine_interfaces": schema.ListNestedAttribute{
				MarkdownDescription: "List of virtual machine interfaces.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the interface.",
							Computed:            true,
						},
						"virtual_machine_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the virtual machine.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the interface.",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the interface is enabled.",
							Computed:            true,
						},
						"mac_address": schema.StringAttribute{
							MarkdownDescription: "The MAC address of the interface.",
							Computed:            true,
						},
						"mtu": schema.Int64Attribute{
							MarkdownDescription: "The MTU of the interface.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the interface.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *virtualMachineInterfacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *virtualMachineInterfacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state virtualMachineInterfacesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.VirtualMachineInterfaces = []virtualMachineInterfaceModel{}

	params := map[string]string{}
	int64FilterParam(params, "virtual_machine_id", state.VirtualMachineId)
	stringFilterParam(params, "name", state.Name)
	if !state.Enabled.IsNull() && !state.Enabled.IsUnknown() {
		if state.Enabled.ValueBool() {
			params["enabled"] = "true"
		} else {
			params["enabled"] = "false"
		}
	}

	apiPath := "api/virtualization/interfaces/"
	if q := buildFilterQuery(params); q != "" {
		apiPath += "?" + q
	}

	bodyStr, err := d.client.Get(ctx, apiPath)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch virtual machine interfaces, got error: %s", err))
		return
	}

	type ApiVMInterface struct {
		ID             int64                  `json:"id"`
		VirtualMachine map[string]interface{} `json:"virtual_machine"`
		Name           string                 `json:"name"`
		Enabled        bool                   `json:"enabled"`
		MacAddress     *string                `json:"mac_address"`
		Mtu            *float64               `json:"mtu"`
		Description    string                 `json:"description"`
	}

	type ApiResponse struct {
		Count   int              `json:"count"`
		Results []ApiVMInterface `json:"results"`
	}

	var response ApiResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		m := virtualMachineInterfaceModel{
			Id:          types.Int64Value(result.ID),
			Name:        types.StringValue(result.Name),
			Enabled:     types.BoolValue(result.Enabled),
			Description: types.StringValue(result.Description),
		}
		if idFloat, ok := result.VirtualMachine["id"].(float64); ok {
			m.VirtualMachineId = types.Int64Value(int64(idFloat))
		}
		if result.MacAddress != nil {
			m.MacAddress = types.StringValue(*result.MacAddress)
		} else {
			m.MacAddress = types.StringValue("")
		}
		if result.Mtu != nil {
			m.Mtu = types.Int64Value(int64(*result.Mtu))
		} else {
			m.Mtu = types.Int64Null()
		}
		state.VirtualMachineInterfaces = append(state.VirtualMachineInterfaces, m)
	}

	state.Id = types.StringValue("netbox_virtual_machine_interfaces")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
