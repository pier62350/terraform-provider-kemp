// Copyright (c) Pierre Bailly
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/pier62350/terraform-provider-kemp/internal/loadmaster"
)

var _ provider.Provider = &KempProvider{}

type KempProvider struct {
	version string
}

type KempProviderModel struct {
	Host               types.String `tfsdk:"host"`
	APIKey             types.String `tfsdk:"api_key"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
}

func (p *KempProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "kemp"
	resp.Version = p.version
}

func (p *KempProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provider for Progress Kemp LoadMaster.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Base URL of the LoadMaster, including scheme and port (e.g. `https://10.0.0.5:9443`). May also be set via the `KEMP_HOST` environment variable.",
				Optional:            true,
			},
			"api_key": schema.StringAttribute{
				MarkdownDescription: "API key. May also be set via the `KEMP_API_KEY` environment variable. Mutually exclusive with `username`/`password` (api_key wins if both are set).",
				Optional:            true,
				Sensitive:           true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for basic auth. May also be set via the `KEMP_USERNAME` environment variable.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for basic auth. May also be set via the `KEMP_PASSWORD` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification. Defaults to true since LoadMasters typically present self-signed certificates. Set to false to enforce verification.",
				Optional:            true,
			},
		},
	}
}

func (p *KempProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data KempProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := envOr(data.Host, "KEMP_HOST")
	apiKey := envOr(data.APIKey, "KEMP_API_KEY")
	username := envOr(data.Username, "KEMP_USERNAME")
	password := envOr(data.Password, "KEMP_PASSWORD")

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing LoadMaster host",
			"Set the `host` attribute or the KEMP_HOST environment variable to the base URL of the LoadMaster (e.g. https://10.0.0.5:9443).",
		)
	}
	if apiKey == "" && (username == "" || password == "") {
		resp.Diagnostics.AddError(
			"Missing LoadMaster credentials",
			"Provide either `api_key` (or KEMP_API_KEY) or both `username` and `password` (or KEMP_USERNAME / KEMP_PASSWORD).",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	insecure := true
	if !data.InsecureSkipVerify.IsNull() && !data.InsecureSkipVerify.IsUnknown() {
		insecure = data.InsecureSkipVerify.ValueBool()
	}

	opts := []loadmaster.Option{loadmaster.WithInsecureSkipVerify(insecure)}
	if apiKey != "" {
		opts = append(opts, loadmaster.WithAPIKey(apiKey))
	} else {
		opts = append(opts, loadmaster.WithBasicAuth(username, password))
	}

	client := loadmaster.NewClient(host, opts...)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *KempProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVirtualServiceResource,
		NewSubVirtualServiceResource,
		NewRealServerResource,
		NewCertificateResource,
		NewACMECertificateResource,
		NewMatchContentRuleResource,
		NewAddHeaderRuleResource,
		NewDeleteHeaderRuleResource,
		NewReplaceHeaderRuleResource,
		NewModifyURLRuleResource,
		NewReplaceBodyRuleResource,
		NewVirtualServiceRuleResource,
		NewVirtualServiceWafRuleResource,
		NewOwaspCustomRuleResource,
		NewOwaspCustomDataResource,
		NewWafCustomRuleResource,
		NewWafCustomDataResource,
		NewACMEAccountResource,
	}
}

func (p *KempProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewVirtualServiceDataSource,
		NewSubVirtualServiceDataSource,
		NewRealServerDataSource,
		NewCertificateDataSource,
		NewACMECertificateDataSource,
		NewACMEAccountDataSource,
		NewMatchContentRuleDataSource,
		NewAddHeaderRuleDataSource,
		NewDeleteHeaderRuleDataSource,
		NewReplaceHeaderRuleDataSource,
		NewModifyURLRuleDataSource,
		NewReplaceBodyRuleDataSource,
		NewVirtualServiceRuleDataSource,
		NewVirtualServiceWafRuleDataSource,
		NewOwaspCustomRuleDataSource,
		NewOwaspCustomDataDataSource,
		NewWafCustomRuleDataSource,
		NewWafCustomDataDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &KempProvider{version: version}
	}
}

func envOr(v types.String, key string) string {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueString()
	}
	return os.Getenv(key)
}

func boolPtr(b bool) *bool { return &b }

// boolFromPtr converts a *bool from the API into a types.Bool, mapping nil
// to false (the conventional default for ESP / SSL-related toggles).
func boolFromPtr(b *bool) types.Bool {
	if b == nil {
		return types.BoolValue(false)
	}
	return types.BoolValue(*b)
}

// int64FromPtr converts a *int32 from the API into a types.Int64, mapping
// nil to zero.
func int64FromPtr(p *int32) types.Int64 {
	if p == nil {
		return types.Int64Value(0)
	}
	return types.Int64Value(int64(*p))
}
