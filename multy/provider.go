package multy

import (
	"context"
	"crypto/x509"
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
				Description: "The Multy API Key necessary to deploy Multy resources. Value can be passed through the `MULTY_API_KEY` environmnet variable",
				Optional:    true,
				Sensitive:   true,
			},
			"server_endpoint": {
				Type:        types.StringType,
				Description: "Address of the multy server. If local, it will be run without SSL.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}, nil
}

type providerData struct {
	ApiKey         types.String `tfsdk:"api_key"`
	ServerEndpoint types.String `tfsdk:"server_endpoint"`
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
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create Client",
			"Unable to create multy Client:\n\n"+err.Error(),
		)
		return
	}

	c.Client = client
	c.ApiKey = apiKey

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
