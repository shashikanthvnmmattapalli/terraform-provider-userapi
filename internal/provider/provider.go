package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &UserAPIProvider{}

type UserAPIProvider struct {
	version string
}

type UserAPIProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

func (p *UserAPIProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "userapi"
	resp.Version = p.version
}

func (p *UserAPIProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "User API endpoint",
				Optional:            true,
			},
		},
	}
}

func (p *UserAPIProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data UserAPIProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Endpoint.IsNull() {
		data.Endpoint = types.StringValue("http://localhost:5000")
	}

	client := &APIClient{
		Endpoint: data.Endpoint.ValueString(),
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *UserAPIProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewUserResource,
	}
}

func (p *UserAPIProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewUserDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UserAPIProvider{
			version: version,
		}
	}
}

type APIClient struct {
	Endpoint string
}
