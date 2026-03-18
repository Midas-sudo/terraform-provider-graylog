package provider

import (
	"context"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-graylog/internal/client"
)

var _ provider.Provider = &GraylogProvider{}

type GraylogProvider struct {
	version string
}

type GraylogProviderModel struct {
	Endpoint           types.String `tfsdk:"endpoint"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	Token              types.String `tfsdk:"token"`
	InsecureSkipVerify types.Bool   `tfsdk:"insecure_skip_verify"`
	CACert             types.String `tfsdk:"ca_cert"`
	Timeout            types.Int64  `tfsdk:"timeout"`
}

func (p *GraylogProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "graylog"
	resp.Version = p.version
}

func (p *GraylogProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Graylog provider configures Graylog resources via its REST API. " +
			"It supports Graylog 6.x and aims for forward compatibility with 7.x.",
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Base URL of the Graylog REST API (e.g. `https://graylog.example.com/api`). " +
					"Can also be set via the `GRAYLOG_ENDPOINT` environment variable.",
				Optional: true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Username for HTTP Basic authentication. " +
					"Can also be set via the `GRAYLOG_USERNAME` environment variable.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for HTTP Basic authentication. " +
					"Can also be set via the `GRAYLOG_PASSWORD` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "API token for bearer authentication. Mutually exclusive with username/password. " +
					"Can also be set via the `GRAYLOG_TOKEN` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification. Defaults to `false`.",
				Optional:            true,
			},
			"ca_cert": schema.StringAttribute{
				MarkdownDescription: "PEM-encoded CA certificate to trust for the Graylog server's TLS certificate.",
				Optional:            true,
			},
			"timeout": schema.Int64Attribute{
				MarkdownDescription: "HTTP request timeout in seconds. Defaults to `30`.",
				Optional:            true,
			},
		},
	}
}

func (p *GraylogProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data GraylogProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := stringValueOrEnv(data.Endpoint, "GRAYLOG_ENDPOINT")
	if endpoint == "" {
		resp.Diagnostics.AddError(
			"Missing Graylog Endpoint",
			"The provider requires an endpoint. Set it in the provider block or via GRAYLOG_ENDPOINT.",
		)
		return
	}

	username := stringValueOrEnv(data.Username, "GRAYLOG_USERNAME")
	password := stringValueOrEnv(data.Password, "GRAYLOG_PASSWORD")
	token := stringValueOrEnv(data.Token, "GRAYLOG_TOKEN")

	insecure := false
	if !data.InsecureSkipVerify.IsNull() {
		insecure = data.InsecureSkipVerify.ValueBool()
	}

	caCert := ""
	if !data.CACert.IsNull() {
		caCert = data.CACert.ValueString()
	}

	var timeout time.Duration
	if !data.Timeout.IsNull() {
		timeout = time.Duration(data.Timeout.ValueInt64()) * time.Second
	}

	c, err := client.New(client.Config{
		Endpoint:           endpoint,
		Username:           username,
		Password:           password,
		Token:              token,
		InsecureSkipVerify: insecure,
		CACert:             caCert,
		Timeout:            timeout,
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create Graylog client", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *GraylogProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewContentPackResource,
		NewContentPackInstallationResource,
		NewOutputResource,
		NewExtractorResource,
		NewGrokPatternResource,
		NewLookupDataAdapterResource,
		NewLookupCacheResource,
		NewLookupTableResource,
		NewEventDefinitionResource,
		NewEventNotificationResource,
		NewEventDefinitionNotificationBindingResource,
		NewViewResource,
		NewDashboardResource,
		NewRoleResource,
		NewUserResource,
		NewEntityShareResource,
		NewIndexSetResource,
		NewInputResource,
		NewStreamResource,
		NewStreamRuleResource,
		NewPipelineResource,
		NewPipelineRuleResource,
		NewPipelineConnectionResource,
	}
}

func (p *GraylogProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewContentPackDataSource,
		NewContentPacksDataSource,
		NewOutputDataSource,
		NewOutputsDataSource,
		NewExtractorDataSource,
		NewExtractorsDataSource,
		NewGrokPatternsDataSource,
		NewLookupDataAdapterDataSource,
		NewLookupDataAdaptersDataSource,
		NewLookupCacheDataSource,
		NewLookupCachesDataSource,
		NewLookupTableDataSource,
		NewLookupTablesDataSource,
		NewEventDefinitionDataSource,
		NewEventDefinitionsDataSource,
		NewEventNotificationDataSource,
		NewEventNotificationsDataSource,
		NewViewDataSource,
		NewViewsDataSource,
		NewDashboardDataSource,
		NewDashboardsDataSource,
		NewRoleDataSource,
		NewRolesDataSource,
		NewUserDataSource,
		NewUsersDataSource,
		NewIndexSetDataSource,
		NewIndexSetsDataSource,
		NewIndexTemplateDataSource,
		NewInputDataSource,
		NewInputsDataSource,
		NewInputTypesDataSource,
		NewStreamDataSource,
		NewStreamsDataSource,
		NewPipelineDataSource,
		NewPipelinesDataSource,
		NewPipelineRuleDataSource,
		NewPipelineRulesDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GraylogProvider{
			version: version,
		}
	}
}

func stringValueOrEnv(val types.String, envKey string) string {
	if !val.IsNull() && !val.IsUnknown() {
		return val.ValueString()
	}
	return os.Getenv(envKey)
}
