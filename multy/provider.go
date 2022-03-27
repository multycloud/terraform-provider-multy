package multy

import (
	"context"
	"crypto/x509"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"strings"
	"terraform-provider-multy/multy/common"
)

func New() tfsdk.Provider {
	return &Provider{}
}

type Provider struct {
	Configured bool
	Client     *common.ProviderConfig
}

func (p *Provider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"api_key": {
				Type:        types.StringType,
				Description: "The Multy API Key necessary to deploy Multy resources. Value can be passed through the `MULTY_API_KEY` environment variable",
				Optional:    true,
				Sensitive:   true,
			},
			"aws": {
				Optional:    true,
				Description: "Credentials for AWS Cloud",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"access_key_id": {
						Required:    true,
						Description: "AWS Access Key ID. " + common.HelperValueViaEnvVar("AWS_ACCESS_KEY_ID"),
						Type:        types.StringType,
						Sensitive:   true,
					},
					"secret_access_key": {
						Required:    true,
						Description: "AWS Secret Access Key. " + common.HelperValueViaEnvVar("AWS_SECRET_ACCESS_KEY"),
						Type:        types.StringType,
						Sensitive:   true,
					},
				}),
			},
			"azure": {
				Optional:    true,
				Description: "Credentials for Azure Cloud. See how to authenticate through Service Principal in the [Azure docs](https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/guides/service_principal_client_secret#creating-a-service-principal)",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"subscription_id": {
						Required:    true,
						Description: "Azure Subscription ID. " + common.HelperValueViaEnvVar("ARM_SUBSCRIPTION_ID"),
						Type:        types.StringType,
						Sensitive:   true,
					},
					"client_id": {
						Required:    true,
						Description: "Azure Client ID " + common.HelperValueViaEnvVar("ARM_CLIENT_ID"),
						Type:        types.StringType,
						Sensitive:   true,
					},
					"client_secret": {
						Required:    true,
						Description: "Azure Client Secret " + common.HelperValueViaEnvVar("ARM_CLIENT_SECRET"),
						Type:        types.StringType,
						Sensitive:   true,
					},
					"tenant_id": {
						Required:    true,
						Description: "Azure Tenant ID " + common.HelperValueViaEnvVar("ARM_TENANT_ID"),
						Type:        types.StringType,
						Sensitive:   true,
					},
				}),
			},
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
	ApiKey         types.String        `tfsdk:"api_key"`
	ServerEndpoint types.String        `tfsdk:"server_endpoint"`
	Aws            providerAwsConfig   `tfsdk:"aws"`
	Azure          providerAzureConfig `tfsdk:"azure"`
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

	var apiKey string
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

	awsConfig, err := p.validateAwsConfig(config.Aws)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect to AWS",
			err.Error(),
		)
		return
	}

	azureConfig, err := p.validateAzureConfig(config.Azure)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to connect to AWS",
			err.Error(),
		)
		return
	}

	endpoint := "api.multy.dev:443"
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
	// fix me no err from previous func
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Client",
			"Unable to create multy Client:\n\n"+err.Error(),
		)
		return
	}

	c.Client = client
	c.ApiKey = apiKey
	c.Aws = *awsConfig
	c.Azure = *azureConfig

	p.Client = &c
	p.Configured = true
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
	}, nil
}

// GetDataSources - Defines Provider data sources
func (p *Provider) GetDataSources(_ context.Context) (map[string]tfsdk.DataSourceType, diag.Diagnostics) {
	return map[string]tfsdk.DataSourceType{
		//"multy_virtual_network": data.DataVirtualNetwork(),
	}, nil
}

func (p *Provider) validateAwsConfig(config providerAwsConfig) (*common.AwsConfig, error) {
	var awsConfig common.AwsConfig
	if config.AccessKeyId.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as access_key_id")
	}
	if config.AccessKeyId.Null {
		awsConfig.AccessKeyId = os.Getenv("AWS_ACCESS_KEY_ID")
	} else {
		awsConfig.AccessKeyId = config.AccessKeyId.Value
	}
	if awsConfig.AccessKeyId == "" {
		return nil, fmt.Errorf("access_key_id cannot be an empty string")
	}

	if config.AccessKeySecret.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as access_key_id")
	}
	if config.AccessKeySecret.Null {
		awsConfig.AccessKeySecret = os.Getenv("AWS_SECRET_ACCESS_KEY")
	} else {
		awsConfig.AccessKeySecret = config.AccessKeySecret.Value
	}
	if awsConfig.AccessKeySecret == "" {
		return nil, fmt.Errorf("access_secret_key cannot be an empty string")
	}

	// todo check if access is valid by calling sts.GetCallerIdentity
	return &awsConfig, nil
}

func (p *Provider) validateAzureConfig(config providerAzureConfig) (*common.AzureConfig, error) {
	var azureConfig common.AzureConfig
	if config.SubscriptionId.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as subscription_id")
	}
	if config.SubscriptionId.Null {
		azureConfig.SubscriptionId = os.Getenv("ARM_SUBSCRIPTION_ID")
	} else {
		azureConfig.SubscriptionId = config.SubscriptionId.Value
	}
	if azureConfig.SubscriptionId == "" {
		return nil, fmt.Errorf("subscription_id cannot be an empty string")
	}

	if config.ClientId.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as client_id")
	}
	if config.ClientId.Null {
		azureConfig.ClientId = os.Getenv("ARM_CLIENT_ID")
	} else {
		azureConfig.ClientId = config.ClientId.Value
	}
	if azureConfig.ClientId == "" {
		return nil, fmt.Errorf("client_id cannot be an empty string")
	}

	if config.ClientSecret.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as client_secret")
	}
	if config.ClientSecret.Null {
		azureConfig.ClientSecret = os.Getenv("ARM_CLIENT_SECRET")
	} else {
		azureConfig.ClientSecret = config.ClientSecret.Value
	}
	if azureConfig.ClientSecret == "" {
		return nil, fmt.Errorf("client_secret cannot be an empty string")
	}

	if config.TenantId.Unknown {
		return nil, fmt.Errorf("cannot use unknown value as tenant_id")
	}
	if config.TenantId.Null {
		azureConfig.TenantId = os.Getenv("ARM_TENANT_ID")
	} else {
		azureConfig.TenantId = config.TenantId.Value
	}
	if azureConfig.TenantId == "" {
		return nil, fmt.Errorf("tenant_id cannot be an empty string")
	}

	// todo check if access is valid by calling sts.GetCallerIdentity
	return &azureConfig, nil
}
