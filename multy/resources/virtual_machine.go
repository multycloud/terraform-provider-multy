package resources

import (
	"context"
	"github.com/hashicorp/go-uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"terraform-provider-multy/multy/common"
)

var (
	operatingSystem = []string{"linux"}
	size            = []string{"micro", "small"}
)

func VirtualMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: virtualMachineCreate,
		ReadContext:   virtualMachineRead,
		UpdateContext: virtualMachineUpdate,
		DeleteContext: virtualMachineDelete,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"operating_system": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(operatingSystem, true),
			},
			"size": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(size, true),
			},
			"subnet_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"network_interface_ids": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"network_security_group_ids": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"user_data": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssh_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"public_ip_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"public_ip": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"clouds":  common.CloudsSchema,
			"rg_vars": common.RgVarsSchema,
		},
	}
}

func virtualMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	pIpId := d.Get("public_ip_id").(string)
	pIp := d.Get("public_ip").(bool)

	// fixme check isnt working
	if pIp == true && pIpId != "" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "conflict between public_ip and public_ip_id",
			Detail:   "cannot set both public_ip and public_ip_id",
		})

		return diags
	}

	id, _ := uuid.GenerateUUID()
	d.SetId(id)

	virtualMachineRead(ctx, d, m)

	return diags
}

func virtualMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}

func virtualMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return virtualMachineRead(ctx, d, m)
}

func virtualMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
}
