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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &macAddressResource{}
var _ resource.ResourceWithConfigure = &macAddressResource{}

func NewMacAddressResource() resource.Resource {
	return &macAddressResource{}
}

type macAddressResource struct {
	client *client.NetboxClient
}

type macAddressResourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	MacAddress  types.String `tfsdk:"mac_address"`
	Description types.String `tfsdk:"description"`
	Comments    types.String `tfsdk:"comments"`
}

func (r *macAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mac_address"
}

func (r *macAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a MAC address within Netbox. Interface assignment is managed via primary_mac_address_id on the interface resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the MAC address.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"mac_address": schema.StringAttribute{
				MarkdownDescription: "The MAC address (e.g., AA:BB:CC:DD:EE:FF).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the MAC address.",
				Optional:            true,
			},
			"comments": schema.StringAttribute{
				MarkdownDescription: "Comments for the MAC address.",
				Optional:            true,
			},
		},
	}
}

func (r *macAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *macAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan macAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"mac_address": plan.MacAddress.ValueString(),
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.Comments.IsNull() && !plan.Comments.IsUnknown() {
		payload["comments"] = plan.Comments.ValueString()
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	bodyStr, err := r.client.Post(ctx, "api/dcim/mac-addresses/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating MAC address", err.Error())
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

	if macAddr, ok := apiResponse["mac_address"].(string); ok {
		plan.MacAddress = types.StringValue(macAddr)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *macAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state macAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/mac-addresses/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read MAC address, assuming it was deleted", map[string]interface{}{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing read response", err.Error())
		return
	}

	if macAddr, ok := apiResponse["mac_address"].(string); ok {
		state.MacAddress = types.StringValue(macAddr)
	}
	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
	}
	if comments, ok := apiResponse["comments"].(string); ok && !state.Comments.IsNull() {
		state.Comments = types.StringValue(comments)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *macAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state macAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}
	if !plan.Description.Equal(state.Description) {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.Comments.Equal(state.Comments) {
		payload["comments"] = plan.Comments.ValueString()
	}

	if len(payload) > 0 {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		path := fmt.Sprintf("api/dcim/mac-addresses/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating MAC address", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *macAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state macAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/mac-addresses/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
