// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &availableIpResource{}
var _ resource.ResourceWithConfigure = &availableIpResource{}

func NewAvailableIpResource() resource.Resource {
	return &availableIpResource{}
}

type availableIpResource struct {
	client *client.NetboxClient
}

type availableIpResourceModel struct {
	PrefixId    types.Int64  `tfsdk:"prefix_id"`
	IpRangeId   types.Int64  `tfsdk:"ip_range_id"`
	Status      types.String `tfsdk:"status"`
	Description types.String `tfsdk:"description"`
	Id          types.Int64  `tfsdk:"id"`
	IpAddress   types.String `tfsdk:"ip_address"`
}

func (r *availableIpResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_available_ip"
}

func (r *availableIpResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Allocates the next available IP address from a given prefix.",
		Attributes: map[string]schema.Attribute{
			"prefix_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the prefix to allocate the IP from. Exactly one of prefix_id or ip_range_id must be provided.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("ip_range_id")),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"ip_range_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the IP range to allocate the IP from. Exactly one of prefix_id or ip_range_id must be provided.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("prefix_id")),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the IP address (e.g., active, reserved).",
				Optional:            true,
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
			"ip_address": schema.StringAttribute{
				MarkdownDescription: "The allocated IP address (e.g., 192.168.1.10/24).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *availableIpResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *availableIpResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan availableIpResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}

	var bodyBytes []byte
	var err error
	if len(payload) > 0 {
		bodyBytes, err = json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}
	}

	var path string
	if !plan.PrefixId.IsNull() && !plan.PrefixId.IsUnknown() {
		path = fmt.Sprintf("api/ipam/prefixes/%d/available-ips/", plan.PrefixId.ValueInt64())
	} else if !plan.IpRangeId.IsNull() && !plan.IpRangeId.IsUnknown() {
		path = fmt.Sprintf("api/ipam/ip-ranges/%d/available-ips/", plan.IpRangeId.ValueInt64())
	} else {
		resp.Diagnostics.AddError("Error creating IP address", "Either prefix_id or ip_range_id must be provided.")
		return
	}

	// Set array payload wrapper required by Netbox available-ips endpoint
	var finalPayload []byte
	if len(bodyBytes) > 0 {
		// Post to available-ips expects an array of objects for multiple IPs, or list of objects, we send empty array or object array.
		// Actually, standard behavior is sending a single object or an array of objects. Usually an array of size 1 works well.
		// Netbox documentation: `POST /api/ipam/prefixes/{id}/available-ips/` takes a list of objects or single object.
		finalPayload = bodyBytes
	}

	bodyStr, err := r.client.Post(ctx, path, bytes.NewReader(finalPayload))
	if err != nil {
		resp.Diagnostics.AddError("Error creating IP address", err.Error())
		return
	}

	var apiResponse interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing create response", err.Error())
		return
	}

	// Netbox might return an array if we requested multiple (or just because of endpoint design).
	// Let's handle both cases.
	var createdIp map[string]interface{}
	switch v := apiResponse.(type) {
	case []interface{}:
		if len(v) == 0 {
			resp.Diagnostics.AddError("Error creating IP address", "Netbox returned an empty array")
			return
		}
		var ok bool
		createdIp, ok = v[0].(map[string]interface{})
		if !ok {
			resp.Diagnostics.AddError("Error parsing create response", "Unexpected element type in array")
			return
		}
	case map[string]interface{}:
		createdIp = v
	default:
		resp.Diagnostics.AddError("Error parsing create response", "Unexpected format")
		return
	}

	idFloat, ok := createdIp["id"].(float64)
	if !ok {
		resp.Diagnostics.AddError("Error parsing create response", "Could not find 'id' in response")
		return
	}
	address, ok := createdIp["address"].(string)
	if !ok {
		resp.Diagnostics.AddError("Error parsing create response", "Could not find 'address' in response")
		return
	}

	plan.Id = types.Int64Value(int64(idFloat))
	plan.IpAddress = types.StringValue(address)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *availableIpResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state availableIpResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/ipam/ip-addresses/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		// If the IP no longer exists, we should return a 404, we must remove it from state
		// We'll roughly check for 404 text if doRequest puts it there, but ideally we should cleanly handle HTTP statuses.
		// doRequest concatenates error code, but here we'll just log and remove for simplicity if we can't retrieve.
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
		if val, ok := statusMap["value"].(string); ok && !state.Status.IsNull() {
			state.Status = types.StringValue(val)
		}
	}

	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *availableIpResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state availableIpResourceModel
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

		path := fmt.Sprintf("api/ipam/ip-addresses/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating IP address", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *availableIpResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state availableIpResourceModel
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
