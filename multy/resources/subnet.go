package resources

import (
	"context"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-multy/multy/common"
)

func Subnet() *schema.Resource {
	return &schema.Resource{
		CreateContext: subnetCreate,
		ReadContext:   subnetRead,
		UpdateContext: subnetUpdate,
		DeleteContext: subnetDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"cidr_block": &schema.Schema{
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: common.ValidateCidr,
			},
			"virtual_network_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func subnetCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	id, _ := uuid.GenerateUUID()

	d.SetId(id)

	subnetRead(ctx, d, m)

	return diags
}

func subnetRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func subnetUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return subnetRead(ctx, d, m)
}

func subnetDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
