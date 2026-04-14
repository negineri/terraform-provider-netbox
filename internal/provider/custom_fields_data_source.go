// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &customFieldsDataSource{}
var _ datasource.DataSourceWithConfigure = &customFieldsDataSource{}

func NewCustomFieldsDataSource() datasource.DataSource {
	return &customFieldsDataSource{}
}

type customFieldsDataSource struct {
	client *client.NetboxClient
}

type customFieldsDataSourceModel struct {
	Id           types.String       `tfsdk:"id"`
	CustomFields []customFieldModel `tfsdk:"custom_fields"`
}

type customFieldModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Label        types.String `tfsdk:"label"`
	Type         types.String `tfsdk:"type"`
	ContentTypes types.List   `tfsdk:"content_types"`
	Required     types.Bool   `tfsdk:"required"`
	Description  types.String `tfsdk:"description"`
	Weight       types.Int64  `tfsdk:"weight"`
	FilterLogic  types.String `tfsdk:"filter_logic"`
	GroupName    types.String `tfsdk:"group_name"`
	Choices      types.List   `tfsdk:"choices"`
	UIVisible    types.String `tfsdk:"ui_visible"`
	UIEditable   types.String `tfsdk:"ui_editable"`
	IsCloneable  types.Bool   `tfsdk:"is_cloneable"`
	SearchWeight types.Int64  `tfsdk:"search_weight"`
}

func customFieldNestedAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			MarkdownDescription: "The numeric ID of the custom field.",
			Computed:            true,
		},
		"name": schema.StringAttribute{
			MarkdownDescription: "Internal field name.",
			Computed:            true,
		},
		"label": schema.StringAttribute{
			MarkdownDescription: "Display name for the field in the UI.",
			Computed:            true,
		},
		"type": schema.StringAttribute{
			MarkdownDescription: "The type of the custom field.",
			Computed:            true,
		},
		"content_types": schema.ListAttribute{
			MarkdownDescription: "List of content types this custom field is assigned to.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"required": schema.BoolAttribute{
			MarkdownDescription: "Whether a value is required for this field.",
			Computed:            true,
		},
		"description": schema.StringAttribute{
			MarkdownDescription: "Description for the custom field.",
			Computed:            true,
		},
		"weight": schema.Int64Attribute{
			MarkdownDescription: "Fields with higher weights appear lower in the form.",
			Computed:            true,
		},
		"filter_logic": schema.StringAttribute{
			MarkdownDescription: "Filtering logic for this field.",
			Computed:            true,
		},
		"group_name": schema.StringAttribute{
			MarkdownDescription: "Custom fields within the same group will be displayed together.",
			Computed:            true,
		},
		"choices": schema.ListAttribute{
			MarkdownDescription: "Available choices for select/multiselect fields.",
			Computed:            true,
			ElementType:         types.StringType,
		},
		"ui_visible": schema.StringAttribute{
			MarkdownDescription: "Controls visibility in the UI.",
			Computed:            true,
		},
		"ui_editable": schema.StringAttribute{
			MarkdownDescription: "Controls editability in the UI.",
			Computed:            true,
		},
		"is_cloneable": schema.BoolAttribute{
			MarkdownDescription: "Whether this field is copied when cloning an object.",
			Computed:            true,
		},
		"search_weight": schema.Int64Attribute{
			MarkdownDescription: "Weighting for search.",
			Computed:            true,
		},
	}
}

func (d *customFieldsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_fields"
}

func (d *customFieldsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of Custom Fields from Netbox.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier for the data source.",
				Computed:            true,
			},
			"custom_fields": schema.ListNestedAttribute{
				MarkdownDescription: "List of custom fields.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: customFieldNestedAttributes(),
				},
			},
		},
	}
}

func (d *customFieldsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.NetboxClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.NetboxClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *customFieldsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state customFieldsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	state.CustomFields = []customFieldModel{}

	bodyStr, err := d.client.Get(ctx, "api/extras/custom-fields/")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to fetch custom fields, got error: %s", err))
		return
	}

	type ApiCustomField struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Label string `json:"label"`
		Type  struct {
			Value string `json:"value"`
		} `json:"type"`
		ContentTypes []string `json:"content_types"`
		Required     bool     `json:"required"`
		Description  string   `json:"description"`
		Weight       int64    `json:"weight"`
		FilterLogic  struct {
			Value string `json:"value"`
		} `json:"filter_logic"`
		GroupName string   `json:"group_name"`
		Choices   []string `json:"choices"`
		UIVisible struct {
			Value string `json:"value"`
		} `json:"ui_visible"`
		UIEditable struct {
			Value string `json:"value"`
		} `json:"ui_editable"`
		IsCloneable  bool  `json:"is_cloneable"`
		SearchWeight int64 `json:"search_weight"`
	}

	type ApiCustomFieldsResponse struct {
		Count   int              `json:"count"`
		Results []ApiCustomField `json:"results"`
	}

	var response ApiCustomFieldsResponse
	if err := json.Unmarshal([]byte(*bodyStr), &response); err != nil {
		resp.Diagnostics.AddError("JSON Parse Error", fmt.Sprintf("Unable to parse API response: %s", err))
		return
	}

	for _, result := range response.Results {
		contentTypes := make([]types.String, 0, len(result.ContentTypes))
		for _, ct := range result.ContentTypes {
			contentTypes = append(contentTypes, types.StringValue(ct))
		}
		ctList, diags := types.ListValueFrom(ctx, types.StringType, contentTypes)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		choices := make([]types.String, 0, len(result.Choices))
		for _, c := range result.Choices {
			choices = append(choices, types.StringValue(c))
		}
		choicesList, diags := types.ListValueFrom(ctx, types.StringType, choices)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		state.CustomFields = append(state.CustomFields, customFieldModel{
			Id:           types.Int64Value(result.ID),
			Name:         types.StringValue(result.Name),
			Label:        types.StringValue(result.Label),
			Type:         types.StringValue(result.Type.Value),
			ContentTypes: ctList,
			Required:     types.BoolValue(result.Required),
			Description:  types.StringValue(result.Description),
			Weight:       types.Int64Value(result.Weight),
			FilterLogic:  types.StringValue(result.FilterLogic.Value),
			GroupName:    types.StringValue(result.GroupName),
			Choices:      choicesList,
			UIVisible:    types.StringValue(result.UIVisible.Value),
			UIEditable:   types.StringValue(result.UIEditable.Value),
			IsCloneable:  types.BoolValue(result.IsCloneable),
			SearchWeight: types.Int64Value(result.SearchWeight),
		})
	}

	state.Id = types.StringValue("netbox_custom_fields")

	tflog.Trace(ctx, "read a data source")

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
