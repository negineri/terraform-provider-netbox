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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &deviceRoleResource{}
var _ resource.ResourceWithConfigure = &deviceRoleResource{}

func NewDeviceRoleResource() resource.Resource {
	return &deviceRoleResource{}
}

type deviceRoleResource struct {
	client *client.NetboxClient
}

type deviceRoleResourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Slug         types.String `tfsdk:"slug"`
	Color        types.String `tfsdk:"color"`
	VmRole       types.Bool   `tfsdk:"vm_role"`
	Description  types.String `tfsdk:"description"`
	CustomFields types.Map    `tfsdk:"custom_fields"`
}

func (r *deviceRoleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device_role"
}

func (r *deviceRoleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Device Role in Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the device role.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the device role.",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-friendly unique shorthand for the device role. If omitted, auto-generated from name.",
				Optional:            true,
				Computed:            true,
			},
			"color": schema.StringAttribute{
				MarkdownDescription: "Color for the device role as a 6-digit hex string (e.g. \"aa1409\").",
				Required:            true,
			},
			"vm_role": schema.BoolAttribute{
				MarkdownDescription: "Whether this role is used for virtual machines.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the device role.",
				Optional:            true,
			},
			"custom_fields": customFieldsSchema(),
		},
	}
}

func (r *deviceRoleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deviceRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := plan.Slug.ValueString()
	if plan.Slug.IsNull() || plan.Slug.IsUnknown() || slug == "" {
		slug = slugify(plan.Name.ValueString())
	}

	payload := map[string]interface{}{
		"name":  plan.Name.ValueString(),
		"slug":  slug,
		"color": plan.Color.ValueString(),
	}
	if !plan.VmRole.IsNull() && !plan.VmRole.IsUnknown() {
		payload["vm_role"] = plan.VmRole.ValueBool()
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

	bodyStr, err := r.client.Post(ctx, "api/dcim/device-roles/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating Device Role", err.Error())
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
	if vmRole, ok := apiResponse["vm_role"].(bool); ok {
		plan.VmRole = types.BoolValue(vmRole)
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

func (r *deviceRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/device-roles/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read device role, assuming it was deleted", map[string]interface{}{"error": err.Error()})
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
	if color, ok := apiResponse["color"].(string); ok {
		state.Color = types.StringValue(color)
	}
	if vmRole, ok := apiResponse["vm_role"].(bool); ok {
		state.VmRole = types.BoolValue(vmRole)
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

func (r *deviceRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state deviceRoleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}
	if !plan.Name.Equal(state.Name) {
		payload["name"] = plan.Name.ValueString()
		if plan.Slug.IsNull() || plan.Slug.IsUnknown() {
			payload["slug"] = slugify(plan.Name.ValueString())
		}
	}
	if !plan.Slug.Equal(state.Slug) && !plan.Slug.IsNull() && !plan.Slug.IsUnknown() {
		payload["slug"] = plan.Slug.ValueString()
	}
	if !plan.Color.Equal(state.Color) {
		payload["color"] = plan.Color.ValueString()
	}
	if !plan.VmRole.Equal(state.VmRole) {
		payload["vm_role"] = plan.VmRole.ValueBool()
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

		path := fmt.Sprintf("api/dcim/device-roles/%d/", state.Id.ValueInt64())
		bodyStr, err := r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating Device Role", err.Error())
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
		if vmRole, ok := apiResponse["vm_role"].(bool); ok {
			plan.VmRole = types.BoolValue(vmRole)
		}
	} else if plan.Slug.IsNull() || plan.Slug.IsUnknown() {
		plan.Slug = state.Slug
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceRoleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/device-roles/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
