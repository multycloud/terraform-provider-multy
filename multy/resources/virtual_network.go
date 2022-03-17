package resources

import (
	"context"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-multy/multy/common"
)

func VirtualNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: virtualNetworkCreate,
		ReadContext:   virtualNetworkRead,
		UpdateContext: virtualNetworkUpdate,
		DeleteContext: virtualNetworkDelete,
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
		},
	}
}

func virtualNetworkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	id, _ := uuid.GenerateUUID()

	d.SetId(id)

	virtualNetworkRead(ctx, d, m)

	return diags
}

func virtualNetworkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func virtualNetworkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return virtualNetworkRead(ctx, d, m)
}

func virtualNetworkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
