package multy

import (
	"context"
	"crypto/x509"
	"fmt"
	"github.com/multycloud/multy/api/proto/commonpb"
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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

func New() tfsdk.Provider {
	return &Provider{}
}

type Provider struct {
	Configured  bool
	Client      *common.ProviderConfig
	refreshCall *sync.Once
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
}

type providerAwsConfig struct {
	AccessKeyId     types.String `tfsdk:"access_key_id"`
	AccessKeySecret types.String `tfsdk:"access_key_secret"`
}

type providerAzureConfig struct {
	SubscriptionId types.String `tfsdk:"subscription_id"`
	ClientId       types.String `tfsdk:"client_id"`
	ClientSecret   types.String `tfsdk:"client_secret"`
	TenantId       types.String `tfsdk:"tenant_id"`
}

func (p *Provider) Configure(ctx context.Context, req tfsdk.ConfigureProviderRequest, resp *tfsdk.ConfigureProviderResponse) {
	var config providerData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	p.ConfigureProvider(ctx, config, resp)
}

func (p *Provider) ConfigureProvider(ctx context.Context, config providerData, resp *tfsdk.ConfigureProviderResponse) {
	var apiKey string
	var err error
	if config.ApiKey.Unknown {
		resp.Diagnostics.AddWarning(
			"Unable to create Client",
			"Cannot use unknown value as api_key",
		)
		return
	}

	if config.ApiKey.Null {
		apiKey = os.Getenv("MULTY_API_KEY")
	} else {
		apiKey = config.ApiKey.Value
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
				"Unable to connect to AWS",
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
				"Unable to connect to Azure",
				err.Error(),
			)
			return
		}

	}
	endpoint := "api2.multy.dev:443"
	if !config.ServerEndpoint.Null {
		endpoint = config.ServerEndpoint.Value
	}

	creds := insecure.NewCredentials()
	if !strings.HasPrefix(endpoint, "localhost") {
		cp, err := x509.SystemCertPool()
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to create multy client",
				"Unable to get system cert pool: "+err.Error(),
			)
			return
		}
		creds = credentials.NewClientTLSFromCert(cp, "")
	}

	conn, err := grpc.Dial(endpoint, grpc.WithTransportCredentials(creds))
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Client",
			"Unable to create multy Client:\n\n"+err.Error(),
		)
		return
	}

	c := common.ProviderConfig{}
	client := proto.NewMultyResourceServiceClient(conn)

	c.Client = client
	c.ApiKey = apiKey
	c.Aws = awsConfig
	c.Azure = azureConfig

	ctx, err = c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect to multy server",
			"Unable to connect to multy server:\n\n"+err.Error(),
		)
		return
	}

	p.refreshCall = &sync.Once{}
	p.Client = &c
	p.Configured = true
}

func (p *Provider) Refresh(ctx context.Context, diags diag.Diagnostics) {
	p.refreshCall.Do(func() {
		_, err := p.Client.Client.RefreshState(ctx, &commonpb.Empty{})
		if err != nil {
			diags.AddError(
				"Unable to connect to multy server",
				"Unable to connect to multy server:\n\n"+common.ParseGrpcErrors(err),
			)
		}
	})
}

func (p *Provider) GetResources(_ context.Context) (map[string]tfsdk.ResourceType, diag.Diagnostics) {
	return map[string]tfsdk.ResourceType{
		"multy_virtual_network":         ResourceVirtualNetworkType{},
		"multy_subnet":                  ResourceSubnetType{},
		"multy_virtual_machine":         ResourceVirtualMachineType{},
		"multy_network_security_group":  ResourceNetworkSecurityGroupType{},
		"multy_network_interface":       ResourceNetworkInterfaceType{},
		"multy_route_table":             ResourceRouteTableType{},
		"multy_route_table_association": ResourceRouteTableAssociationType{},
		"multy_object_storage_object":   ResourceObjectStorageObjectType{},
		"multy_object_storage":          ResourceObjectStorageType{},
		"multy_database":                ResourceDatabaseType{},
		"multy_vault":                   ResourceVaultType{},
		"multy_vault_secret":            ResourceVaultSecretType{},
		"multy_vault_access_policy":     ResourceVaultAccessPolicyType{},
	}, nil
}

// GetDataSources - Defines Provider data sources
func (p *Provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		//"multy_virtual_network": data.DataVirtualNetwork(),
	}, nil
}

func (p *Provider) validateAwsConfig(ctx context.Context, config *providerAwsConfig) (*common.AwsConfig, error) {
	var awsConfig common.AwsConfig
	if config.AccessKeyId.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as access_key_id")
	}
	if config.AccessKeySecret.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as access_key_seceret")
	}
	awsConfig = common.AwsConfig{
		AccessKeyId:     config.AccessKeyId.Value,
		AccessKeySecret: config.AccessKeySecret.Value,
	}
	if len(awsConfig.AccessKeyId) > 0 && len(awsConfig.AccessKeyId) > 0 {
		return &awsConfig, nil
	}
	awsConfig.AccessKeyId = os.Getenv("AWS_ACCESS_KEY_ID")
	awsConfig.AccessKeySecret = os.Getenv("AWS_SECRET_ACCESS_KEY")
	if awsConfig.AccessKeyId != "" && awsConfig.AccessKeySecret != "" {
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
	if config.SubscriptionId.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as subscription_id")
	}
	if config.ClientId.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as client_id")
	}
	if config.ClientSecret.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as client_secret")
	}
	if config.TenantId.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as tenant_id")
	}

	azureConfig = common.AzureConfig{
		SubscriptionId: config.SubscriptionId.Value,
		ClientId:       config.ClientId.Value,
		ClientSecret:   config.ClientSecret.Value,
		TenantId:       config.TenantId.Value,
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

	if config != nil && !config.ClientSecret.Unknown && !config.ClientSecret.Null && config.ClientSecret.Value != "" {
		azureConfig.ClientSecret = config.ClientSecret.Value
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
