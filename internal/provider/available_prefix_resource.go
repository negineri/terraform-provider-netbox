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

var _ resource.Resource = &availablePrefixResource{}
var _ resource.ResourceWithConfigure = &availablePrefixResource{}

func NewAvailablePrefixResource() resource.Resource {
	return &availablePrefixResource{}
}

type availablePrefixResource struct {
	client *client.NetboxClient
}

type availablePrefixResourceModel struct {
	ParentPrefixId types.Int64  `tfsdk:"parent_prefix_id"`
	PrefixLength   types.Int64  `tfsdk:"prefix_length"`
	Status         types.String `tfsdk:"status"`
	Description    types.String `tfsdk:"description"`
	Id             types.Int64  `tfsdk:"id"`
	Prefix         types.String `tfsdk:"prefix"`
}

func (r *availablePrefixResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_available_prefix"
}

func (r *availablePrefixResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Allocates the next available prefix of the given prefix length from a parent prefix.",
		Attributes: map[string]schema.Attribute{
			"parent_prefix_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the parent prefix to allocate the new prefix from.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"prefix_length": schema.Int64Attribute{
				MarkdownDescription: "The prefix length (e.g. 24 for a /24) of the new prefix to allocate.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the prefix (e.g., active, reserved).",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the prefix.",
				Optional:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the allocated prefix.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"prefix": schema.StringAttribute{
				MarkdownDescription: "The allocated prefix in CIDR notation (e.g., 192.168.1.0/24).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *availablePrefixResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *availablePrefixResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan availablePrefixResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"prefix_length": plan.PrefixLength.ValueInt64(),
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	path := fmt.Sprintf("api/ipam/prefixes/%d/available-prefixes/", plan.ParentPrefixId.ValueInt64())
	bodyStr, err := r.client.Post(ctx, path, bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating available prefix", err.Error())
		return
	}

	var apiResponse interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing create response", err.Error())
		return
	}

	var createdPrefix map[string]interface{}
	switch v := apiResponse.(type) {
	case []interface{}:
		if len(v) == 0 {
			resp.Diagnostics.AddError("Error creating available prefix", "Netbox returned an empty array")
			return
		}
		var ok bool
		createdPrefix, ok = v[0].(map[string]interface{})
		if !ok {
			resp.Diagnostics.AddError("Error parsing create response", "Unexpected element type in array")
			return
		}
	case map[string]interface{}:
		createdPrefix = v
	default:
		resp.Diagnostics.AddError("Error parsing create response", "Unexpected response format")
		return
	}

	idFloat, ok := createdPrefix["id"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Error parsing create response", "Could not find 'id' in response")
		return
	}
	prefixVal, ok := createdPrefix["prefix"].(string)
	if !ok {
		resp.Diagnostics.AddError("Error parsing create response", "Could not find 'prefix' in response")
		return
	}

	plan.Id = types.Int64Value(int64(idFloat))
	plan.Prefix = types.StringValue(prefixVal)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *availablePrefixResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state availablePrefixResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/ipam/prefixes/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read prefix, assuming it was deleted", map[string]interface{}{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing read response", err.Error())
		return
	}

	if prefixVal, ok := apiResponse["prefix"].(string); ok {
		state.Prefix = types.StringValue(prefixVal)
	}

	if statusMap, ok := apiResponse["status"].(map[string]interface{}); ok {
		if val, ok := statusMap["value"].(string); ok && !state.Status.IsNull() {
			state.Status = types.StringValue(val)
		}
	}

	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *availablePrefixResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state availablePrefixResourceModel
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

	if len(payload) > 0 {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		path := fmt.Sprintf("api/ipam/prefixes/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating prefix", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *availablePrefixResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state availablePrefixResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/ipam/prefixes/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
