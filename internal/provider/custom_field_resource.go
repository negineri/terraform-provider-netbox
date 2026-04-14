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

var _ resource.Resource = &customFieldResource{}
var _ resource.ResourceWithConfigure = &customFieldResource{}

func NewCustomFieldResource() resource.Resource {
	return &customFieldResource{}
}

type customFieldResource struct {
	client *client.NetboxClient
}

type customFieldResourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Label        types.String `tfsdk:"label"`
	Type         types.String `tfsdk:"type"`
	ContentTypes types.List   `tfsdk:"content_types"`
	Required     types.Bool   `tfsdk:"required"`
	Description  types.String `tfsdk:"description"`
	Default      types.String `tfsdk:"default"`
	Weight       types.Int64  `tfsdk:"weight"`
	FilterLogic  types.String `tfsdk:"filter_logic"`
	GroupName    types.String `tfsdk:"group_name"`
	Choices      types.List   `tfsdk:"choices"`
	UIVisible    types.String `tfsdk:"ui_visible"`
	UIEditable   types.String `tfsdk:"ui_editable"`
	IsCloneable  types.Bool   `tfsdk:"is_cloneable"`
	SearchWeight types.Int64  `tfsdk:"search_weight"`
}

func (r *customFieldResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_field"
}

func (r *customFieldResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Custom Field in Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the custom field.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Internal field name. Must be unique and contain only alphanumerics and underscores.",
				Required:            true,
			},
			"label": schema.StringAttribute{
				MarkdownDescription: "Display name for the field in the UI. If omitted, the name is used.",
				Optional:            true,
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the custom field. One of: text, longtext, integer, decimal, boolean, date, datetime, url, json, select, multiselect, object, multiobject.",
				Required:            true,
			},
			"content_types": schema.ListAttribute{
				MarkdownDescription: "List of content types this custom field is assigned to (e.g. \"dcim.device\").",
				Required:            true,
				ElementType:         types.StringType,
			},
			"required": schema.BoolAttribute{
				MarkdownDescription: "Whether a value is required for this field.",
				Optional:            true,
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the custom field.",
				Optional:            true,
				Computed:            true,
			},
			"default": schema.StringAttribute{
				MarkdownDescription: "Default value for the custom field (as a JSON-encoded string).",
				Optional:            true,
			},
			"weight": schema.Int64Attribute{
				MarkdownDescription: "Fields with higher weights appear lower in the form.",
				Optional:            true,
				Computed:            true,
			},
			"filter_logic": schema.StringAttribute{
				MarkdownDescription: "Filtering logic for this field. One of: disabled, loose, exact.",
				Optional:            true,
				Computed:            true,
			},
			"group_name": schema.StringAttribute{
				MarkdownDescription: "Custom fields within the same group will be displayed together.",
				Optional:            true,
				Computed:            true,
			},
			"choices": schema.ListAttribute{
				MarkdownDescription: "Comma-separated list of available choices (for select/multiselect fields).",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"ui_visible": schema.StringAttribute{
				MarkdownDescription: "Controls visibility in the UI. One of: always, if-set, hidden.",
				Optional:            true,
				Computed:            true,
			},
			"ui_editable": schema.StringAttribute{
				MarkdownDescription: "Controls editability in the UI. One of: yes, no, hidden.",
				Optional:            true,
				Computed:            true,
			},
			"is_cloneable": schema.BoolAttribute{
				MarkdownDescription: "Whether this field is copied when cloning an object.",
				Optional:            true,
				Computed:            true,
			},
			"search_weight": schema.Int64Attribute{
				MarkdownDescription: "Weighting for search. Lower values appear higher in search results.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *customFieldResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *customFieldResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan customFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"name": plan.Name.ValueString(),
		"type": plan.Type.ValueString(),
	}

	var contentTypes []string
	resp.Diagnostics.Append(plan.ContentTypes.ElementsAs(ctx, &contentTypes, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	payload["object_types"] = contentTypes

	if !plan.Label.IsNull() && !plan.Label.IsUnknown() {
		payload["label"] = plan.Label.ValueString()
	}
	if !plan.Required.IsNull() && !plan.Required.IsUnknown() {
		payload["required"] = plan.Required.ValueBool()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.Default.IsNull() && !plan.Default.IsUnknown() {
		payload["default"] = plan.Default.ValueString()
	}
	if !plan.Weight.IsNull() && !plan.Weight.IsUnknown() {
		payload["weight"] = plan.Weight.ValueInt64()
	}
	if !plan.FilterLogic.IsNull() && !plan.FilterLogic.IsUnknown() {
		payload["filter_logic"] = plan.FilterLogic.ValueString()
	}
	if !plan.GroupName.IsNull() && !plan.GroupName.IsUnknown() {
		payload["group_name"] = plan.GroupName.ValueString()
	}
	if !plan.Choices.IsNull() && !plan.Choices.IsUnknown() {
		var choices []string
		resp.Diagnostics.Append(plan.Choices.ElementsAs(ctx, &choices, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload["choices"] = choices
	}
	if !plan.UIVisible.IsNull() && !plan.UIVisible.IsUnknown() {
		payload["ui_visible"] = plan.UIVisible.ValueString()
	}
	if !plan.UIEditable.IsNull() && !plan.UIEditable.IsUnknown() {
		payload["ui_editable"] = plan.UIEditable.ValueString()
	}
	if !plan.IsCloneable.IsNull() && !plan.IsCloneable.IsUnknown() {
		payload["is_cloneable"] = plan.IsCloneable.ValueBool()
	}
	if !plan.SearchWeight.IsNull() && !plan.SearchWeight.IsUnknown() {
		payload["search_weight"] = plan.SearchWeight.ValueInt64()
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	bodyStr, err := r.client.Post(ctx, "api/extras/custom-fields/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating Custom Field", err.Error())
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

	setCustomFieldStateFromAPI(&plan, apiResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customFieldResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state customFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/extras/custom-fields/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read custom field, assuming it was deleted", map[string]interface{}{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing read response", err.Error())
		return
	}

	setCustomFieldStateFromAPI(&state, apiResponse)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *customFieldResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state customFieldResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}

	if !plan.Name.Equal(state.Name) && !plan.Name.IsUnknown() {
		payload["name"] = plan.Name.ValueString()
	}
	if !plan.Label.Equal(state.Label) && !plan.Label.IsUnknown() {
		payload["label"] = plan.Label.ValueString()
	}
	if !plan.Type.Equal(state.Type) && !plan.Type.IsUnknown() {
		payload["type"] = plan.Type.ValueString()
	}
	if !plan.ContentTypes.Equal(state.ContentTypes) && !plan.ContentTypes.IsUnknown() {
		var contentTypes []string
		resp.Diagnostics.Append(plan.ContentTypes.ElementsAs(ctx, &contentTypes, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload["object_types"] = contentTypes
	}
	if !plan.Required.Equal(state.Required) && !plan.Required.IsUnknown() {
		payload["required"] = plan.Required.ValueBool()
	}
	if !plan.Description.Equal(state.Description) && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.Default.Equal(state.Default) && !plan.Default.IsUnknown() {
		if plan.Default.IsNull() {
			payload["default"] = nil
		} else {
			payload["default"] = plan.Default.ValueString()
		}
	}
	if !plan.Weight.Equal(state.Weight) && !plan.Weight.IsUnknown() {
		payload["weight"] = plan.Weight.ValueInt64()
	}
	if !plan.FilterLogic.Equal(state.FilterLogic) && !plan.FilterLogic.IsUnknown() {
		payload["filter_logic"] = plan.FilterLogic.ValueString()
	}
	if !plan.GroupName.Equal(state.GroupName) && !plan.GroupName.IsUnknown() {
		payload["group_name"] = plan.GroupName.ValueString()
	}
	if !plan.Choices.Equal(state.Choices) && !plan.Choices.IsUnknown() {
		var choices []string
		resp.Diagnostics.Append(plan.Choices.ElementsAs(ctx, &choices, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		payload["choices"] = choices
	}
	if !plan.UIVisible.Equal(state.UIVisible) && !plan.UIVisible.IsUnknown() {
		payload["ui_visible"] = plan.UIVisible.ValueString()
	}
	if !plan.UIEditable.Equal(state.UIEditable) && !plan.UIEditable.IsUnknown() {
		payload["ui_editable"] = plan.UIEditable.ValueString()
	}
	if !plan.IsCloneable.Equal(state.IsCloneable) && !plan.IsCloneable.IsUnknown() {
		payload["is_cloneable"] = plan.IsCloneable.ValueBool()
	}
	if !plan.SearchWeight.Equal(state.SearchWeight) && !plan.SearchWeight.IsUnknown() {
		payload["search_weight"] = plan.SearchWeight.ValueInt64()
	}

	if len(payload) > 0 {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		apiPath := fmt.Sprintf("api/extras/custom-fields/%d/", state.Id.ValueInt64())
		patchBody, err := r.client.Patch(ctx, apiPath, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating Custom Field", err.Error())
			return
		}

		// PATCH レスポンスから Computed フィールドを更新する
		var patchResponse map[string]interface{}
		if err := json.Unmarshal([]byte(*patchBody), &patchResponse); err != nil {
			resp.Diagnostics.AddError("Error parsing update response", err.Error())
			return
		}
		setCustomFieldStateFromAPI(&plan, patchResponse)
	} else {
		// 変更なしの場合は state の Computed フィールドを保持する
		plan.Label = state.Label
		plan.Required = state.Required
		plan.Description = state.Description
		plan.Weight = state.Weight
		plan.FilterLogic = state.FilterLogic
		plan.GroupName = state.GroupName
		plan.Choices = state.Choices
		plan.UIVisible = state.UIVisible
		plan.UIEditable = state.UIEditable
		plan.IsCloneable = state.IsCloneable
		plan.SearchWeight = state.SearchWeight
		plan.ContentTypes = state.ContentTypes
	}

	plan.Id = state.Id

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *customFieldResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state customFieldResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/extras/custom-fields/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}

// setCustomFieldStateFromAPI は API レスポンスから customFieldResourceModel にフィールドを設定します。
func setCustomFieldStateFromAPI(state *customFieldResourceModel, api map[string]interface{}) {
	if name, ok := api["name"].(string); ok {
		state.Name = types.StringValue(name)
	}
	if label, ok := api["label"].(string); ok {
		state.Label = types.StringValue(label)
	}
	if t, ok := api["type"].(map[string]interface{}); ok {
		if v, ok := t["value"].(string); ok {
			state.Type = types.StringValue(v)
		}
	}
	ctRaw, _ := api["object_types"].([]interface{})
	{
		elems := make([]attr.Value, 0, len(ctRaw))
		for _, v := range ctRaw {
			if s, ok := v.(string); ok {
				elems = append(elems, types.StringValue(s))
			}
		}
		state.ContentTypes = types.ListValueMust(types.StringType, elems)
	}
	if req, ok := api["required"].(bool); ok {
		state.Required = types.BoolValue(req)
	}
	if desc, ok := api["description"].(string); ok {
		state.Description = types.StringValue(desc)
	}
	if def, ok := api["default"]; ok && def != nil {
		switch v := def.(type) {
		case string:
			state.Default = types.StringValue(v)
		case float64:
			state.Default = types.StringValue(fmt.Sprintf("%v", v))
		case bool:
			if v {
				state.Default = types.StringValue("true")
			} else {
				state.Default = types.StringValue("false")
			}
		}
	}
	if weight, ok := api["weight"].(float64); ok {
		state.Weight = types.Int64Value(int64(weight))
	}
	if fl, ok := api["filter_logic"].(map[string]interface{}); ok {
		if v, ok := fl["value"].(string); ok {
			state.FilterLogic = types.StringValue(v)
		}
	}
	if gn, ok := api["group_name"].(string); ok {
		state.GroupName = types.StringValue(gn)
	}
	if choicesRaw, ok := api["choices"].([]interface{}); ok {
		elems := make([]attr.Value, 0, len(choicesRaw))
		for _, v := range choicesRaw {
			if s, ok := v.(string); ok {
				elems = append(elems, types.StringValue(s))
			}
		}
		state.Choices = types.ListValueMust(types.StringType, elems)
	} else if !state.Choices.IsNull() {
		state.Choices = types.ListValueMust(types.StringType, []attr.Value{})
	}
	if uiv, ok := api["ui_visible"].(map[string]interface{}); ok {
		if v, ok := uiv["value"].(string); ok {
			state.UIVisible = types.StringValue(v)
		}
	}
	if uie, ok := api["ui_editable"].(map[string]interface{}); ok {
		if v, ok := uie["value"].(string); ok {
			state.UIEditable = types.StringValue(v)
		}
	}
	if ic, ok := api["is_cloneable"].(bool); ok {
		state.IsCloneable = types.BoolValue(ic)
	}
	if sw, ok := api["search_weight"].(float64); ok {
		state.SearchWeight = types.Int64Value(int64(sw))
	}
}
