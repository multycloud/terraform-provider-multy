package data

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/multycloud/multy/api/proto/resources"
	"strings"
	"terraform-provider-multy/multy/common"
)

func DataVirtualNetwork() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataVirtualNetworkRead,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataVirtualNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
