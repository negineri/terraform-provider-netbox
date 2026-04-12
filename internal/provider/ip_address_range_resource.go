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

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	AllocatedIPs types.List   `tfsdk:"allocated_ips"`
}

type allocatedIPModel struct {
	Id        types.Int64  `tfsdk:"id"`
	IpAddress types.String `tfsdk:"ip_address"`
}

var allocatedIPAttrTypes = map[string]attr.Type{
	"id":         types.Int64Type,
	"ip_address": types.StringType,
}

var ipRangeRegex = regexp.MustCompile(`^(\d+\.\d+\.\d+)\.\[(\d+)-(\d+)\]\/(\d+)$`)

// parseIPRange は "10.18.48.[224-239]/24" のような記法を個別の IP アドレス一覧に展開します。
func parseIPRange(ipRange string) ([]string, error) {
	matches := ipRangeRegex.FindStringSubmatch(ipRange)
	if matches == nil {
		return nil, fmt.Errorf("invalid IP range format: %q (expected format: x.x.x.[start-end]/mask)", ipRange)
	}

	prefix := matches[1]
	start, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid start value: %s", matches[2])
	}
	end, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid end value: %s", matches[3])
	}
	mask := matches[4]

	if start > end {
		return nil, fmt.Errorf("start (%d) must be <= end (%d)", start, end)
	}
	if start < 0 || end > 255 {
		return nil, fmt.Errorf("octet values must be between 0 and 255")
	}

	ips := make([]string, 0, end-start+1)
	for i := start; i <= end; i++ {
		ips = append(ips, fmt.Sprintf("%s.%d/%s", prefix, i, mask))
	}
	return ips, nil
}

func (r *ipAddressRangeResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ip_address_range"
}

func (r *ipAddressRangeResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "IP アドレス範囲記法 (例: `10.18.48.[224-239]/24`) を使って複数の IP アドレスを一括で確保・削除します。",
		Attributes: map[string]schema.Attribute{
			"ip_range": schema.StringAttribute{
				MarkdownDescription: "IP 範囲記法 (例: `10.18.48.[224-239]/24`)。変更すると既存リソースを削除して再作成します。",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "範囲内すべての IP アドレスに適用するステータス (例: active, reserved, dhcp)。",
				Optional:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "範囲内すべての IP アドレスに適用する説明。",
				Optional:            true,
			},
			"allocated_ips": schema.ListNestedAttribute{
				MarkdownDescription: "確保された IP アドレスの一覧 (Netbox ID と IP アドレスを含む)。",
				Computed:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "IP アドレスの Netbox ID。",
							Computed:            true,
						},
						"ip_address": schema.StringAttribute{
							MarkdownDescription: "IP アドレス (例: 10.18.48.224/24)。",
							Computed:            true,
						},
					},
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

	ips, err := parseIPRange(plan.IpRange.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid IP range", err.Error())
		return
	}

	allocatedIPs := make([]attr.Value, 0, len(ips))

	for _, ip := range ips {
		payload := map[string]any{
			"address": ip,
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

		bodyStr, err := r.client.Post(ctx, "api/ipam/ip-addresses/", bytes.NewReader(bodyBytes))
		if err != nil {
			resp.Diagnostics.AddError(fmt.Sprintf("Error creating IP address %s", ip), err.Error())
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
		address, ok := apiResponse["address"].(string)
		if !ok {
			resp.Diagnostics.AddError("Error parsing create response", "Could not find 'address' in response")
			return
		}

		obj, diags := types.ObjectValue(allocatedIPAttrTypes, map[string]attr.Value{
			"id":         types.Int64Value(int64(idFloat)),
			"ip_address": types.StringValue(address),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		allocatedIPs = append(allocatedIPs, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: allocatedIPAttrTypes}, allocatedIPs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.AllocatedIPs = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipAddressRangeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ipAddressRangeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var storedIPs []allocatedIPModel
	resp.Diagnostics.Append(state.AllocatedIPs.ElementsAs(ctx, &storedIPs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newAllocatedIPs := make([]attr.Value, 0, len(storedIPs))
	for _, m := range storedIPs {
		apiPath := fmt.Sprintf("api/ipam/ip-addresses/%d/", m.Id.ValueInt64())
		bodyStr, err := r.client.Get(ctx, apiPath)
		if err != nil {
			tflog.Warn(ctx, "IP address not found, removing range from state", map[string]any{
				"id":    m.Id.ValueInt64(),
				"error": err.Error(),
			})
			resp.State.RemoveResource(ctx)
			return
		}

		var apiResponse map[string]any
		if err := json.Unmarshal([]byte(*bodyStr), &apiResponse); err != nil {
			resp.Diagnostics.AddError("Error parsing read response", err.Error())
			return
		}

		address := m.IpAddress.ValueString()
		if addr, ok := apiResponse["address"].(string); ok {
			address = addr
		}

		obj, diags := types.ObjectValue(allocatedIPAttrTypes, map[string]attr.Value{
			"id":         m.Id,
			"ip_address": types.StringValue(address),
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		newAllocatedIPs = append(newAllocatedIPs, obj)
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: allocatedIPAttrTypes}, newAllocatedIPs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.AllocatedIPs = list
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ipAddressRangeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ipAddressRangeResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Status.Equal(state.Status) || !plan.Description.Equal(state.Description) {
		payload := map[string]any{}
		if !plan.Status.IsNull() && !plan.Status.IsUnknown() {
			payload["status"] = plan.Status.ValueString()
		}
		if !plan.Description.IsNull() && !plan.Description.IsUnknown() {
			payload["description"] = plan.Description.ValueString()
		}

		if len(payload) > 0 {
			var storedIPs []allocatedIPModel
			resp.Diagnostics.Append(state.AllocatedIPs.ElementsAs(ctx, &storedIPs, false)...)
			if resp.Diagnostics.HasError() {
				return
			}

			bodyBytes, err := json.Marshal(payload)
			if err != nil {
				resp.Diagnostics.AddError("Error marshaling payload", err.Error())
				return
			}

			for _, m := range storedIPs {
				apiPath := fmt.Sprintf("api/ipam/ip-addresses/%d/", m.Id.ValueInt64())
				_, err := r.client.Patch(ctx, apiPath, bytes.NewReader(bodyBytes))
				if err != nil {
					resp.Diagnostics.AddError(
						fmt.Sprintf("Error updating IP address %s", m.IpAddress.ValueString()),
						err.Error(),
					)
					return
				}
			}
		}
	}

	plan.AllocatedIPs = state.AllocatedIPs
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ipAddressRangeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ipAddressRangeResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var storedIPs []allocatedIPModel
	resp.Diagnostics.Append(state.AllocatedIPs.ElementsAs(ctx, &storedIPs, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for _, m := range storedIPs {
		apiPath := fmt.Sprintf("api/ipam/ip-addresses/%d/", m.Id.ValueInt64())
		if err := r.client.Delete(ctx, apiPath); err != nil {
			tflog.Warn(ctx, "Delete failed, assuming already deleted", map[string]any{
				"id":    m.Id.ValueInt64(),
				"error": err.Error(),
			})
		}
	}
}
