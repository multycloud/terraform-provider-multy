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
			"cidr_block": &schema.Schema{
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
			"clouds":  common.CloudsSchema,
			"rg_vars": common.RgVarsSchema,
		},
	}
}

func virtualNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	vn, err := c.Client.CreateVirtualNetwork(ctx, &resources.CreateVirtualNetworkRequest{
		Resources: []*resources.CloudSpecificCreateVirtualNetworkRequest{
			{
				CommonParameters: &common_proto.CloudSpecificCreateResourceCommonParameters{
					ResourceGroupId: "vn-rg",
					Location:        common_proto.Location_IRELAND,
					CloudProvider:   common_proto.CloudProvider_AWS,
				},
				Name:      d.Get("name").(string),
				CidrBlock: d.Get("cidr_block").(string),
			},
		},
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
			"cidr_block": cloudR.Name,
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

	_, err := c.Client.UpdateVirtualNetwork(ctx, &resources.UpdateVirtualNetworkRequest{ResourceId: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

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
