// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &virtualMachineResource{}
var _ resource.ResourceWithConfigure = &virtualMachineResource{}

func NewVirtualMachineResource() resource.Resource {
	return &virtualMachineResource{}
}

type virtualMachineResource struct {
	client *client.NetboxClient
}

type virtualMachineResourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ClusterId    types.Int64  `tfsdk:"cluster_id"`
	Status       types.String `tfsdk:"status"`
	RoleId       types.Int64  `tfsdk:"role_id"`
	SiteId       types.Int64  `tfsdk:"site_id"`
	PlatformId   types.Int64  `tfsdk:"platform_id"`
	Vcpus        types.Int64  `tfsdk:"vcpus"`
	Memory       types.Int64  `tfsdk:"memory"`
	Disk         types.Int64  `tfsdk:"disk"`
	Description  types.String `tfsdk:"description"`
	Tags         types.List   `tfsdk:"tags"`
	CustomFields types.Map    `tfsdk:"custom_fields"`
}

func (r *virtualMachineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine"
}

func (r *virtualMachineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Virtual Machine in Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the virtual machine.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the virtual machine.",
				Required:            true,
			},
			"cluster_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the cluster where the virtual machine is hosted.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the virtual machine (e.g., active, offline, planned, staged, failed, decommissioning).",
				Optional:            true,
			},
			"role_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the role assigned to the virtual machine.",
				Optional:            true,
			},
			"site_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the site where the virtual machine is located.",
				Optional:            true,
			},
			"platform_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the platform (OS) of the virtual machine.",
				Optional:            true,
			},
			"vcpus": schema.Int64Attribute{
				MarkdownDescription: "The number of virtual CPUs allocated to the virtual machine.",
				Optional:            true,
			},
			"memory": schema.Int64Attribute{
				MarkdownDescription: "The amount of memory (in MB) allocated to the virtual machine.",
				Optional:            true,
			},
			"disk": schema.Int64Attribute{
				MarkdownDescription: "The disk size (in GB) allocated to the virtual machine.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the virtual machine.",
				Optional:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.Int64Type,
				MarkdownDescription: "List of tag IDs to assign to the virtual machine.",
				Optional:            true,
			},
			"custom_fields": customFieldsSchema(),
		},
	}
}

func (r *virtualMachineResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.NetboxClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.NetboxClient, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *virtualMachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan virtualMachineResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}
	if !plan.ClusterId.IsNull() && !plan.ClusterId.IsUnknown() {
		payload["cluster"] = map[string]interface{}{"id": plan.ClusterId.ValueInt64()}
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.RoleId.IsNull() && !plan.RoleId.IsUnknown() {
		payload["role"] = map[string]interface{}{"id": plan.RoleId.ValueInt64()}
	}
	if !plan.SiteId.IsNull() && !plan.SiteId.IsUnknown() {
		payload["site"] = map[string]interface{}{"id": plan.SiteId.ValueInt64()}
	}
	if !plan.PlatformId.IsNull() && !plan.PlatformId.IsUnknown() {
		payload["platform"] = map[string]interface{}{"id": plan.PlatformId.ValueInt64()}
	}
	if !plan.Vcpus.IsNull() && !plan.Vcpus.IsUnknown() {
		payload["vcpus"] = plan.Vcpus.ValueInt64()
	}
	if !plan.Memory.IsNull() && !plan.Memory.IsUnknown() {
		payload["memory"] = plan.Memory.ValueInt64()
	}
	if !plan.Disk.IsNull() && !plan.Disk.IsUnknown() {
		payload["disk"] = plan.Disk.ValueInt64()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tagIDs []int64
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tagIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		tags := make([]map[string]interface{}, len(tagIDs))
		for i, id := range tagIDs {
			tags[i] = map[string]interface{}{"id": id}
		}
		payload["tags"] = tags
	}
	if cf := customFieldsToPayload(ctx, plan.CustomFields, &resp.Diagnostics); cf != nil {
		payload["custom_fields"] = cf
	}
	if resp.Diagnostics.HasError() {
		return
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	bodyStr, err := r.client.Post(ctx, "api/virtualization/virtual-machines/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating Virtual Machine", err.Error())
		return
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing create response", err.Error())
		return
	}

	idFloat, ok := apiResponse["id"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Error parsing create response", "Could not find 'id' in response")
		return
	}
	plan.Id = types.Int64Value(int64(idFloat))

	if cfRaw, ok := apiResponse["custom_fields"]; ok {
		cf, diags := customFieldsFromAPI(cfRaw)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			plan.CustomFields = cf
		}
	} else {
		plan.CustomFields = types.MapValueMust(types.StringType, map[string]attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *virtualMachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state virtualMachineResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/virtualization/virtual-machines/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read virtual machine, assuming it was deleted", map[string]interface{}{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing read response", err.Error())
		return
	}

	if name, ok := apiResponse["name"].(string); ok {
		state.Name = types.StringValue(name)
	}

	if clusterMap, ok := apiResponse["cluster"].(map[string]interface{}); ok {
		if idFloat, ok := clusterMap["id"].(float64); ok {
			state.ClusterId = types.Int64Value(int64(idFloat))
		}
	} else if !state.ClusterId.IsNull() {
		state.ClusterId = types.Int64Null()
	}

	if statusMap, ok := apiResponse["status"].(map[string]interface{}); ok {
		if val, ok := statusMap["value"].(string); ok && !state.Status.IsNull() {
			state.Status = types.StringValue(val)
		}
	}

	if roleMap, ok := apiResponse["role"].(map[string]interface{}); ok {
		if idFloat, ok := roleMap["id"].(float64); ok {
			state.RoleId = types.Int64Value(int64(idFloat))
		}
	} else if !state.RoleId.IsNull() {
		state.RoleId = types.Int64Null()
	}

	if siteMap, ok := apiResponse["site"].(map[string]interface{}); ok {
		if idFloat, ok := siteMap["id"].(float64); ok && !state.SiteId.IsNull() {
			state.SiteId = types.Int64Value(int64(idFloat))
		}
	} else if !state.SiteId.IsNull() {
		state.SiteId = types.Int64Null()
	}

	if platformMap, ok := apiResponse["platform"].(map[string]interface{}); ok {
		if idFloat, ok := platformMap["id"].(float64); ok {
			state.PlatformId = types.Int64Value(int64(idFloat))
		}
	} else if !state.PlatformId.IsNull() {
		state.PlatformId = types.Int64Null()
	}

	if vcpus, ok := apiResponse["vcpus"].(float64); ok && !state.Vcpus.IsNull() {
		state.Vcpus = types.Int64Value(int64(vcpus))
	}

	if memory, ok := apiResponse["memory"].(float64); ok && !state.Memory.IsNull() {
		state.Memory = types.Int64Value(int64(memory))
	}

	if disk, ok := apiResponse["disk"].(float64); ok && !state.Disk.IsNull() {
		state.Disk = types.Int64Value(int64(disk))
	}

	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
	}

	if !state.Tags.IsNull() {
		if tagsRaw, ok := apiResponse["tags"].([]interface{}); ok {
			tagVals := make([]attr.Value, 0, len(tagsRaw))
			for _, t := range tagsRaw {
				if tagMap, ok := t.(map[string]interface{}); ok {
					if idFloat, ok := tagMap["id"].(float64); ok {
						tagVals = append(tagVals, types.Int64Value(int64(idFloat)))
					}
				}
			}
			listVal, diags := types.ListValue(types.Int64Type, tagVals)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				state.Tags = listVal
			}
		}
	}

	if cfRaw, ok := apiResponse["custom_fields"]; ok {
		cf, diags := customFieldsFromAPI(cfRaw)
		resp.Diagnostics.Append(diags...)
		if !resp.Diagnostics.HasError() {
			state.CustomFields = cf
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *virtualMachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state virtualMachineResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}
	if !plan.Name.Equal(state.Name) {
		payload["name"] = plan.Name.ValueString()
	}
	if !plan.ClusterId.Equal(state.ClusterId) {
		if plan.ClusterId.IsNull() {
			payload["cluster"] = nil
		} else {
			payload["cluster"] = map[string]interface{}{"id": plan.ClusterId.ValueInt64()}
		}
	}
	if !plan.Status.Equal(state.Status) {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.RoleId.Equal(state.RoleId) {
		if plan.RoleId.IsNull() {
			payload["role"] = nil
		} else {
			payload["role"] = map[string]interface{}{"id": plan.RoleId.ValueInt64()}
		}
	}
	if !plan.SiteId.Equal(state.SiteId) {
		if plan.SiteId.IsNull() {
			payload["site"] = nil
		} else {
			payload["site"] = map[string]interface{}{"id": plan.SiteId.ValueInt64()}
		}
	}
	if !plan.PlatformId.Equal(state.PlatformId) {
		if plan.PlatformId.IsNull() {
			payload["platform"] = nil
		} else {
			payload["platform"] = map[string]interface{}{"id": plan.PlatformId.ValueInt64()}
		}
	}
	if !plan.Vcpus.Equal(state.Vcpus) {
		if plan.Vcpus.IsNull() {
			payload["vcpus"] = nil
		} else {
			payload["vcpus"] = plan.Vcpus.ValueInt64()
		}
	}
	if !plan.Memory.Equal(state.Memory) {
		if plan.Memory.IsNull() {
			payload["memory"] = nil
		} else {
			payload["memory"] = plan.Memory.ValueInt64()
		}
	}
	if !plan.Disk.Equal(state.Disk) {
		if plan.Disk.IsNull() {
			payload["disk"] = nil
		} else {
			payload["disk"] = plan.Disk.ValueInt64()
		}
	}
	if !plan.Description.Equal(state.Description) {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.Tags.Equal(state.Tags) {
		if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
			var tagIDs []int64
			resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tagIDs, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			tags := make([]map[string]interface{}, len(tagIDs))
			for i, id := range tagIDs {
				tags[i] = map[string]interface{}{"id": id}
			}
			payload["tags"] = tags
		} else {
			payload["tags"] = []map[string]interface{}{}
		}
	}
	if !plan.CustomFields.Equal(state.CustomFields) {
		if cf := customFieldsToPayload(ctx, plan.CustomFields, &resp.Diagnostics); cf != nil {
			payload["custom_fields"] = cf
		}
		if resp.Diagnostics.HasError() {
			return
		}
	}

	if len(payload) > 0 {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		path := fmt.Sprintf("api/virtualization/virtual-machines/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating Virtual Machine", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *virtualMachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state virtualMachineResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/virtualization/virtual-machines/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
