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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &deviceInterfaceResource{}
var _ resource.ResourceWithConfigure = &deviceInterfaceResource{}

func NewDeviceInterfaceResource() resource.Resource {
	return &deviceInterfaceResource{}
}

type deviceInterfaceResource struct {
	client *client.NetboxClient
}

type deviceInterfaceResourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	DeviceId    types.Int64  `tfsdk:"device_id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	Enabled     types.Bool   `tfsdk:"enabled"`
	MacAddress  types.String `tfsdk:"mac_address"`
	Mtu         types.Int64  `tfsdk:"mtu"`
	Description types.String `tfsdk:"description"`
}

func (r *deviceInterfaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_interface"
}

func (r *deviceInterfaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Device Interface in Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the interface.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"device_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the device this interface belongs to.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the interface.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The interface type (e.g., virtual, 1000base-t, 10gbase-x-sfpp).",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the interface is enabled.",
				Optional:            true,
				Computed:            true,
			},
			"mac_address": schema.StringAttribute{
				MarkdownDescription: "The MAC address of the interface.",
				Optional:            true,
			},
			"mtu": schema.Int64Attribute{
				MarkdownDescription: "The MTU of the interface.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the interface.",
				Optional:            true,
			},
		},
	}
}

func (r *deviceInterfaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deviceInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceInterfaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"device": map[string]interface{}{"id": plan.DeviceId.ValueInt64()},
		"name":   plan.Name.ValueString(),
		"type":   plan.Type.ValueString(),
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		payload["enabled"] = plan.Enabled.ValueBool()
	}
	if !plan.MacAddress.IsNull() && !plan.MacAddress.IsUnknown() {
		payload["mac_address"] = plan.MacAddress.ValueString()
	}
	if !plan.Mtu.IsNull() && !plan.Mtu.IsUnknown() {
		payload["mtu"] = plan.Mtu.ValueInt64()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	bodyStr, err := r.client.Post(ctx, "api/dcim/interfaces/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating Device Interface", err.Error())
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

	if enabled, ok := apiResponse["enabled"].(bool); ok {
		plan.Enabled = types.BoolValue(enabled)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceInterfaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/interfaces/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read device interface, assuming it was deleted", map[string]interface{}{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing read response", err.Error())
		return
	}

	if deviceMap, ok := apiResponse["device"].(map[string]interface{}); ok {
		if idFloat, ok := deviceMap["id"].(float64); ok {
			state.DeviceId = types.Int64Value(int64(idFloat))
		}
	}

	if name, ok := apiResponse["name"].(string); ok {
		state.Name = types.StringValue(name)
	}

	if typeMap, ok := apiResponse["type"].(map[string]interface{}); ok {
		if val, ok := typeMap["value"].(string); ok {
			state.Type = types.StringValue(val)
		}
	}

	if enabled, ok := apiResponse["enabled"].(bool); ok {
		state.Enabled = types.BoolValue(enabled)
	}

	if macAddr, ok := apiResponse["mac_address"].(string); ok && !state.MacAddress.IsNull() {
		state.MacAddress = types.StringValue(macAddr)
	}

	if mtu, ok := apiResponse["mtu"].(float64); ok && !state.Mtu.IsNull() {
		state.Mtu = types.Int64Value(int64(mtu))
	}

	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *deviceInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state deviceInterfaceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}
	if !plan.Name.Equal(state.Name) {
		payload["name"] = plan.Name.ValueString()
	}
	if !plan.Type.Equal(state.Type) {
		payload["type"] = plan.Type.ValueString()
	}
	if !plan.Enabled.Equal(state.Enabled) {
		payload["enabled"] = plan.Enabled.ValueBool()
	}
	if !plan.MacAddress.Equal(state.MacAddress) {
		payload["mac_address"] = plan.MacAddress.ValueString()
	}
	if !plan.Mtu.Equal(state.Mtu) {
		if plan.Mtu.IsNull() {
			payload["mtu"] = nil
		} else {
			payload["mtu"] = plan.Mtu.ValueInt64()
		}
	}
	if !plan.Description.Equal(state.Description) {
		payload["description"] = plan.Description.ValueString()
	}

	if len(payload) > 0 {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		path := fmt.Sprintf("api/dcim/interfaces/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating Device Interface", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceInterfaceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/interfaces/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
