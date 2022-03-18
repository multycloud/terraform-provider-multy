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

func Subnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: subnetCreate,
		ReadContext:   subnetRead,
		UpdateContext: subnetUpdate,
		DeleteContext: subnetDelete,
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
			"virtual_network_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"clouds": common.CloudsSchema,
		},
	}
}

func subnetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	clouds := c.GetClouds(d)

	var subnetResources []*resources.CloudSpecificSubnetArgs
	for _, cloud := range clouds {
		subnetResources = append(subnetResources, &resources.CloudSpecificSubnetArgs{
			CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
				//ResourceGroupId: "subnet-rg",
				Location:      c.GetLocation(d),
				CloudProvider: cloud,
			},
			Name:             d.Get("name").(string),
			CidrBlock:        d.Get("cidr_block").(string),
			VirtualNetworkId: d.Get("virtual_network_id").(string),
		})
	}

	subnet, err := c.Client.CreateSubnet(ctx, &resources.CreateSubnetRequest{
		Resources: subnetResources,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(subnet.CommonParameters.ResourceId)
	return nil
}

func subnetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	vn, err := c.Client.ReadSubnet(ctx, &resources.ReadSubnetRequest{ResourceId: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}
	for _, cloudR := range vn.Resources {
		err = d.Set(strings.ToLower(cloudR.CommonParameters.CloudProvider.String()), map[string]any{
			"name":               cloudR.Name,
			"cidr_block":         cloudR.CidrBlock,
			"virtual_network_id": cloudR.VirtualNetworkId,
		})

		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func subnetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	clouds := c.GetClouds(d)

	var subnetResources []*resources.CloudSpecificSubnetArgs
	for _, cloud := range clouds {
		subnetResources = append(subnetResources, &resources.CloudSpecificSubnetArgs{
			CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
				//ResourceGroupId: "subnet-rg",
				Location:      c.GetLocation(d),
				CloudProvider: cloud,
			},
			Name:             d.Get("name").(string),
			CidrBlock:        d.Get("cidr_block").(string),
			VirtualNetworkId: d.Get("virtual_network_id").(string),
		})
	}

	subnet, err := c.Client.UpdateSubnet(ctx, &resources.UpdateSubnetRequest{
		ResourceId: d.Id(),
		Resources:  subnetResources,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(subnet.CommonParameters.ResourceId)
	return nil
}

func subnetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	_, err := c.Client.DeleteSubnet(ctx, &resources.DeleteSubnetRequest{ResourceId: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
