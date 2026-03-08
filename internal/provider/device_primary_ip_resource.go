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

var _ resource.Resource = &devicePrimaryIPResource{}
var _ resource.ResourceWithConfigure = &devicePrimaryIPResource{}

func NewDevicePrimaryIPResource() resource.Resource {
	return &devicePrimaryIPResource{}
}

type devicePrimaryIPResource struct {
	client *client.NetboxClient
}

type devicePrimaryIPResourceModel struct {
	Id            types.Int64 `tfsdk:"id"`
	DeviceId      types.Int64 `tfsdk:"device_id"`
	PrimaryIPv4Id types.Int64 `tfsdk:"primary_ipv4_id"`
	PrimaryIPv6Id types.Int64 `tfsdk:"primary_ipv6_id"`
}

func (r *devicePrimaryIPResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_primary_ip"
}

func (r *devicePrimaryIPResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages the primary IPv4/IPv6 addresses of a Device in Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the device (same as device_id).",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"device_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the device to configure.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"primary_ipv4_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the IP address to set as the primary IPv4 address.",
				Optional:            true,
			},
			"primary_ipv6_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the IP address to set as the primary IPv6 address.",
				Optional:            true,
			},
		},
	}
}

func (r *devicePrimaryIPResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *devicePrimaryIPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan devicePrimaryIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := buildPrimaryIPPayload(plan)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	path := fmt.Sprintf("api/dcim/devices/%d/", plan.DeviceId.ValueInt64())
	_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error setting primary IPs on Device", err.Error())
		return
	}

	plan.Id = plan.DeviceId

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *devicePrimaryIPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state devicePrimaryIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/devices/%d/", state.DeviceId.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read device, assuming it was deleted", map[string]interface{}{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing read response", err.Error())
		return
	}

	if ipv4Map, ok := apiResponse["primary_ip4"].(map[string]interface{}); ok {
		if idFloat, ok := ipv4Map["id"].(float64); ok {
			state.PrimaryIPv4Id = types.Int64Value(int64(idFloat))
		}
	} else {
		state.PrimaryIPv4Id = types.Int64Null()
	}

	if ipv6Map, ok := apiResponse["primary_ip6"].(map[string]interface{}); ok {
		if idFloat, ok := ipv6Map["id"].(float64); ok {
			state.PrimaryIPv6Id = types.Int64Value(int64(idFloat))
		}
	} else {
		state.PrimaryIPv6Id = types.Int64Null()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *devicePrimaryIPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan devicePrimaryIPResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := buildPrimaryIPPayload(plan)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	path := fmt.Sprintf("api/dcim/devices/%d/", plan.DeviceId.ValueInt64())
	_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error updating primary IPs on Device", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *devicePrimaryIPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state devicePrimaryIPResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"primary_ip4": nil,
		"primary_ip6": nil,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	path := fmt.Sprintf("api/dcim/devices/%d/", state.DeviceId.ValueInt64())
	_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
	if err != nil {
		tflog.Warn(ctx, "Failed to clear primary IPs on device, assuming already cleared", map[string]interface{}{"error": err.Error()})
	}
}

func buildPrimaryIPPayload(plan devicePrimaryIPResourceModel) map[string]interface{} {
	payload := map[string]interface{}{}

	if !plan.PrimaryIPv4Id.IsNull() && !plan.PrimaryIPv4Id.IsUnknown() {
		payload["primary_ip4"] = map[string]interface{}{"id": plan.PrimaryIPv4Id.ValueInt64()}
	} else {
		payload["primary_ip4"] = nil
	}

	if !plan.PrimaryIPv6Id.IsNull() && !plan.PrimaryIPv6Id.IsUnknown() {
		payload["primary_ip6"] = map[string]interface{}{"id": plan.PrimaryIPv6Id.ValueInt64()}
	} else {
		payload["primary_ip6"] = nil
	}

	return payload
}
