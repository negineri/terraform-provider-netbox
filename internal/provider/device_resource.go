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

var _ resource.Resource = &deviceResource{}
var _ resource.ResourceWithConfigure = &deviceResource{}

func NewDeviceResource() resource.Resource {
	return &deviceResource{}
}

type deviceResource struct {
	client *client.NetboxClient
}

type deviceResourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	DeviceTypeId types.Int64  `tfsdk:"device_type_id"`
	RoleId       types.Int64  `tfsdk:"role_id"`
	SiteId       types.Int64  `tfsdk:"site_id"`
	LocationId   types.Int64  `tfsdk:"location_id"`
	RackId       types.Int64  `tfsdk:"rack_id"`
	Position     types.Int64  `tfsdk:"position"`
	Face         types.String `tfsdk:"face"`
	TenantId     types.Int64  `tfsdk:"tenant_id"`
	PlatformId   types.Int64  `tfsdk:"platform_id"`
	Status       types.String `tfsdk:"status"`
	Description  types.String `tfsdk:"description"`
	Serial       types.String `tfsdk:"serial"`
	AssetTag     types.String `tfsdk:"asset_tag"`
	Tags         types.Set    `tfsdk:"tags"`
	CustomFields types.Map    `tfsdk:"custom_fields"`
}

func (r *deviceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_device"
}

func (r *deviceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Device in Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the device.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the device.",
				Required:            true,
			},
			"device_type_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the device type.",
				Required:            true,
			},
			"role_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the device role.",
				Required:            true,
			},
			"site_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the site where the device is located.",
				Required:            true,
			},
			"location_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the location within the site.",
				Optional:            true,
			},
			"rack_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the rack where the device is installed.",
				Optional:            true,
			},
			"position": schema.Int64Attribute{
				MarkdownDescription: "The rack unit position of the device.",
				Optional:            true,
			},
			"face": schema.StringAttribute{
				MarkdownDescription: "The face of the rack unit (front or rear).",
				Optional:            true,
			},
			"tenant_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the tenant assigned to the device.",
				Optional:            true,
			},
			"platform_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the platform assigned to the device.",
				Optional:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The status of the device (e.g., active, offline, planned, staged, failed, decommissioning).",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the device.",
				Optional:            true,
			},
			"serial": schema.StringAttribute{
				MarkdownDescription: "Serial number of the device.",
				Optional:            true,
			},
			"asset_tag": schema.StringAttribute{
				MarkdownDescription: "A unique tag used to identify the device.",
				Optional:            true,
			},
			"tags": schema.SetAttribute{
				ElementType:         types.Int64Type,
				MarkdownDescription: "Set of tag IDs to assign to the device.",
				Optional:            true,
			},
			"custom_fields": customFieldsSchema(),
		},
	}
}

func (r *deviceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *deviceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan deviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{
		"name":        plan.Name.ValueString(),
		"device_type": map[string]interface{}{"id": plan.DeviceTypeId.ValueInt64()},
		"role":        map[string]interface{}{"id": plan.RoleId.ValueInt64()},
		"site":        map[string]interface{}{"id": plan.SiteId.ValueInt64()},
	}
	if !plan.LocationId.IsNull() && !plan.LocationId.IsUnknown() {
		payload["location"] = map[string]interface{}{"id": plan.LocationId.ValueInt64()}
	}
	if !plan.RackId.IsNull() && !plan.RackId.IsUnknown() {
		payload["rack"] = map[string]interface{}{"id": plan.RackId.ValueInt64()}
	}
	if !plan.Position.IsNull() && !plan.Position.IsUnknown() {
		payload["position"] = plan.Position.ValueInt64()
	}
	if !plan.Face.IsNull() && !plan.Face.IsUnknown() {
		payload["face"] = plan.Face.ValueString()
	}
	if !plan.TenantId.IsNull() && !plan.TenantId.IsUnknown() {
		payload["tenant"] = map[string]interface{}{"id": plan.TenantId.ValueInt64()}
	}
	if !plan.PlatformId.IsNull() && !plan.PlatformId.IsUnknown() {
		payload["platform"] = map[string]interface{}{"id": plan.PlatformId.ValueInt64()}
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.Serial.IsNull() && !plan.Serial.IsUnknown() {
		payload["serial"] = plan.Serial.ValueString()
	}
	if !plan.AssetTag.IsNull() && !plan.AssetTag.IsUnknown() {
		payload["asset_tag"] = plan.AssetTag.ValueString()
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

	bodyStr, err := r.client.Post(ctx, "api/dcim/devices/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating Device", err.Error())
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

func (r *deviceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state deviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/devices/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read device, assuming it was deleted", map[string]interface{}{"error": err.Error()})
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

	if deviceTypeMap, ok := apiResponse["device_type"].(map[string]interface{}); ok {
		if idFloat, ok := deviceTypeMap["id"].(float64); ok {
			state.DeviceTypeId = types.Int64Value(int64(idFloat))
		}
	}

	if roleMap, ok := apiResponse["role"].(map[string]interface{}); ok {
		if idFloat, ok := roleMap["id"].(float64); ok {
			state.RoleId = types.Int64Value(int64(idFloat))
		}
	}

	if siteMap, ok := apiResponse["site"].(map[string]interface{}); ok {
		if idFloat, ok := siteMap["id"].(float64); ok {
			state.SiteId = types.Int64Value(int64(idFloat))
		}
	}

	if locationMap, ok := apiResponse["location"].(map[string]interface{}); ok {
		if idFloat, ok := locationMap["id"].(float64); ok {
			// rack 指定時に NetBox が自動設定した location を無視して null を維持する
			if !state.LocationId.IsNull() || state.RackId.IsNull() {
				state.LocationId = types.Int64Value(int64(idFloat))
			}
		}
	} else if !state.LocationId.IsNull() {
		state.LocationId = types.Int64Null()
	}

	if rackMap, ok := apiResponse["rack"].(map[string]interface{}); ok {
		if idFloat, ok := rackMap["id"].(float64); ok {
			state.RackId = types.Int64Value(int64(idFloat))
		}
	} else if !state.RackId.IsNull() {
		state.RackId = types.Int64Null()
	}

	if position, ok := apiResponse["position"].(float64); ok {
		state.Position = types.Int64Value(int64(position))
	} else if !state.Position.IsNull() {
		state.Position = types.Int64Null()
	}

	if faceMap, ok := apiResponse["face"].(map[string]interface{}); ok {
		if val, ok := faceMap["value"].(string); ok && !state.Face.IsNull() {
			state.Face = types.StringValue(val)
		}
	}

	if tenantMap, ok := apiResponse["tenant"].(map[string]interface{}); ok {
		if idFloat, ok := tenantMap["id"].(float64); ok {
			state.TenantId = types.Int64Value(int64(idFloat))
		}
	} else if !state.TenantId.IsNull() {
		state.TenantId = types.Int64Null()
	}

	if platformMap, ok := apiResponse["platform"].(map[string]interface{}); ok {
		if idFloat, ok := platformMap["id"].(float64); ok {
			state.PlatformId = types.Int64Value(int64(idFloat))
		}
	} else if !state.PlatformId.IsNull() {
		state.PlatformId = types.Int64Null()
	}

	if statusMap, ok := apiResponse["status"].(map[string]interface{}); ok {
		if val, ok := statusMap["value"].(string); ok && !state.Status.IsNull() {
			state.Status = types.StringValue(val)
		}
	}

	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
	}

	if serial, ok := apiResponse["serial"].(string); ok && !state.Serial.IsNull() {
		state.Serial = types.StringValue(serial)
	}

	if assetTag, ok := apiResponse["asset_tag"].(string); ok && !state.AssetTag.IsNull() {
		state.AssetTag = types.StringValue(assetTag)
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

func (r *deviceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state deviceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}
	if !plan.Name.Equal(state.Name) {
		payload["name"] = plan.Name.ValueString()
	}
	if !plan.DeviceTypeId.Equal(state.DeviceTypeId) {
		payload["device_type"] = map[string]interface{}{"id": plan.DeviceTypeId.ValueInt64()}
	}
	if !plan.RoleId.Equal(state.RoleId) {
		payload["role"] = map[string]interface{}{"id": plan.RoleId.ValueInt64()}
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
	if !plan.RackId.Equal(state.RackId) {
		if plan.RackId.IsNull() {
			payload["rack"] = nil
		} else {
			payload["rack"] = map[string]interface{}{"id": plan.RackId.ValueInt64()}
		}
	}
	if !plan.Position.Equal(state.Position) {
		if plan.Position.IsNull() {
			payload["position"] = nil
		} else {
			payload["position"] = plan.Position.ValueInt64()
		}
	}
	if !plan.Face.Equal(state.Face) {
		payload["face"] = plan.Face.ValueString()
	}
	if !plan.TenantId.Equal(state.TenantId) {
		if plan.TenantId.IsNull() {
			payload["tenant"] = nil
		} else {
			payload["tenant"] = map[string]interface{}{"id": plan.TenantId.ValueInt64()}
		}
	}
	if !plan.PlatformId.Equal(state.PlatformId) {
		if plan.PlatformId.IsNull() {
			payload["platform"] = nil
		} else {
			payload["platform"] = map[string]interface{}{"id": plan.PlatformId.ValueInt64()}
		}
	}
	if !plan.Status.Equal(state.Status) {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.Description.Equal(state.Description) {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.Serial.Equal(state.Serial) {
		payload["serial"] = plan.Serial.ValueString()
	}
	if !plan.AssetTag.Equal(state.AssetTag) {
		if plan.AssetTag.IsNull() {
			payload["asset_tag"] = nil
		} else {
			payload["asset_tag"] = plan.AssetTag.ValueString()
		}
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

		path := fmt.Sprintf("api/dcim/devices/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating Device", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *deviceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state deviceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/devices/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
