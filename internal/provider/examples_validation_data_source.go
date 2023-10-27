// Copyright (c) Brendan Thompson

package provider

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	// "net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ExamplesValidationDataSource{}

func NewExamplesValidationDataSource() datasource.DataSource {
	return &ExamplesValidationDataSource{}
}

// ExamplesValidationDataSource defines the data source implementation.
type ExamplesValidationDataSource struct {
}

// ExamplesValidationDataSourceModel describes the data source data model.
type ExamplesValidationDataSourceModel struct {
	Id                types.String `tfsdk:"id"`
	ExamplesDirectory types.String `tfsdk:"examples_directory"`
	TestsDirectory    types.String `tfsdk:"tests_directory"`
	Filter            types.String `tfsdk:"filter"`
	MissingTests      types.List   `tfsdk:"missing_tests"`
}

func (d *ExamplesValidationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_examples_validation"
}

func (d *ExamplesValidationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Validate that there are tests for all examples in the example directory.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID",
				Computed:            true,
			},
			"examples_directory": schema.StringAttribute{
				MarkdownDescription: "Filepath to the examples directory for the module.",
				Required:            true,
			},
			"tests_directory": schema.StringAttribute{
				MarkdownDescription: "Filepath to the tests directory for the module.",
				Required:            true,
			},
			"filter": schema.StringAttribute{
				MarkdownDescription: "Filter to use to find tests responsible for validating the examples.",
				Required:            true,
			},
			"missing_tests": schema.ListAttribute{
				MarkdownDescription: "List of example directories that are missing tests",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

func (d *ExamplesValidationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

}

func (d *ExamplesValidationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExamplesValidationDataSourceModel
	var diags diag.Diagnostics

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	data.Id = types.StringValue("examples-validation")

	examplesDirectory := data.ExamplesDirectory.ValueString()
	testsDirectory := data.TestsDirectory.ValueString()
	filter := data.Filter.ValueString()

	missingTests := findMissingTests(examplesDirectory, testsDirectory, filter, ctx)

	data.MissingTests, diags = types.ListValueFrom(ctx, types.StringType, missingTests)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func findMissingTests(examplesDirectory string, testsDirectory string, filter string, ctx context.Context) []string {
	var sourceFilter = `source = "./examples`
	var examples []string
	var tests []string
	var missing []string

	files, err := os.ReadDir(examplesDirectory)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if file.IsDir() {
			// tflog.Info(ctx, file.Name())
			examples = append(examples, file.Name())
		}
	}

	testFiles, err := os.ReadDir(testsDirectory)
	if err != nil {
		panic(err)
	}

	tflog.Info(ctx, "Source Filter: '"+sourceFilter+"'")

	for _, file := range testFiles {
		if !file.IsDir() && strings.Contains(file.Name(), filter) {
			filepath := filepath.Join(testsDirectory, file.Name())
			tflog.Info(ctx, filepath)
			tests = append(tests, searchFile(filepath, sourceFilter, examplesDirectory, ctx)...)
		}
	}

	for _, e := range examples {
		if !slices.Contains(tests, e) {
			missing = append(missing, e)
		}
	}

	return missing
}

func searchFile(file string, pattern string, examplesDirectory string, ctx context.Context) []string {
	var result []string

	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	r := regexp.MustCompile(`^\s*source\s*=\s*".+examples.+"$`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if r.MatchString(line) {
			re := regexp.MustCompile(`"(.*?)"`)
			matches := re.FindAllStringSubmatch(line, -1)

			for _, match := range matches {
				strippedText := filepath.Base(match[1])
				tflog.Info(ctx, strippedText)
				result = append(result, strippedText)
			}
		}
	}

	return result
}
