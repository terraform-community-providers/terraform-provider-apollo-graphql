package provider

import (
	"context"
	"net/http"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/Khan/genqlient/graphql"
)

var (
	envVarName          = "APOLLO_GRAPHQL_TOKEN"
	errMissingAuthToken = "Required token could not be found. Please set the token using an input variable in the provider configuration block or by using the `" + envVarName + "` environment variable."
)

var _ provider.Provider = &ApolloGraphQLProvider{}

type ApolloGraphQLProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type ApolloGraphQLProviderModel struct {
	Token types.String `tfsdk:"token"`
}

func (p *ApolloGraphQLProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "apollographql"
	resp.Version = p.version
}

func (p *ApolloGraphQLProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				MarkdownDescription: "The token used to authenticate with Apollo GraphQL.",
				Optional:            true,
			},
		},
	}
}

func (p *ApolloGraphQLProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ApolloGraphQLProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	token := ""

	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}

	// If a token wasn't set in the provider configuration block, try and fetch it
	// from the environment variable.
	if token == "" {
		token = os.Getenv(envVarName)
	}

	// If we still don't have a token at this point, we return an error.
	if token == "" {
		resp.Diagnostics.AddError("Missing API token", errMissingAuthToken)
		return
	}

	httpClient := http.Client{
		Transport: &authedTransport{
			token:   token,
			wrapped: http.DefaultTransport,
		},
	}

	client := graphql.NewClient("https://graphql.api.apollographql.com/api/graphql", &httpClient)

	resp.DataSourceData = &client
	resp.ResourceData = &client
}

func (p *ApolloGraphQLProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGraphResource,
		NewVariantResource,
		NewKeyResource,
	}
}

func (p *ApolloGraphQLProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ApolloGraphQLProvider{
			version: version,
		}
	}
}
