package multy

import (
	"context"
	"crypto/x509"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/mitchellh/go-homedir"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"terraform-provider-multy/multy/common"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type connectionCache struct {
	sync.Mutex
	cache map[string]proto.MultyResourceServiceClient
}

var connCache = connectionCache{cache: map[string]proto.MultyResourceServiceClient{}}
var refreshCache = &common.RefreshCache{}

func New() provider.Provider {
	return &Provider{}
}

type Provider struct {
	Configured bool
	Client     *common.ProviderConfig
}

var awsSchema = tfsdk.Attribute{
	Optional:    true,
	Computed:    true,
	Description: "Credentials for AWS Cloud",
	Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"access_key_id": {
			Optional:    true,
			Description: "AWS Access Key ID. " + common.HelperValueViaEnvVar("AWS_ACCESS_KEY_ID"),
			Type:        types.StringType,
			Sensitive:   true,
		},
		"access_key_secret": {
			Optional:    true,
			Description: "AWS Secret Access Key. " + common.HelperValueViaEnvVar("AWS_SECRET_ACCESS_KEY"),
			Type:        types.StringType,
			Sensitive:   true,
		},
		"session_token": {
			Optional:    true,
			Description: "Optional AWS session token. Used to authenticate  " + common.HelperValueViaEnvVar("AWS_SESSION_TOKEN"),
			Type:        types.StringType,
			Sensitive:   true,
		},
	}),
}

var azureSchema = tfsdk.Attribute{
	Optional:    true,
	Description: "Credentials for Azure Cloud. See how to authenticate through Service Principal in the [Azure docs](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/service_principal_client_secret#creating-a-service-principal)",
	Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"subscription_id": {
			Optional:    true,
			Description: "Azure Subscription ID. " + common.HelperValueViaEnvVar("ARM_SUBSCRIPTION_ID"),
			Type:        types.StringType,
			Sensitive:   true,
		},
		"client_id": {
			Optional:    true,
			Description: "Azure Client ID " + common.HelperValueViaEnvVar("ARM_CLIENT_ID"),
			Type:        types.StringType,
			Sensitive:   true,
		},
		"client_secret": {
			Optional:    true,
			Description: "Azure Client Secret " + common.HelperValueViaEnvVar("ARM_CLIENT_SECRET"),
			Type:        types.StringType,
			Sensitive:   true,
		},
		"tenant_id": {
			Optional:    true,
			Description: "Azure Tenant ID " + common.HelperValueViaEnvVar("ARM_TENANT_ID"),
			Type:        types.StringType,
			Sensitive:   true,
		},
	}),
}
var gcpSchema = tfsdk.Attribute{
	Optional:    true,
	Description: "Credentials for Google Cloud. See how to authenticate through Service Principals in the [Google docs](https://cloud.google.com/compute/docs/authentication)",
	Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
		"credentials": {
			Optional:    true,
			Description: "Either the path to or the contents of a service account key file in JSON format. " + common.HelperValueViaEnvVar("GOOGLE_APPLICATION_CREDENTIALS"),
			Type:        types.StringType,
			Sensitive:   true,
		},
		"project": {
			Optional:    true,
			Description: "The project to manage resources in. " + common.HelperValueViaEnvVar("GOOGLE_CREDENTIALS"),
			Type:        types.StringType,
		},
	}),
}

func (p *Provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Terraform provider to manage the lifecycle of Multy resources.",
		Attributes: map[string]tfsdk.Attribute{
			"api_key": {
				Type:        types.StringType,
				Description: "The Multy API Key necessary to deploy Multy resources. Value can be passed through the `MULTY_API_KEY` environment variable",
				Optional:    true,
				Sensitive:   true,
			},
			"aws":   awsSchema,
			"azure": azureSchema,
			"gcp":   gcpSchema,
			"server_endpoint": {
				Type:        types.StringType,
				Description: "Address of the multy server. Defaults to `api.multy.dev`. If local, it will be run without SSL",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}, nil
}

type providerData struct {
	ApiKey         types.String         `tfsdk:"api_key"`
	ServerEndpoint types.String         `tfsdk:"server_endpoint"`
	Aws            *providerAwsConfig   `tfsdk:"aws"`
	Azure          *providerAzureConfig `tfsdk:"azure"`
	Gcp            *providerGcpConfig   `tfsdk:"gcp"`
}

type providerAwsConfig struct {
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	AccessKeySecret types.String `tfsdk:"access_key_secret"`
	SessionToken    types.String `tfsdk:"session_token"`
}

type providerAzureConfig struct {
	SubscriptionId types.String `tfsdk:"subscription_id"`
	ClientId       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
	TenantId       types.String `tfsdk:"tenant_id"`
}

type providerGcpConfig struct {
	Credentials types.String `tfsdk:"credentials"`
	Project     types.String `tfsdk:"project"`
}

func (p *Provider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	p.ConfigureProvider(ctx, config, resp)
}

func (p *Provider) ConfigureProvider(ctx context.Context, config providerData, resp *provider.ConfigureResponse) {
	var apiKey string
	var err error
	if config.ApiKey.IsUnknown() {
		resp.Diagnostics.AddWarning(
			"Unable to create Client",
			"Cannot use unknown value as api_key",
		)
		return
	}

	if config.ApiKey.IsNull() {
		apiKey = os.Getenv("MULTY_API_KEY")
	} else {
		apiKey = config.ApiKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Unable to find api_key",
			"Api Key cannot be an empty string",
		)
		return
	}

	var awsConfig *common.AwsConfig
	if config.Aws != nil {
		var err error
		awsConfig, err = p.validateAwsConfig(ctx, config.Aws)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to retrieve AWS credentials.",
				err.Error(),
			)
			return
		}
	}

	var azureConfig *common.AzureConfig
	if config.Azure != nil {
		azureConfig, err = p.validateAzureConfig(config.Azure)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to retrieve Azure credentials.",
				err.Error(),
			)
			return
		}
	}

	var gcpConfig *common.GcpConfig
	if config.Gcp != nil {
		gcpConfig, err = p.validateGcpConfig(config.Gcp)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to retrieve Google credentials.",
				err.Error(),
			)
			return
		}
	}

	client := p.getConnToServer(config, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	c := common.ProviderConfig{}
	c.Client = client
	c.ApiKey = apiKey
	c.Aws = awsConfig
	c.Azure = azureConfig
	c.Gcp = gcpConfig
	c.RefreshCache = refreshCache

	ctx, err = c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect to multy server",
			"Unable to connect to multy server:\n\n"+err.Error(),
		)
		return
	}

	p.Client = &c
	p.Configured = true
}

func (p *Provider) getConnToServer(config providerData, resp *provider.ConfigureResponse) proto.MultyResourceServiceClient {
	connCache.Lock()
	defer connCache.Unlock()

	endpoint := "api.multy.dev:443"
	if !config.ServerEndpoint.IsNull() {
		endpoint = config.ServerEndpoint.ValueString()
	}
	if _, ok := connCache.cache[endpoint]; !ok {
		creds := insecure.NewCredentials()
		if !strings.HasPrefix(endpoint, "localhost") {
			cp, err := x509.SystemCertPool()
			if err != nil {
				resp.Diagnostics.AddError(
					"Unable to create multy client",
					"Unable to get system cert pool: "+err.Error(),
				)
				return nil
			}
			creds = credentials.NewClientTLSFromCert(cp, "")
		}

		conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(creds))
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to create Client",
				"Unable to create multy Client:\n\n"+err.Error(),
			)
			return nil
		}

		connCache.cache[endpoint] = proto.NewMultyResourceServiceClient(conn)
	}

	client := connCache.cache[endpoint]
	return client
}

func (p *Provider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return ResourceDatabaseType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceKubernetesClusterType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceKubernetesNodePoolType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceNetworkInterfaceType{}.NewResource(ctx, p) },
		func() resource.Resource {
			return ResourceNetworkInterfaceSecurityGroupAssociationType{}.NewResource(ctx, p)
		},
		func() resource.Resource { return ResourceNetworkSecurityGroupType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceObjectStorageType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceObjectStorageObjectType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourcePublicIpType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceRouteTableType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceRouteTableAssociationType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceSubnetType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceVaultType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceVaultAccessPolicyType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceVaultSecretType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceVirtualMachineType{}.NewResource(ctx, p) },
		func() resource.Resource { return ResourceVirtualNetworkType{}.NewResource(ctx, p) },
	}
}

// GetDataSources - Defines Provider data sources
func (p *Provider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func (p *Provider) validateAwsConfig(ctx context.Context, config *providerAwsConfig) (*common.AwsConfig, error) {
	var awsConfig common.AwsConfig
	if config.AccessKeyId.IsUnknown() {
		return nil, fmt.Errorf("cannot use unknown value as access_key_id")
	}
	if config.AccessKeySecret.IsUnknown() {
		return nil, fmt.Errorf("cannot use unknown value as access_key_secret")
	}
	if config.SessionToken.IsUnknown() {
		return nil, fmt.Errorf("cannot use unknown value as session_token")
	}
	awsConfig = common.AwsConfig{
		AccessKeyId:     config.AccessKeyId.ValueString(),
		AccessKeySecret: config.AccessKeySecret.ValueString(),
		SessionToken:    config.SessionToken.ValueString(),
	}
	if len(awsConfig.AccessKeyId) > 0 && len(awsConfig.AccessKeyId) > 0 {
		return &awsConfig, nil
	}
	awsConfig.AccessKeyId = os.Getenv("AWS_ACCESS_KEY_ID")
	awsConfig.AccessKeySecret = os.Getenv("AWS_SECRET_ACCESS_KEY")
	awsConfig.SessionToken = os.Getenv("AWS_SESSION_TOKEN")
	if len(awsConfig.AccessKeyId) > 0 && len(awsConfig.AccessKeyId) > 0 {
		return &awsConfig, nil
	}

	defaultConfig, err := awscfg.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("aws credentials not set, unable to retrieve default config: %s", err.Error())
	}
	awsCreds, err := defaultConfig.Credentials.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("aws credentials not set, unable to retrieve default config: %s", err.Error())
	}
	awsConfig.AccessKeyId = awsCreds.AccessKeyID
	awsConfig.AccessKeySecret = awsCreds.SecretAccessKey
	return &awsConfig, nil
}

func (p *Provider) validateAzureConfig(config *providerAzureConfig) (*common.AzureConfig, error) {
	var azureConfig common.AzureConfig
	if config.SubscriptionId.IsUnknown() {
		return nil, fmt.Errorf("cannot use unknown value as subscription_id")
	}
	if config.ClientId.IsUnknown() {
		return nil, fmt.Errorf("cannot use unknown value as client_id")
	}
	if config.ClientSecret.IsUnknown() {
		return nil, fmt.Errorf("cannot use unknown value as client_secret")
	}
	if config.TenantId.IsUnknown() {
		return nil, fmt.Errorf("cannot use unknown value as tenant_id")
	}

	azureConfig = common.AzureConfig{
		SubscriptionId: config.SubscriptionId.ValueString(),
		ClientId:       config.ClientId.ValueString(),
		ClientSecret:   config.ClientSecret.ValueString(),
		TenantId:       config.TenantId.ValueString(),
	}

	if azureConfig.SubscriptionId != "" && azureConfig.ClientId != "" && azureConfig.ClientSecret != "" && azureConfig.TenantId != "" {
		return &azureConfig, nil
	}

	azureConfig.SubscriptionId = os.Getenv("ARM_SUBSCRIPTION_ID")
	azureConfig.ClientId = os.Getenv("ARM_CLIENT_ID")
	azureConfig.ClientSecret = os.Getenv("ARM_CLIENT_SECRET")
	azureConfig.TenantId = os.Getenv("ARM_TENANT_ID")

	if azureConfig.SubscriptionId != "" && azureConfig.ClientId != "" && azureConfig.ClientSecret != "" && azureConfig.TenantId != "" {
		return &azureConfig, nil
	}

	// terraform-helper for azure authentication
	azConfig := authentication.Builder{
		//SupportsClientCertAuth:   true,
		SupportsClientSecretAuth: true,
		SupportsAzureCliToken:    true,
	}
	creds, err := azConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate to azure")
	}
	azureConfig.SubscriptionId = creds.SubscriptionID
	azureConfig.ClientId = creds.ClientID
	azureConfig.TenantId = creds.TenantID

	if config != nil && !config.ClientSecret.IsUnknown() && !config.ClientSecret.IsNull() && config.ClientSecret.ValueString() != "" {
		azureConfig.ClientSecret = config.ClientSecret.ValueString()
	} else if os.Getenv("ARM_CLIENT_SECRET") != "" {
		azureConfig.ClientSecret = os.Getenv("ARM_CLIENT_SECRET")
	} else {
		return &azureConfig, fmt.Errorf("ARM_CLIENT_SECRET has not been set")
	}

	if azureConfig.SubscriptionId != "" && azureConfig.ClientId != "" && azureConfig.ClientSecret != "" && azureConfig.TenantId != "" {
		return &azureConfig, nil
	}

	// todo check if access is valid by calling sts.GetCallerIdentity
	return &azureConfig, fmt.Errorf("azure credentials not set")
}

func (p *Provider) validateGcpConfig(config *providerGcpConfig) (*common.GcpConfig, error) {
	var c common.GcpConfig

	if config.Credentials.IsUnknown() {
		return nil, fmt.Errorf("cannot use unknown value as credentials")
	}
	if config.Project.IsUnknown() {
		return nil, fmt.Errorf("cannot use unknown value as project")
	}

	if !config.Project.IsNull() {
		c.Project = config.Project.ValueString()
	} else if project, ok := os.LookupEnv("GOOGLE_PROJECT"); ok {
		c.Project = project
	} else {
		return nil, fmt.Errorf("google project is not set")
	}

	if !config.Credentials.IsNull() {
		contents, _, err := pathOrContents(config.Credentials.ValueString())
		if err != nil {
			return nil, err
		}
		c.Credentials = contents
	} else if creds, ok := os.LookupEnv("GOOGLE_CREDENTIALS"); ok {
		c.Credentials = creds
	} else if credsFile, ok := os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS"); ok {
		contents, fromFile, err := pathOrContents(credsFile)
		if err != nil {
			return nil, err
		}
		if !fromFile {
			return nil, fmt.Errorf(
				"GOOGLE_APPLICATION_CREDENTIALS should contain a path to the JSON file, but the content was found" +
					" instead. Did you mean to use GOOGLE_CREDENTIALS?")
		}
		c.Credentials = contents
	} else {
		return nil, fmt.Errorf("google credentials not set")
	}

	return &c, nil
}

func pathOrContents(pathOrContent string) (string, bool, error) {
	if len(pathOrContent) == 0 {
		return pathOrContent, false, nil
	}

	path := pathOrContent
	if path[0] == '~' {
		var err error
		path, err = homedir.Expand(path)
		if err != nil {
			return path, true, err
		}
	}

	if _, err := os.Stat(path); err == nil {
		contents, err := ioutil.ReadFile(path)
		if err != nil {
			return string(contents), true, err
		}
		return string(contents), true, nil
	}

	return pathOrContent, false, nil
}
