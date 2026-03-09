package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &ShopifyProvider{}

type ShopifyProvider struct {
	version      string
	mockEndpoint string // Used in tests to override the API endpoint.
}

type ShopifyProviderModel struct {
	StoreURL    types.String `tfsdk:"store_url"`
	AccessToken types.String `tfsdk:"access_token"`
	APIVersion  types.String `tfsdk:"api_version"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ShopifyProvider{version: version}
	}
}

func (p *ShopifyProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "shopify"
	resp.Version = p.version
}

func (p *ShopifyProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Shopify store resources via the GraphQL Admin API.",
		Attributes: map[string]schema.Attribute{
			"store_url": schema.StringAttribute{
				Description: "Shopify store URL (e.g. your-store.myshopify.com). Can also be set via SHOPIFY_STORE_URL environment variable.",
				Optional:    true,
			},
			"access_token": schema.StringAttribute{
				Description: "Shopify Admin API access token. Can also be set via SHOPIFY_ACCESS_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_version": schema.StringAttribute{
				Description: "Shopify API version (e.g. 2025-04). Defaults to 2025-04.",
				Optional:    true,
			},
		},
	}
}

func (p *ShopifyProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ShopifyProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	storeURL := os.Getenv("SHOPIFY_STORE_URL")
	if !config.StoreURL.IsNull() {
		storeURL = config.StoreURL.ValueString()
	}

	accessToken := os.Getenv("SHOPIFY_ACCESS_TOKEN")
	if !config.AccessToken.IsNull() {
		accessToken = config.AccessToken.ValueString()
	}

	apiVersion := "2025-04"
	if !config.APIVersion.IsNull() {
		apiVersion = config.APIVersion.ValueString()
	}

	if storeURL == "" {
		resp.Diagnostics.AddError(
			"Missing store URL",
			"store_url must be set in provider config or SHOPIFY_STORE_URL environment variable.",
		)
		return
	}
	if accessToken == "" {
		resp.Diagnostics.AddError(
			"Missing access token",
			"access_token must be set in provider config or SHOPIFY_ACCESS_TOKEN environment variable.",
		)
		return
	}

	client := NewClient(storeURL, accessToken, apiVersion)
	if p.mockEndpoint != "" {
		client.Endpoint = p.mockEndpoint
	}
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ShopifyProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProductResource,
	}
}

func (p *ShopifyProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
