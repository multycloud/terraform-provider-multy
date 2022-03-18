package resource

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"strconv"
	"terraform-provider-multy/multy/common"
)

var (
	ruleDirections = []string{"ingress", "egress", "both"}
	ruleProtocols  = []string{"tcp"}
)

func NetworkSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: networkSecurityGroupCreate,
		ReadContext:   networkSecurityGroupRead,
		UpdateContext: networkSecurityGroupUpdate,
		DeleteContext: networkSecurityGroupDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"virtual_network_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"rule": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ruleProtocols, true),
						},
						"priority": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
						"from_port": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateRulePort,
						},
						"to_port": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateRulePort,
						},
						"cidr_block": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
						"direction": &schema.Schema{
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(ruleDirections, true),
						},
						"clouds": common.CloudsSchema,
					},
				},
			},
		},
	}
}

func networkSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	id, _ := uuid.GenerateUUID()

	d.SetId(id)

	networkSecurityGroupRead(ctx, d, m)

	return diags
}

func networkSecurityGroupRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func networkSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return networkSecurityGroupRead(ctx, d, m)
}

func networkSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func validateRulePort(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	if i, err := strconv.Atoi(v); err != nil || i < -1 {
		errors = append(errors, fmt.Errorf("expected %s to be between greater than -1, got %s", k, v))
	}
	return warnings, errors
}
