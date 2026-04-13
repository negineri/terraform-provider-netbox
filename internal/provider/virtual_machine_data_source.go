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

var _ datasource.DataSource = &virtualMachineDataSource{}
var _ datasource.DataSourceWithConfigure = &virtualMachineDataSource{}

func NewVirtualMachineDataSource() datasource.DataSource {
	return &virtualMachineDataSource{}
}

type virtualMachineDataSource struct {
	client *client.NetboxClient
}

type virtualMachineDataSourceModel struct {
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

func (d *virtualMachineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (d *virtualMachineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single virtual machine from Netbox by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the virtual machine.",
				Required:            true,
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
	}
}

func (d *virtualMachineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *virtualMachineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state virtualMachineDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/virtualization/virtual-machines/%d/", state.Id.ValueInt64())
	bodyStr, err := d.client.Get(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch virtual machine, got error: %s", err))
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
	if statusMap, ok := apiResponse["status"].(map[string]any); ok {
		if val, ok := statusMap["value"].(string); ok {
			state.Status = types.StringValue(val)
		}
	}
	if desc, ok := apiResponse["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}
	if clusterMap, ok := apiResponse["cluster"].(map[string]any); ok {
		if idFloat, ok := clusterMap["id"].(float64); ok {
			state.ClusterId = types.Int64Value(int64(idFloat))
		} else {
			state.ClusterId = types.Int64Null()
		}
	} else {
		state.ClusterId = types.Int64Null()
	}
	if roleMap, ok := apiResponse["role"].(map[string]any); ok {
		if idFloat, ok := roleMap["id"].(float64); ok {
			state.RoleId = types.Int64Value(int64(idFloat))
		} else {
			state.RoleId = types.Int64Null()
		}
	} else {
		state.RoleId = types.Int64Null()
	}
	if siteMap, ok := apiResponse["site"].(map[string]any); ok {
		if idFloat, ok := siteMap["id"].(float64); ok {
			state.SiteId = types.Int64Value(int64(idFloat))
		} else {
			state.SiteId = types.Int64Null()
		}
	} else {
		state.SiteId = types.Int64Null()
	}
	if platformMap, ok := apiResponse["platform"].(map[string]any); ok {
		if idFloat, ok := platformMap["id"].(float64); ok {
			state.PlatformId = types.Int64Value(int64(idFloat))
		} else {
			state.PlatformId = types.Int64Null()
		}
	} else {
		state.PlatformId = types.Int64Null()
	}
	if vcpus, ok := apiResponse["vcpus"].(float64); ok {
		state.Vcpus = types.Int64Value(int64(vcpus))
	} else {
		state.Vcpus = types.Int64Null()
	}
	if memory, ok := apiResponse["memory"].(float64); ok {
		state.Memory = types.Int64Value(int64(memory))
	} else {
		state.Memory = types.Int64Null()
	}
	if disk, ok := apiResponse["disk"].(float64); ok {
		state.Disk = types.Int64Value(int64(disk))
	} else {
		state.Disk = types.Int64Null()
	}

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
