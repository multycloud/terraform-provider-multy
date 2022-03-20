package resource

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	common_proto "github.com/multycloud/multy/api/proto/common"
	"github.com/multycloud/multy/api/proto/resources"
	"strings"
	"terraform-provider-multy/multy/common"
)

func VirtualNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: virtualNetworkCreate,
		ReadContext:   virtualNetworkRead,
		UpdateContext: virtualNetworkUpdate,
		DeleteContext: virtualNetworkDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsCIDR,
			},
			"aws": {
				Type:     schema.TypeMap,
				Computed: true,
			},

			"azure": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"clouds":   common.CloudsSchema,
			"location": common.LocationSchema,
		},
	}
}

func virtualNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	clouds := c.GetClouds(d)

	var vnResources []*resources.CloudSpecificVirtualNetworkArgs
	for _, cloud := range clouds {
		vnResources = append(vnResources, &resources.CloudSpecificVirtualNetworkArgs{
			CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
				//ResourceGroupId: "vn-rg",
				Location:      c.GetLocation(d),
				CloudProvider: cloud,
			},
			Name:      d.Get("name").(string),
			CidrBlock: d.Get("cidr_block").(string),
		})
	}

	vn, err := c.Client.CreateVirtualNetwork(ctx, &resources.CreateVirtualNetworkRequest{
		Resources: vnResources,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(vn.CommonParameters.ResourceId)
	return nil
}

func virtualNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	vn, err := c.Client.ReadVirtualNetwork(ctx, &resources.ReadVirtualNetworkRequest{ResourceId: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}
	for _, cloudR := range vn.Resources {
		err = d.Set(strings.ToLower(cloudR.CommonParameters.CloudProvider.String()), map[string]any{
			"name":       cloudR.Name,
			"cidr_block": cloudR.CidrBlock,
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func virtualNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	clouds := c.GetClouds(d)
	var vnResources []*resources.CloudSpecificVirtualNetworkArgs
	for _, cloud := range clouds {
		vnResources = append(vnResources, &resources.CloudSpecificVirtualNetworkArgs{
			CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
				Location:      c.GetLocation(d),
				CloudProvider: cloud,
			},
			Name:      d.Get("name").(string),
			CidrBlock: d.Get("cidr_block").(string),
		})
	}

	vn, err := c.Client.UpdateVirtualNetwork(ctx, &resources.UpdateVirtualNetworkRequest{
		ResourceId: d.Id(),
		Resources:  vnResources,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(vn.CommonParameters.ResourceId)
	return virtualNetworkRead(ctx, d, m)
}

func virtualNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	_, err := c.Client.DeleteVirtualNetwork(ctx, &resources.DeleteVirtualNetworkRequest{ResourceId: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
