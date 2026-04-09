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

var _ resource.Resource = &tagResource{}
var _ resource.ResourceWithConfigure = &tagResource{}

func NewTagResource() resource.Resource {
	return &tagResource{}
}

type tagResource struct {
	client *client.NetboxClient
}

type tagResourceModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Slug        types.String `tfsdk:"slug"`
	Color       types.String `tfsdk:"color"`
	Description types.String `tfsdk:"description"`
}

func (r *tagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (r *tagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Tag in Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the tag.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the tag.",
				Required:            true,
			},
			"slug": schema.StringAttribute{
				MarkdownDescription: "URL-friendly unique shorthand for the tag. If omitted, auto-generated from name.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"color": schema.StringAttribute{
				MarkdownDescription: "The color of the tag as a 6-digit hex code (e.g., aa1409).",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the tag.",
				Optional:            true,
			},
		},
	}
}

func (r *tagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	slug := plan.Slug.ValueString()
	if plan.Slug.IsNull() || plan.Slug.IsUnknown() || slug == "" {
		slug = slugify(plan.Name.ValueString())
	}

	payload := map[string]interface{}{
		"name": plan.Name.ValueString(),
		"slug": slug,
	}
	if !plan.Color.IsNull() && !plan.Color.IsUnknown() {
		payload["color"] = plan.Color.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	bodyStr, err := r.client.Post(ctx, "api/extras/tags/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating Tag", err.Error())
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

func (r *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/extras/tags/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read tag, assuming it was deleted", map[string]interface{}{"error": err.Error()})
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
	if color, ok := apiResponse["color"].(string); ok && !state.Color.IsNull() {
		state.Color = types.StringValue(color)
	}
	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state tagResourceModel
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
	if !plan.Description.Equal(state.Description) {
		payload["description"] = plan.Description.ValueString()
	}

	if len(payload) > 0 {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		path := fmt.Sprintf("api/extras/tags/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating Tag", err.Error())
			return
		}
	}

	plan.Id = state.Id
	if plan.Slug.IsNull() || plan.Slug.IsUnknown() {
		plan.Slug = state.Slug
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/extras/tags/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
