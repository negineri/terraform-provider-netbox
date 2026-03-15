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

var _ resource.Resource = &vlanGroupResource{}
var _ resource.ResourceWithConfigure = &vlanGroupResource{}

func NewVlanGroupResource() resource.Resource {
	return &vlanGroupResource{}
}

type vlanGroupResource struct {
	client *client.NetboxClient
}

type vlanGroupResourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Description types.String `tfsdk:"description"`
	MinVid      types.Int64  `tfsdk:"min_vid"`
	MaxVid      types.Int64  `tfsdk:"max_vid"`
	ScopeType   types.String `tfsdk:"scope_type"`
	ScopeId     types.Int64  `tfsdk:"scope_id"`
}

func (r *vlanGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vlan_group"
}

func (r *vlanGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VLAN Group in Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the VLAN group.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the VLAN group.",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-friendly unique shorthand for the VLAN group. If omitted, auto-generated from name.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the VLAN group.",
				Optional:            true,
			},
			"min_vid": schema.Int64Attribute{
				MarkdownDescription: "Minimum VLAN ID (1-4094).",
				Optional:            true,
			},
			"max_vid": schema.Int64Attribute{
				MarkdownDescription: "Maximum VLAN ID (1-4094).",
				Optional:            true,
			},
			"scope_type": schema.StringAttribute{
				MarkdownDescription: "The type of the scope object (e.g., dcim.site, dcim.location, dcim.region, dcim.sitegroup, virtualization.cluster, virtualization.clustergroup).",
				Optional:            true,
			},
			"scope_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the scope object.",
				Optional:            true,
			},
		},
	}
}

func (r *vlanGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *vlanGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vlanGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Auto-generate slug from name if not provided
	slug := plan.Slug.ValueString()
	if plan.Slug.IsNull() || plan.Slug.IsUnknown() || slug == "" {
		slug = slugify(plan.Name.ValueString())
	}

	payload := map[string]interface{}{
		"name": plan.Name.ValueString(),
		"slug": slug,
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.MinVid.IsNull() && !plan.MinVid.IsUnknown() {
		payload["min_vid"] = plan.MinVid.ValueInt64()
	}
	if !plan.MaxVid.IsNull() && !plan.MaxVid.IsUnknown() {
		payload["max_vid"] = plan.MaxVid.ValueInt64()
	}
	if !plan.ScopeType.IsNull() && !plan.ScopeType.IsUnknown() {
		payload["scope_type"] = plan.ScopeType.ValueString()
	}
	if !plan.ScopeId.IsNull() && !plan.ScopeId.IsUnknown() {
		payload["scope_id"] = plan.ScopeId.ValueInt64()
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	bodyStr, err := r.client.Post(ctx, "api/ipam/vlan-groups/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating VLAN Group", err.Error())
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

	if slugVal, ok := apiResponse["slug"].(string); ok {
		plan.Slug = types.StringValue(slugVal)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vlanGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vlanGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/ipam/vlan-groups/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read VLAN group, assuming it was deleted", map[string]interface{}{"error": err.Error()})
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
	if slugVal, ok := apiResponse["slug"].(string); ok {
		state.Slug = types.StringValue(slugVal)
	}
	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
	}
	if minVid, ok := apiResponse["min_vid"].(float64); ok && !state.MinVid.IsNull() {
		state.MinVid = types.Int64Value(int64(minVid))
	}
	if maxVid, ok := apiResponse["max_vid"].(float64); ok && !state.MaxVid.IsNull() {
		state.MaxVid = types.Int64Value(int64(maxVid))
	}
	if scopeType, ok := apiResponse["scope_type"].(string); ok && !state.ScopeType.IsNull() {
		state.ScopeType = types.StringValue(scopeType)
	}
	if scopeId, ok := apiResponse["scope_id"].(float64); ok && !state.ScopeId.IsNull() {
		state.ScopeId = types.Int64Value(int64(scopeId))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *vlanGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state vlanGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}
	if !plan.Name.Equal(state.Name) {
		payload["name"] = plan.Name.ValueString()
		// Re-generate slug if name changed and slug was auto-generated
		if plan.Slug.IsNull() || plan.Slug.IsUnknown() {
			payload["slug"] = slugify(plan.Name.ValueString())
		}
	}
	if !plan.Slug.Equal(state.Slug) && !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		payload["slug"] = plan.Slug.ValueString()
	}
	if !plan.Description.Equal(state.Description) {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.MinVid.Equal(state.MinVid) {
		payload["min_vid"] = plan.MinVid.ValueInt64()
	}
	if !plan.MaxVid.Equal(state.MaxVid) {
		payload["max_vid"] = plan.MaxVid.ValueInt64()
	}
	if !plan.ScopeType.Equal(state.ScopeType) {
		payload["scope_type"] = plan.ScopeType.ValueString()
	}
	if !plan.ScopeId.Equal(state.ScopeId) {
		payload["scope_id"] = plan.ScopeId.ValueInt64()
	}

	if len(payload) > 0 {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		path := fmt.Sprintf("api/ipam/vlan-groups/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating VLAN Group", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *vlanGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vlanGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/ipam/vlan-groups/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
