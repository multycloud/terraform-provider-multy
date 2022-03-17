package provider

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/data"
	"terraform-provider-multy/multy/resource"

	"github.com/multycloud/multy/api/proto"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"clouds": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				//ValidateFunc: validation.StringInSlice(common.Clouds, true),
			},
			"location": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				//ValidateFunc: validation.StringInSlice(common.Locations, true),
			},
			"default_resource_group_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"multy_virtual_network":        resource.VirtualNetwork(),
			"multy_subnet":                 resource.Subnet(),
			"multy_virtual_machine":        resource.VirtualMachine(),
			"multy_network_security_group": resource.NetworkSecurityGroup(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"multy_virtual_network": data.DataVirtualNetwork(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	c := common.ProviderConfig{}
	apiKey := d.Get("api_key").(string)

	conn, err := grpc.Dial("localhost:8000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	client := proto.NewMultyResourceServiceClient(conn)
	c.Client = client
	c.ApiKey = apiKey
	return &c, nil
}
