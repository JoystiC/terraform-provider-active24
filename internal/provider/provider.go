package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type Active24Provider struct {
	version string
}

// Ensure provider implementation
var _ provider.Provider = &Active24Provider{}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Active24Provider{version: version}
	}
}

type providerModel struct {
	APIKey    types.String `tfsdk:"api_key"`
	APISecret types.String `tfsdk:"api_secret"`
	// Deprecated/compat fields
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	APIToken types.String `tfsdk:"api_token"`
	BaseURL  types.String `tfsdk:"base_url"`
}

func (p *Active24Provider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "active24"
	resp.Version = p.version
}

func (p *Active24Provider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   false,
				Description: "Active24 API key (username for HMAC Basic). Also via ACTIVE24_API_KEY.",
			},
			"api_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Active24 API secret (used to sign requests). Also via ACTIVE24_API_SECRET.",
			},
			// Deprecated/compat inputs
			"username":  schema.StringAttribute{Optional: true, Description: "Deprecated. Use api_key."},
			"password":  schema.StringAttribute{Optional: true, Sensitive: true, Description: "Deprecated. Use api_secret."},
			"api_token": schema.StringAttribute{Optional: true, Sensitive: true, Description: "Deprecated."},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL for Active24 API.",
			},
		},
	}
}

func (p *Active24Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := config.APIKey.ValueString()
	apiSecret := config.APISecret.ValueString()
	if apiKey == "" {
		if v := getEnv("ACTIVE24_API_KEY"); v != "" {
			apiKey = v
		}
	}
	if apiSecret == "" {
		if v := getEnv("ACTIVE24_API_SECRET"); v != "" {
			apiSecret = v
		}
	}
	// Backwards-compat fallbacks
	if apiKey == "" && config.Username.ValueString() != "" {
		apiKey = config.Username.ValueString()
	}
	if apiSecret == "" && config.Password.ValueString() != "" {
		apiSecret = config.Password.ValueString()
	}
	if apiSecret == "" && config.APIToken.ValueString() != "" {
		apiSecret = config.APIToken.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError("Missing API key", "Set `api_key` or env ACTIVE24_API_KEY.")
		return
	}
	if apiSecret == "" {
		resp.Diagnostics.AddError("Missing API secret", "Set `api_secret` or env ACTIVE24_API_SECRET.")
		return
	}

	baseURL := config.BaseURL.ValueString()
	if baseURL == "" {
		// Default to v2 REST API base per docs
		baseURL = "https://rest.active24.cz/v2"
	}

	c, err := NewClient(baseURL, apiKey, apiSecret)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create client", fmt.Sprintf("error: %v", err))
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *Active24Provider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDNSRecordResource,
	}
}

func (p *Active24Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}
