// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"

	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ resource.Resource = &ipAddressRangeResource{}
var _ resource.ResourceWithConfigure = &ipAddressRangeResource{}

func NewIpAddressRangeResource() resource.Resource {
	return &ipAddressRangeResource{}
}

type ipAddressRangeResource struct {
	client *client.NetboxClient
}

type ipAddressRangeResourceModel struct {
	IpRange      types.String `tfsdk:"ip_range"`
	Status       types.String `tfsdk:"status"`
	Description  types.String `tfsdk:"description"`
	Id           types.Int64  `tfsdk:"id"`
	StartAddress types.String `tfsdk:"start_address"`
	EndAddress   types.String `tfsdk:"end_address"`
}

var ipRangeRegex = regexp.MustCompile(`^(\d+\.\d+\.\d+)\.\[(\d+)-(\d+)\]\/(\d+)$`)

// parseIPRange は "10.18.48.[224-239]/24" のような記法を start/end アドレスに変換します。
func parseIPRange(ipRange string) (startAddress, endAddress string, err error) {
	matches := ipRangeRegex.FindStringSubmatch(ipRange)
	if matches == nil {
		return "", "", fmt.Errorf("invalid IP range format: %q (expected format: x.x.x.[start-end]/mask)", ipRange)
	}

	prefix := matches[1]
	start, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", "", fmt.Errorf("invalid start value: %s", matches[2])
	}
	end, err := strconv.Atoi(matches[3])
	if err != nil {
		return "", "", fmt.Errorf("invalid end value: %s", matches[3])
	}
	mask := matches[4]

	if start > end {
		return "", "", fmt.Errorf("start (%d) must be <= end (%d)", start, end)
	}
	if start < 0 || end > 255 {
		return "", "", fmt.Errorf("octet values must be between 0 and 255")
	}

	startAddress = fmt.Sprintf("%s.%d/%s", prefix, start, mask)
	endAddress = fmt.Sprintf("%s.%d/%s", prefix, end, mask)
	return startAddress, endAddress, nil
}

func (r *ipAddressRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address_range"
}

func (r *ipAddressRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "IP アドレス範囲記法 (例: `10.18.48.[224-239]/24`) を使って Netbox に IP レンジを作成します。",
		Attributes: map[string]schema.Attribute{
			"ip_range": schema.StringAttribute{
				MarkdownDescription: "IP 範囲記法 (例: `10.18.48.[224-239]/24`)。変更すると既存リソースを削除して再作成します。",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "IP レンジのステータス (例: active, reserved)。",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "IP レンジの説明。",
				Optional:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "IP レンジの Netbox ID。",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"start_address": schema.StringAttribute{
				MarkdownDescription: "レンジの開始 IP アドレス (例: 10.18.48.224/24)。",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"end_address": schema.StringAttribute{
				MarkdownDescription: "レンジの終了 IP アドレス (例: 10.18.48.239/24)。",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ipAddressRangeResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.NetboxClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.NetboxClient, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *ipAddressRangeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ipAddressRangeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	startAddr, endAddr, err := parseIPRange(plan.IpRange.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid IP range", err.Error())
		return
	}

	payload := map[string]any{
		"start_address": startAddr,
		"end_address":   endAddr,
	}
	if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
		payload["status"] = plan.Status.ValueString()
	}
	if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
		payload["description"] = plan.Description.ValueString()
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError("Error marshaling payload", err.Error())
		return
	}

	bodyStr, err := r.client.Post(ctx, "api/ipam/ip-ranges/", bytes.NewReader(bodyBytes))
	if err != nil {
		resp.Diagnostics.AddError("Error creating IP range", err.Error())
		return
	}

	var apiResponse map[string]any
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
	plan.StartAddress = types.StringValue(startAddr)
	plan.EndAddress = types.StringValue(endAddr)

	if sa, ok := apiResponse["start_address"].(string); ok {
		plan.StartAddress = types.StringValue(sa)
	}
	if ea, ok := apiResponse["end_address"].(string); ok {
		plan.EndAddress = types.StringValue(ea)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipAddressRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipAddressRangeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := fmt.Sprintf("api/ipam/ip-ranges/%d/", state.Id.ValueInt64())
	bodyStr, err := r.client.Get(ctx, apiPath)
	if err != nil {
		tflog.Warn(ctx, "Could not read IP range, assuming it was deleted", map[string]any{"error": err.Error()})
		resp.State.RemoveResource(ctx)
		return
	}

	var apiResponse map[string]any
	if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
		resp.Diagnostics.AddError("Error parsing read response", err.Error())
		return
	}

	if sa, ok := apiResponse["start_address"].(string); ok {
		state.StartAddress = types.StringValue(sa)
	}
	if ea, ok := apiResponse["end_address"].(string); ok {
		state.EndAddress = types.StringValue(ea)
	}

	if statusMap, ok := apiResponse["status"].(map[string]any); ok {
		if val, ok := statusMap["value"].(string); ok && !state.Status.IsNull() {
			state.Status = types.StringValue(val)
		}
	}

	if desc, ok := apiResponse["description"].(string); ok && !state.Description.IsNull() {
		state.Description = types.StringValue(desc)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ipAddressRangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ipAddressRangeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	payload := map[string]any{}
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

		apiPath := fmt.Sprintf("api/ipam/ip-ranges/%d/", state.Id.ValueInt64())
		_, err = r.client.Patch(ctx, apiPath, bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError("Error updating IP range", err.Error())
			return
		}
	}

	plan.Id = state.Id
	plan.StartAddress = state.StartAddress
	plan.EndAddress = state.EndAddress
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipAddressRangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipAddressRangeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiPath := fmt.Sprintf("api/ipam/ip-ranges/%d/", state.Id.ValueInt64())
	if err := r.client.Delete(ctx, apiPath); err != nil {
		tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]any{"error": err.Error()})
	}
}
