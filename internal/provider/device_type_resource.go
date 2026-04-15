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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &deviceTypeResource{}
var _ resource.ResourceWithConfigure = &deviceTypeResource{}

func NewDeviceTypeResource() resource.Resource {
	return &deviceTypeResource{}
}

type deviceTypeResource struct {
	client *client.NetboxClient
}

type deviceTypeResourceModel struct {
	Id             types.Int64   `tfsdk:"id"`
	ManufacturerId types.Int64   `tfsdk:"manufacturer_id"`
	Model          types.String  `tfsdk:"model"`
	Slug           types.String  `tfsdk:"slug"`
	PartNumber     types.String  `tfsdk:"part_number"`
	UHeight        types.Float64 `tfsdk:"u_height"`
	IsFullDepth    types.Bool    `tfsdk:"is_full_depth"`
	Description    types.String  `tfsdk:"description"`
	Tags           types.List    `tfsdk:"tags"`
	CustomFields   types.Map     `tfsdk:"custom_fields"`
}

func (r *deviceTypeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_type"
}

func (r *deviceTypeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Device Type in Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the device type.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"manufacturer_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the manufacturer.",
				Required:            true,
			},
			"model": schema.StringAttribute{
				MarkdownDescription: "The model name of the device type.",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-friendly unique shorthand for the device type. If omitted, auto-generated from model.",
				Optional:            true,
				Computed:            true,
			},
			"part_number": schema.StringAttribute{
				MarkdownDescription: "Discrete part number for the device type.",
				Optional:            true,
			},
			"u_height": schema.Float64Attribute{
				MarkdownDescription: "Device height in rack units (supports fractional values like 0.5).",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"is_full_depth": schema.BoolAttribute{
				MarkdownDescription: "Whether the device type occupies the full depth of a rack.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the device type.",
				Optional:            true,
			},
			"tags": schema.ListAttribute{
				ElementType:         types.Int64Type,
				MarkdownDescription: "List of tag IDs to assign to the device type.",
				Optional:            true,
			},
			"custom_fields": customFieldsSchema(),
		},
	}
}

func (r *deviceTypeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deviceTypeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := plan.Slug.ValueString()
	if plan.Slug.IsNull() || plan.Slug.IsUnknown() || slug == "" {
		slug = slugify(plan.Model.ValueString())
	}

	payload := map[string]interface{}{
		"manufacturer": map[string]interface{}{"id": plan.ManufacturerId.ValueInt64()},
		"model":        plan.Model.ValueString(),
		"slug":         slug,
	}
	if !plan.PartNumber.IsNull() && !plan.PartNumber.IsUnknown() {
		payload["part_number"] = plan.PartNumber.ValueString()
	}
	if !plan.UHeight.IsNull() && !plan.UHeight.IsUnknown() {
		payload["u_height"] = plan.UHeight.ValueFloat64()
	}
	if !plan.IsFullDepth.IsNull() && !plan.IsFullDepth.IsUnknown() {
		payload["is_full_depth"] = plan.IsFullDepth.ValueBool()
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

	bodyStr, err := r.client.Post(ctx, "api/dcim/device-types/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating Device Type", err.Error())
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
	if uHeight, ok := apiResponse["u_height"].(float64); ok {
		plan.UHeight = types.Float64Value(uHeight)
	}
	if isFullDepth, ok := apiResponse["is_full_depth"].(bool); ok {
		plan.IsFullDepth = types.BoolValue(isFullDepth)
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

func (r *deviceTypeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/device-types/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read device type, assuming it was deleted", map[string]interface{}{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing read response", err.Error())
		return
	}

	if manufacturerMap, ok := apiResponse["manufacturer"].(map[string]interface{}); ok {
		if idFloat, ok := manufacturerMap["id"].(float64); ok {
			state.ManufacturerId = types.Int64Value(int64(idFloat))
		}
	}
	if model, ok := apiResponse["model"].(string); ok {
		state.Model = types.StringValue(model)
	}
	if slugVal, ok := apiResponse["slug"].(string); ok {
		state.Slug = types.StringValue(slugVal)
	}
	if partNumber, ok := apiResponse["part_number"].(string); ok && !state.PartNumber.IsNull() {
		state.PartNumber = types.StringValue(partNumber)
	}
	if uHeight, ok := apiResponse["u_height"].(float64); ok {
		state.UHeight = types.Float64Value(uHeight)
	}
	if isFullDepth, ok := apiResponse["is_full_depth"].(bool); ok {
		state.IsFullDepth = types.BoolValue(isFullDepth)
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

func (r *deviceTypeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state deviceTypeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}
	if !plan.ManufacturerId.Equal(state.ManufacturerId) {
		payload["manufacturer"] = map[string]interface{}{"id": plan.ManufacturerId.ValueInt64()}
	}
	if !plan.Model.Equal(state.Model) {
		payload["model"] = plan.Model.ValueString()
		if plan.Slug.IsNull() || plan.Slug.IsUnknown() {
			payload["slug"] = slugify(plan.Model.ValueString())
		}
	}
	if !plan.Slug.Equal(state.Slug) && !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		payload["slug"] = plan.Slug.ValueString()
	}
	if !plan.PartNumber.Equal(state.PartNumber) {
		payload["part_number"] = plan.PartNumber.ValueString()
	}
	if !plan.UHeight.Equal(state.UHeight) {
		payload["u_height"] = plan.UHeight.ValueFloat64()
	}
	if !plan.IsFullDepth.Equal(state.IsFullDepth) {
		payload["is_full_depth"] = plan.IsFullDepth.ValueBool()
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

		path := fmt.Sprintf("api/dcim/device-types/%d/", state.Id.ValueInt64())
		bodyStr, err := r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating Device Type", err.Error())
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
		if uHeight, ok := apiResponse["u_height"].(float64); ok {
			plan.UHeight = types.Float64Value(uHeight)
		}
		if isFullDepth, ok := apiResponse["is_full_depth"].(bool); ok {
			plan.IsFullDepth = types.BoolValue(isFullDepth)
		}
	} else {
		if plan.Slug.IsNull() || plan.Slug.IsUnknown() {
			plan.Slug = state.Slug
		}
		if plan.UHeight.IsNull() || plan.UHeight.IsUnknown() {
			plan.UHeight = state.UHeight
		}
		if plan.IsFullDepth.IsNull() || plan.IsFullDepth.IsUnknown() {
			plan.IsFullDepth = state.IsFullDepth
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceTypeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceTypeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/device-types/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
