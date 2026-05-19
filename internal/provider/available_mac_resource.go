// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &availableMacResource{}
var _ resource.ResourceWithConfigure = &availableMacResource{}

var macPrefixRe = regexp.MustCompile(`^[0-9A-Fa-f]{2}(:[0-9A-Fa-f]{2}){0,4}$`)

func NewAvailableMacResource() resource.Resource {
	return &availableMacResource{}
}

type availableMacResource struct {
	client *client.NetboxClient
}

type availableMacResourceModel struct {
	Prefix             types.String `tfsdk:"prefix"`
	MaxAttempts        types.Int64  `tfsdk:"max_attempts"`
	Id                 types.Int64  `tfsdk:"id"`
	MacAddress         types.String `tfsdk:"mac_address"`
	AssignedObjectType types.String `tfsdk:"assigned_object_type"`
	AssignedObjectId   types.Int64  `tfsdk:"assigned_object_id"`
	Description        types.String `tfsdk:"description"`
	Comments           types.String `tfsdk:"comments"`
}

func (r *availableMacResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_available_mac"
}

func (r *availableMacResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Generates and registers an available MAC address with the given prefix. Retries on duplicate up to max_attempts times.",
		Attributes: map[string]schema.Attribute{
			"prefix": schema.StringAttribute{
				MarkdownDescription: "The MAC address prefix (1–5 octets, colon-separated, e.g. `AA:BB:CC`). The remaining octets are generated randomly.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"max_attempts": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of generation attempts before giving up on duplicate collision. Defaults to 10.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(10),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "The numeric ID of the registered MAC address.",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"mac_address": schema.StringAttribute{
				MarkdownDescription: "The registered MAC address.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"assigned_object_type": schema.StringAttribute{
				MarkdownDescription: "The type of the object this MAC address is assigned to (e.g., dcim.interface, virtualization.vminterface).",
				Optional:            true,
			},
			"assigned_object_id": schema.Int64Attribute{
				MarkdownDescription: "The ID of the object this MAC address is assigned to.",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description for the MAC address.",
				Optional:            true,
			},
			"comments": schema.StringAttribute{
				MarkdownDescription: "Comments for the MAC address.",
				Optional:            true,
			},
		},
	}
}

func (r *availableMacResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// generateMacAddress はプレフィックスに続くランダムなオクテットを生成し、完全な MAC アドレスを返します。
func generateMacAddress(prefix string) (string, error) {
	prefixOctets := strings.Split(prefix, ":")
	remaining := 6 - len(prefixOctets)

	randBytes := make([]byte, remaining)
	if _, err := rand.Read(randBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	octets := make([]string, 6)
	copy(octets, prefixOctets)
	for i := 0; i < remaining; i++ {
		octets[len(prefixOctets)+i] = fmt.Sprintf("%02X", randBytes[i])
	}

	return strings.Join(octets, ":"), nil
}

// macAddressCount は NetBox 上で指定した MAC アドレスが何件登録されているかを返します。
func (r *availableMacResource) macAddressCount(ctx context.Context, mac string) (int, error) {
	searchPath := fmt.Sprintf("api/dcim/mac-addresses/?mac_address=%s", strings.ReplaceAll(mac, ":", "%3A"))
	bodyStr, err := r.client.Get(ctx, searchPath)
	if err != nil {
		return 0, fmt.Errorf("failed to search MAC address: %w", err)
	}

	var searchResp map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &searchResp); err != nil {
		return 0, fmt.Errorf("failed to parse search response: %w", err)
	}

	countFloat, ok := searchResp["count"].(float64)
	if !ok {
		return 0, fmt.Errorf("could not find 'count' in search response")
	}

	return int(countFloat), nil
}

func (r *availableMacResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan availableMacResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	prefix := plan.Prefix.ValueString()
	if !macPrefixRe.MatchString(prefix) {
		resp.Diagnostics.AddError(
			"Invalid MAC address prefix",
			fmt.Sprintf("prefix must be 1–5 colon-separated hex octets (e.g. AA:BB:CC), got: %q", prefix),
		)
		return
	}

	maxAttempts := plan.MaxAttempts.ValueInt64()

	basePayload := map[string]interface{}{}
	if !plan.AssignedObjectId.IsNull() && !plan.AssignedObjectId.IsUnknown() {
		basePayload["assigned_object_id"] = plan.AssignedObjectId.ValueInt64()
		if !plan.AssignedObjectType.IsNull() && !plan.AssignedObjectType.IsUnknown() {
			basePayload["assigned_object_type"] = plan.AssignedObjectType.ValueString()
		}
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		basePayload["description"] = plan.Description.ValueString()
	}
	if !plan.Comments.IsNull() && !plan.Comments.IsUnknown() {
		basePayload["comments"] = plan.Comments.ValueString()
	}

	for attempt := int64(1); attempt <= maxAttempts; attempt++ {
		mac, err := generateMacAddress(prefix)
		if err != nil {
			resp.Diagnostics.AddError("Error generating MAC address", err.Error())
			return
		}

		tflog.Debug(ctx, "Attempting MAC address registration", map[string]interface{}{
			"attempt":     attempt,
			"mac_address": mac,
		})

		payload := map[string]interface{}{"mac_address": mac}
		for k, v := range basePayload {
			payload[k] = v
		}

		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		bodyStr, err := r.client.Post(ctx, "api/dcim/mac-addresses/", bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error creating MAC address", err.Error())
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
		createdId := int64(idFloat)

		// 登録後に同一 MAC アドレスが複数存在するか確認する
		count, err := r.macAddressCount(ctx, mac)
		if err != nil {
			resp.Diagnostics.AddError("Error checking MAC address uniqueness", err.Error())
			return
		}

		if count >= 2 {
			tflog.Debug(ctx, "MAC address is duplicate, deleting and retrying", map[string]interface{}{
				"attempt":     attempt,
				"mac_address": mac,
				"count":       count,
			})
			deletePath := fmt.Sprintf("api/dcim/mac-addresses/%d/", createdId)
			if delErr := r.client.Delete(ctx, deletePath); delErr != nil {
				tflog.Warn(ctx, "Failed to delete duplicate MAC address entry", map[string]interface{}{"error": delErr.Error()})
			}
			continue
		}

		plan.Id = types.Int64Value(createdId)
		if macVal, ok := apiResponse["mac_address"].(string); ok {
			plan.MacAddress = types.StringValue(macVal)
		} else {
			plan.MacAddress = types.StringValue(mac)
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	resp.Diagnostics.AddError(
		"Failed to register available MAC address",
		fmt.Sprintf("Could not find an available MAC address with prefix %q after %d attempts", prefix, maxAttempts),
	)
}

func (r *availableMacResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state availableMacResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/mac-addresses/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Could not read MAC address, assuming it was deleted", map[string]interface{}{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}

	var apiResponse map[string]interface{}
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing read response", err.Error())
		return
	}

	if macAddr, ok := apiResponse["mac_address"].(string); ok {
		state.MacAddress = types.StringValue(macAddr)
	}

	if objType, ok := apiResponse["assigned_object_type"].(string); ok && objType != "" {
		state.AssignedObjectType = types.StringValue(objType)
		if objIdFloat, ok := apiResponse["assigned_object_id"].(float64); ok {
			state.AssignedObjectId = types.Int64Value(int64(objIdFloat))
		}
	}

	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
	}

	if comments, ok := apiResponse["comments"].(string); ok && !state.Comments.IsNull() {
		state.Comments = types.StringValue(comments)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *availableMacResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state availableMacResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]interface{}{}
	if !plan.AssignedObjectId.Equal(state.AssignedObjectId) || !plan.AssignedObjectType.Equal(state.AssignedObjectType) {
		if plan.AssignedObjectId.IsNull() {
			payload["assigned_object_id"] = nil
			payload["assigned_object_type"] = nil
		} else {
			payload["assigned_object_id"] = plan.AssignedObjectId.ValueInt64()
			if !plan.AssignedObjectType.IsNull() && !plan.AssignedObjectType.IsUnknown() {
				payload["assigned_object_type"] = plan.AssignedObjectType.ValueString()
			}
		}
	}
	if !plan.Description.Equal(state.Description) {
		payload["description"] = plan.Description.ValueString()
	}
	if !plan.Comments.Equal(state.Comments) {
		payload["comments"] = plan.Comments.ValueString()
	}

	if len(payload) > 0 {
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			resp.Diagnostics.AddError("Error marshaling payload", err.Error())
			return
		}

		path := fmt.Sprintf("api/dcim/mac-addresses/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, path, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating MAC address", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *availableMacResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state availableMacResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("api/dcim/mac-addresses/%d/", state.Id.ValueInt64())
	err := r.client.Delete(ctx, path)
	if err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]interface{}{"error": err.Error()})
	}
}
