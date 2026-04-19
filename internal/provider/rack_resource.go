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

var _ resource.Resource = &rackResource{}
var _ resource.ResourceWithConfigure = &rackResource{}

func NewRackResource() resource.Resource {
	return &rackResource{}
}

type rackResource struct {
	client *client.NetboxClient
}

type rackResourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	SiteId       types.Int64  `tfsdk:"site_id"`
	LocationId   types.Int64  `tfsdk:"location_id"`
	RoleId       types.Int64  `tfsdk:"role_id"`
	Status       types.String `tfsdk:"status"`
	FacilityId   types.String `tfsdk:"facility_id"`
	UHeight      types.Int64  `tfsdk:"u_height"`
	Description  types.String `tfsdk:"description"`
	Tags         types.Set    `tfsdk:"tags"`
	CustomFields types.Map    `tfsdk:"custom_fields"`
}

func (r *rackResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rack"
}

func (r *rackResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Rack in Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the rack.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the rack.",
				Required:            true,
			},
			"site_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the site where the rack is located.",
				Required:            true,
			},
			"location_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the location within the site.",
				Optional:            true,
			},
			"role_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the rack role.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the rack (e.g., active, planned, staged, failed, decommissioning, retired).",
				Optional:            true,
			},
			"facility_id": schema.StringAttribute{
				MarkdownDescription: "A field used to identify the rack by a facility-specific identifier.",
				Optional:            true,
			},
			"u_height": schema.Int64Attribute{
				MarkdownDescription: "Height in rack units.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the rack.",
				Optional:            true,
			},
			"tags": schema.SetAttribute{
				ElementType:         types.Int64Type,
				MarkdownDescription: "Set of tag IDs to assign to the rack.",
				Optional:            true,
			},
			"custom_fields": customFieldsSchema(),
		},
	}
}

func (r *rackResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *rackResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan rackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"name": plan.Name.ValueString(),
		"site": map[string]interface{}{"id": plan.SiteId.ValueInt64()},
	}
	if !plan.LocationId.IsNull() && !plan.LocationId.IsUnknown() {
		payload["location"] = map[string]interface{}{"id": plan.LocationId.ValueInt64()}
	}
	if !plan.RoleId.IsNull() && !plan.RoleId.IsUnknown() {
		payload["role"] = map[string]interface{}{"id": plan.RoleId.ValueInt64()}
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.FacilityId.IsNull() && !plan.FacilityId.IsUnknown() {
		payload["facility_id"] = plan.FacilityId.ValueString()
	}
	if !plan.UHeight.IsNull() && !plan.UHeight.IsUnknown() {
		payload["u_height"] = plan.UHeight.ValueInt64()
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

	bodyStr, err := r.client.Post(ctx, "api/dcim/racks/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating Rack", err.Error())
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

func (r *rackResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state rackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/racks/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read rack, assuming it was deleted", map[string]interface{}{"error": err.Error()})
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
	if siteMap, ok := apiResponse["site"].(map[string]interface{}); ok {
		if idFloat, ok := siteMap["id"].(float64); ok {
			state.SiteId = types.Int64Value(int64(idFloat))
		}
	}
	if locationMap, ok := apiResponse["location"].(map[string]interface{}); ok {
		if idFloat, ok := locationMap["id"].(float64); ok {
			state.LocationId = types.Int64Value(int64(idFloat))
		}
	} else if !state.LocationId.IsNull() {
		state.LocationId = types.Int64Null()
	}
	if roleMap, ok := apiResponse["role"].(map[string]interface{}); ok {
		if idFloat, ok := roleMap["id"].(float64); ok {
			state.RoleId = types.Int64Value(int64(idFloat))
		}
	} else if !state.RoleId.IsNull() {
		state.RoleId = types.Int64Null()
	}
	if statusMap, ok := apiResponse["status"].(map[string]interface{}); ok {
		if val, ok := statusMap["value"].(string); ok && !state.Status.IsNull() {
			state.Status = types.StringValue(val)
		}
	}
	if facilityId, ok := apiResponse["facility_id"].(string); ok && !state.FacilityId.IsNull() {
		state.FacilityId = types.StringValue(facilityId)
	}
	if uHeight, ok := apiResponse["u_height"].(float64); ok && !state.UHeight.IsNull() {
		state.UHeight = types.Int64Value(int64(uHeight))
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
			setVal, diags := types.SetValue(types.Int64Type, tagVals)
			resp.Diagnostics.Append(diags...)
			if !resp.Diagnostics.HasError() {
				state.Tags = setVal
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

func (r *rackResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state rackResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}
	if !plan.Name.Equal(state.Name) {
		payload["name"] = plan.Name.ValueString()
	}
	if !plan.SiteId.Equal(state.SiteId) {
		payload["site"] = map[string]interface{}{"id": plan.SiteId.ValueInt64()}
	}
	if !plan.LocationId.Equal(state.LocationId) {
		if plan.LocationId.IsNull() {
			payload["location"] = nil
		} else {
			payload["location"] = map[string]interface{}{"id": plan.LocationId.ValueInt64()}
		}
	}
	if !plan.RoleId.Equal(state.RoleId) {
		if plan.RoleId.IsNull() {
			payload["role"] = nil
		} else {
			payload["role"] = map[string]interface{}{"id": plan.RoleId.ValueInt64()}
		}
	}
	if !plan.Status.Equal(state.Status) {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.FacilityId.Equal(state.FacilityId) {
		payload["facility_id"] = plan.FacilityId.ValueString()
	}
	if !plan.UHeight.Equal(state.UHeight) {
		payload["u_height"] = plan.UHeight.ValueInt64()
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

		path := fmt.Sprintf("api/dcim/racks/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating Rack", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *rackResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state rackResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/racks/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
