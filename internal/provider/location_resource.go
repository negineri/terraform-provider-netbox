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

var _ resource.Resource = &locationResource{}
var _ resource.ResourceWithConfigure = &locationResource{}

func NewLocationResource() resource.Resource {
	return &locationResource{}
}

type locationResource struct {
	client *client.NetboxClient
}

type locationResourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Slug         types.String `tfsdk:"slug"`
	SiteId       types.Int64  `tfsdk:"site_id"`
	ParentId     types.Int64  `tfsdk:"parent_id"`
	Status       types.String `tfsdk:"status"`
	Description  types.String `tfsdk:"description"`
	CustomFields types.Map    `tfsdk:"custom_fields"`
}

func (r *locationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_location"
}

func (r *locationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Location in Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the location.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the location.",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-friendly unique shorthand for the location. If omitted, auto-generated from name.",
				Optional:            true,
				Computed:            true,
			},
			"site_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the site this location belongs to.",
				Required:            true,
			},
			"parent_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the parent location (for hierarchical locations).",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the location (e.g., active, planned, staging, decommissioning, retired).",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the location.",
				Optional:            true,
			},
			"custom_fields": customFieldsSchema(),
		},
	}
}

func (r *locationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *locationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan locationResourceModel
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
		"site": plan.SiteId.ValueInt64(),
	}
	if !plan.ParentId.IsNull() && !plan.ParentId.IsUnknown() {
		payload["parent"] = plan.ParentId.ValueInt64()
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}
	if cf := customFieldsToPayload(ctx, r.client, plan.CustomFields, &resp.Diagnostics); cf != nil {
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

	bodyStr, err := r.client.Post(ctx, "api/dcim/locations/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating Location", err.Error())
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

func (r *locationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state locationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/locations/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read location, assuming it was deleted", map[string]interface{}{"error": err.Error()})
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
	if siteMap, ok := apiResponse["site"].(map[string]interface{}); ok {
		if siteID, ok := siteMap["id"].(float64); ok {
			state.SiteId = types.Int64Value(int64(siteID))
		}
	}
	if parentMap, ok := apiResponse["parent"].(map[string]interface{}); ok {
		if parentID, ok := parentMap["id"].(float64); ok {
			state.ParentId = types.Int64Value(int64(parentID))
		}
	} else if !state.ParentId.IsNull() {
		state.ParentId = types.Int64Null()
	}
	if statusMap, ok := apiResponse["status"].(map[string]interface{}); ok {
		if val, ok := statusMap["value"].(string); ok && !state.Status.IsNull() {
			state.Status = types.StringValue(val)
		}
	}
	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
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

func (r *locationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state locationResourceModel
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
	if !plan.SiteId.Equal(state.SiteId) {
		payload["site"] = plan.SiteId.ValueInt64()
	}
	if !plan.ParentId.Equal(state.ParentId) {
		if plan.ParentId.IsNull() {
			payload["parent"] = nil
		} else {
			payload["parent"] = plan.ParentId.ValueInt64()
		}
	}
	if !plan.Status.Equal(state.Status) {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.Description.Equal(state.Description) {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.CustomFields.Equal(state.CustomFields) {
		if cf := customFieldsToPayload(ctx, r.client, plan.CustomFields, &resp.Diagnostics); cf != nil {
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

		path := fmt.Sprintf("api/dcim/locations/%d/", state.Id.ValueInt64())
		bodyStr, err := r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating Location", err.Error())
			return
		}

		var apiResponse map[string]interface{}
		if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
			resp.Diagnostics.AddError("Error parsing update response", err.Error())
			return
		}
		if slugVal, ok := apiResponse["slug"].(string); ok {
			plan.Slug = types.StringValue(slugVal)
		}
	} else if plan.Slug.IsNull() || plan.Slug.IsUnknown() {
		plan.Slug = state.Slug
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *locationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state locationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/locations/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
