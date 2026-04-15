// Copyright (c) negineri 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"net/url"
	"os"
	"terraform-provider-netbox/internal/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// 実装が期待されるインターフェースを満たしていることを保証します。
var (
	_ provider.Provider = &netboxProvider{}
)

// New はプロバイダーサーバーやテスト実装を簡素化するヘルパー関数です。
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &netboxProvider{
			version: version,
		}
	}
}

// netboxProvider はプロバイダーの実装です。
type netboxProvider struct {
	// version はリリース時にプロバイダーのバージョンが設定され、ローカルでビルド・実行時は "dev"、受け入れテスト時は "test" となります。
	version string
}

// Metadata はプロバイダーのタイプ名を返します。
func (p *netboxProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "netbox"
	resp.Version = p.version
}

// Schema はプロバイダーの設定データ用スキーマを定義します。
// Schema defines the provider-level schema for configuration data.
func (p *netboxProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"server_url": schema.StringAttribute{
				Description: "URL for Netbox API. May also be provided via NETBOX_SERVER_URL environment variable.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "API v1 token for Netbox. May also be provided via NETBOX_TOKEN environment variable. Use this or the key_v2/token_v2 pair.",
				Optional:    true,
				Sensitive:   true,
			},
			"key_v2": schema.StringAttribute{
				Description: "API key for Netbox v2 token. May also be provided via NETBOX_KEY_V2 environment variable.",
				Optional:    true,
			},
			"token_v2": schema.StringAttribute{
				Description: "API token for Netbox v2 token. May also be provided via NETBOX_TOKEN_V2 environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure はデータソースやリソース用の Netbox API クライアントを準備します。
func (p *netboxProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config netboxProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.ServerURL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("server_url"),
			"Unknown Netbox API Server URL",
			"The provider cannot create the Netbox API client as there is an unknown configuration value for the Netbox API server URL. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NETBOX_SERVER_URL environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Netbox API Token (v1)",
			"The provider cannot create the Netbox API client as there is an unknown configuration value for the Netbox API v1 token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NETBOX_TOKEN environment variable.",
		)
	}

	if config.KeyV2.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("key_v2"),
			"Unknown Netbox API Key",
			"The provider cannot create the Netbox API client as there is an unknown configuration value for the Netbox API key. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NETBOX_KEY_V2 environment variable.",
		)
	}

	if config.TokenV2.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token_v2"),
			"Unknown Netbox API Token (v2)",
			"The provider cannot create the Netbox API client as there is an unknown configuration value for the Netbox API v2 token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the NETBOX_TOKEN_V2 environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	serverURL := os.Getenv("NETBOX_SERVER_URL")
	token := os.Getenv("NETBOX_TOKEN")
	keyV2 := os.Getenv("NETBOX_KEY_V2")
	tokenV2 := os.Getenv("NETBOX_TOKEN_V2")

	if !config.ServerURL.IsNull() {
		serverURL = config.ServerURL.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	if !config.KeyV2.IsNull() {
		keyV2 = config.KeyV2.ValueString()
	}

	if !config.TokenV2.IsNull() {
		tokenV2 = config.TokenV2.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if serverURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("server_url"),
			"Missing Netbox API Server URL",
			"The provider cannot create the Netbox API client as there is a missing or empty value for the Netbox API server URL. "+
				"Set the server URL value in the configuration or use the NETBOX_SERVER_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	// token (v1) か key_v2+token_v2 (v2) のどちらかが必要
	useV1 := token != ""
	useV2 := keyV2 != "" || tokenV2 != ""

	if !useV1 && !useV2 {
		resp.Diagnostics.AddError(
			"Missing Netbox API Credentials",
			"The provider cannot create the Netbox API client as no credentials were provided. "+
				"Provide either a v1 token (token / NETBOX_TOKEN) or a v2 key+token pair (key_v2+token_v2 / NETBOX_KEY_V2+NETBOX_TOKEN_V2).",
		)
	}

	if useV1 && useV2 {
		resp.Diagnostics.AddError(
			"Conflicting Netbox API Credentials",
			"Both a v1 token and v2 key+token pair were provided. Provide only one authentication method.",
		)
	}

	if useV2 {
		if keyV2 == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("key_v2"),
				"Missing Netbox API Key (v2)",
				"token_v2 was provided but key_v2 is missing. "+
					"Set the key value in the configuration or use the NETBOX_KEY_V2 environment variable.",
			)
		}
		if tokenV2 == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("token_v2"),
				"Missing Netbox API Token (v2)",
				"key_v2 was provided but token_v2 is missing. "+
					"Set the token value in the configuration or use the NETBOX_TOKEN_V2 environment variable.",
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new Netbox client using the configuration values
	_, err := url.Parse(serverURL)
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("server_url"),
			"Invalid Netbox API Server URL",
			"The provider cannot create the Netbox API client as the provided server URL is not a valid URL. "+
				"Ensure the server URL is a valid URL and includes the appropriate scheme (http:// or https://).\n\n"+
				"URL Parsing Error: "+err.Error(),
		)
		return
	}

	var netboxClient *client.NetboxClient
	if useV1 {
		netboxClient = client.NewNetboxClientV1(serverURL, token)
	} else {
		netboxClient = client.NewNetboxClient(serverURL, keyV2, tokenV2)
	}

	_, err = netboxClient.Get(ctx, "api/status/")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Netbox API Client",
			"The provider was able to parse the configuration, but was unable to create a Netbox API client that can successfully communicate with the Netbox API. "+
				"Ensure the server URL is correct and the server is reachable, and that the provided credentials are valid.\n\n"+
				"Client Error: "+err.Error(),
		)
		return
	}

	// Make the Netbox client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = netboxClient
	resp.ResourceData = netboxClient
}

// DataSources はプロバイダーで実装されているデータソースを定義します。
func (p *netboxProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDevicesDataSource,
		NewDeviceDataSource,
		NewDeviceRolesDataSource,
		NewDeviceRoleDataSource,
		NewLocationsDataSource,
		NewLocationDataSource,
		NewRegionsDataSource,
		NewRegionDataSource,
		NewSitesDataSource,
		NewSiteDataSource,
		NewTagsDataSource,
		NewTagDataSource,
		NewPrefixesDataSource,
		NewPrefixDataSource,
		NewIpAddressesDataSource,
		NewIpAddressDataSource,
		NewVlansDataSource,
		NewVlanDataSource,
		NewVlanGroupsDataSource,
		NewVlanGroupDataSource,
		NewVirtualMachinesDataSource,
		NewVirtualMachineDataSource,
		NewCustomFieldsDataSource,
		NewCustomFieldDataSource,
		NewRacksDataSource,
		NewRackDataSource,
		NewDeviceTypesDataSource,
		NewDeviceTypeDataSource,
	}
}

// Resources はプロバイダーで実装されているリソースを定義します。
func (p *netboxProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAvailableIpResource,
		NewAvailablePrefixResource,
		NewPrefixResource,
		NewIpAddressResource,
		NewIpAddressRangeResource,
		NewDeviceResource,
		NewDeviceRoleResource,
		NewDeviceInterfaceResource,
		NewDevicePrimaryIPResource,
		NewVirtualMachineResource,
		NewVirtualMachineInterfaceResource,
		NewVirtualMachinePrimaryIPResource,
		NewLocationResource,
		NewRegionResource,
		NewSiteResource,
		NewTagResource,
		NewVlanGroupResource,
		NewVlanResource,
		NewCustomFieldResource,
		NewRackResource,
		NewDeviceTypeResource,
	}
}

// netboxProviderModel maps provider schema data to a Go type.
type netboxProviderModel struct {
	ServerURL types.String `tfsdk:"server_url"`
	Token     types.String `tfsdk:"token"`
	KeyV2     types.String `tfsdk:"key_v2"`
	TokenV2   types.String `tfsdk:"token_v2"`
}
