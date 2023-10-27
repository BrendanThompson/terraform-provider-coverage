// Copyright (c) Brendan Thompson

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &CoverageProvider{}

// CoverageProvider defines the provider implementation.
type CoverageProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ScaffoldingProviderModel describes the provider data model.
type ScaffoldingProviderModel struct{}

func (p *CoverageProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "coverage"
	resp.Version = p.version
}

func (p *CoverageProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{},
	}
}

func (p *CoverageProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ScaffoldingProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

}

func (p *CoverageProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{}
}

func (p *CoverageProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewExamplesValidationDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CoverageProvider{
			version: version,
		}
	}
}
