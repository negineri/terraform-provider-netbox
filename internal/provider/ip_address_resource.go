// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ipAddressResource{}
var _ resource.ResourceWithConfigure = &ipAddressResource{}

func NewIpAddressResource() resource.Resource {
	return &ipAddressResource{}
}

type ipAddressResource struct {
	client *client.NetboxClient
}

type ipAddressResourceModel struct {
	IpAddress     types.String `tfsdk:"ip_address"`
	Status        types.String `tfsdk:"status"`
	Description   types.String `tfsdk:"description"`
	Id            types.Int64  `tfsdk:"id"`
	InterfaceId   types.Int64  `tfsdk:"interface_id"`
	InterfaceType types.String `tfsdk:"interface_type"`
}

func (r *ipAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address"
}

func (r *ipAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an IP address within Netbox.",
		Attributes: map[string]schema.Attribute{
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "The IP address (with or without mask e.g., 192.168.1.10/24).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the IP address (e.g., active, reserved).",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the IP address.",
				Optional:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the allocated IP address.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"interface_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the interface to assign this IP address to.",
				Optional:            true,
			},
			"interface_type": schema.StringAttribute{
				MarkdownDescription: "The type of the interface object (e.g., dcim.interface, virtualization.vminterface).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("dcim.interface"),
			},
		},
	}
}

func (r *ipAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ipAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"address": plan.IpAddress.ValueString(),
		"status":  plan.Status.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.InterfaceId.IsNull() && !plan.InterfaceId.IsUnknown() {
		payload["assigned_object_id"] = plan.InterfaceId.ValueInt64()
		ifType := "dcim.interface"
		if !plan.InterfaceType.IsNull() && !plan.InterfaceType.IsUnknown() {
			ifType = plan.InterfaceType.ValueString()
		}
		payload["assigned_object_type"] = ifType
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	path := "api/ipam/ip-addresses/"
	bodyStr, err := r.client.Post(ctx, path, bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating IP address", err.Error())
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

	// Get saved address from NetBox (can include normalized mask)
	if address, ok := apiResponse["address"].(string); ok {
		plan.IpAddress = types.StringValue(address)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/ipam/ip-addresses/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read IP address, assuming it was deleted", map[string]interface{}{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing read response", err.Error())
		return
	}

	if address, ok := apiResponse["address"].(string); ok {
		state.IpAddress = types.StringValue(address)
	}

	if statusMap, ok := apiResponse["status"].(map[string]interface{}); ok {
		if val, ok := statusMap["value"].(string); ok {
			state.Status = types.StringValue(val)
		}
	}

	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
	}

	if ifType, ok := apiResponse["assigned_object_type"].(string); ok && ifType != "" {
		state.InterfaceType = types.StringValue(ifType)
		if ifIdFloat, ok := apiResponse["assigned_object_id"].(float64); ok {
			state.InterfaceId = types.Int64Value(int64(ifIdFloat))
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ipAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ipAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}
	if !plan.Status.Equal(state.Status) {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.Description.Equal(state.Description) {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.InterfaceId.Equal(state.InterfaceId) || !plan.InterfaceType.Equal(state.InterfaceType) {
		if plan.InterfaceId.IsNull() {
			payload["assigned_object_id"] = nil
			payload["assigned_object_type"] = nil
		} else {
			payload["assigned_object_id"] = plan.InterfaceId.ValueInt64()
			ifType := "dcim.interface"
			if !plan.InterfaceType.IsNull() && !plan.InterfaceType.IsUnknown() {
				ifType = plan.InterfaceType.ValueString()
			}
			payload["assigned_object_type"] = ifType
		}
	}

	if len(payload) > 0 {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		path := fmt.Sprintf("api/ipam/ip-addresses/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating IP address", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/ipam/ip-addresses/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
