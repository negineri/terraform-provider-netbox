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

var _ datasource.DataSource = &virtualMachinesDataSource{}
var _ datasource.DataSourceWithConfigure = &virtualMachinesDataSource{}

func NewVirtualMachinesDataSource() datasource.DataSource {
	return &virtualMachinesDataSource{}
}

type virtualMachinesDataSource struct {
	client *client.NetboxClient
}

type virtualMachinesDataSourceModel struct {
	Id              types.String          `tfsdk:"id"`
	VirtualMachines []virtualMachineModel `tfsdk:"virtual_machines"`
}

type virtualMachineModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	ClusterId   types.Int64  `tfsdk:"cluster_id"`
	Status      types.String `tfsdk:"status"`
	RoleId      types.Int64  `tfsdk:"role_id"`
	SiteId      types.Int64  `tfsdk:"site_id"`
	PlatformId  types.Int64  `tfsdk:"platform_id"`
	Vcpus       types.Int64  `tfsdk:"vcpus"`
	Memory      types.Int64  `tfsdk:"memory"`
	Disk        types.Int64  `tfsdk:"disk"`
	Description types.String `tfsdk:"description"`
}

func (d *virtualMachinesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machines"
}

func (d *virtualMachinesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of virtual machines from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"virtual_machines": schema.ListNestedAttribute{
				MarkdownDescription: "List of virtual machines.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "The numeric ID of the virtual machine.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the virtual machine.",
							Computed:            true,
						},
						"cluster_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the cluster where the virtual machine is hosted.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the virtual machine.",
							Computed:            true,
						},
						"role_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the role assigned to the virtual machine.",
							Computed:            true,
						},
						"site_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the site where the virtual machine is located.",
							Computed:            true,
						},
						"platform_id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the platform (OS) of the virtual machine.",
							Computed:            true,
						},
						"vcpus": schema.Int64Attribute{
							MarkdownDescription: "The number of virtual CPUs allocated.",
							Computed:            true,
						},
						"memory": schema.Int64Attribute{
							MarkdownDescription: "The amount of memory (in MB) allocated.",
							Computed:            true,
						},
						"disk": schema.Int64Attribute{
							MarkdownDescription: "The disk size (in GB) allocated.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description for the virtual machine.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *virtualMachinesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *virtualMachinesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state virtualMachinesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.VirtualMachines = []virtualMachineModel{}

	bodyStr, err := d.client.Get(ctx, "api/virtualization/virtual-machines/")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch virtual machines, got error: %s", err))
		return
	}

	type ApiVirtualMachine struct {
		ID          int64                  `json:"id"`
		Name        string                 `json:"name"`
		Cluster     map[string]interface{} `json:"cluster"`
		Status      map[string]interface{} `json:"status"`
		Role        map[string]interface{} `json:"role"`
		Site        map[string]interface{} `json:"site"`
		Platform    map[string]interface{} `json:"platform"`
		Vcpus       *float64               `json:"vcpus"`
		Memory      *float64               `json:"memory"`
		Disk        *float64               `json:"disk"`
		Description string                 `json:"description"`
	}

	type ApiVirtualMachinesResponse struct {
		Count   int                 `json:"count"`
		Results []ApiVirtualMachine `json:"results"`
	}

	var response ApiVirtualMachinesResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		vm := virtualMachineModel{
			Id:          types.Int64Value(result.ID),
			Name:        types.StringValue(result.Name),
			Description: types.StringValue(result.Description),
		}
		if val, ok := result.Status["value"].(string); ok {
			vm.Status = types.StringValue(val)
		} else {
			vm.Status = types.StringValue("")
		}
		if result.Cluster != nil {
			if idFloat, ok := result.Cluster["id"].(float64); ok {
				vm.ClusterId = types.Int64Value(int64(idFloat))
			} else {
				vm.ClusterId = types.Int64Null()
			}
		} else {
			vm.ClusterId = types.Int64Null()
		}
		if result.Role != nil {
			if idFloat, ok := result.Role["id"].(float64); ok {
				vm.RoleId = types.Int64Value(int64(idFloat))
			} else {
				vm.RoleId = types.Int64Null()
			}
		} else {
			vm.RoleId = types.Int64Null()
		}
		if result.Site != nil {
			if idFloat, ok := result.Site["id"].(float64); ok {
				vm.SiteId = types.Int64Value(int64(idFloat))
			} else {
				vm.SiteId = types.Int64Null()
			}
		} else {
			vm.SiteId = types.Int64Null()
		}
		if result.Platform != nil {
			if idFloat, ok := result.Platform["id"].(float64); ok {
				vm.PlatformId = types.Int64Value(int64(idFloat))
			} else {
				vm.PlatformId = types.Int64Null()
			}
		} else {
			vm.PlatformId = types.Int64Null()
		}
		if result.Vcpus != nil {
			vm.Vcpus = types.Int64Value(int64(*result.Vcpus))
		} else {
			vm.Vcpus = types.Int64Null()
		}
		if result.Memory != nil {
			vm.Memory = types.Int64Value(int64(*result.Memory))
		} else {
			vm.Memory = types.Int64Null()
		}
		if result.Disk != nil {
			vm.Disk = types.Int64Value(int64(*result.Disk))
		} else {
			vm.Disk = types.Int64Null()
		}
		state.VirtualMachines = append(state.VirtualMachines, vm)
	}

	state.Id = types.StringValue("netbox_virtual_machines")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
