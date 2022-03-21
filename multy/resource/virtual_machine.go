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

func VirtualMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: virtualMachineCreate,
		ReadContext:   virtualMachineRead,
		UpdateContext: virtualMachineUpdate,
		DeleteContext: virtualMachineDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"operating_system": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(common.GetVmOperatingSystem(), true),
			},
			"size": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(common.GetVmSize(), true),
			},
			"subnet_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"network_interface_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"network_security_group_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssh_key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"public_ip_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"public_ip": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"clouds":   common.CloudsSchema,
			"location": common.LocationSchema,
		},
	}
}

func virtualMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)
	clouds := c.GetClouds(d)

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

	vmResources := vmConvert(clouds, d, c)

	vm, err := c.Client.CreateVirtualMachine(ctx, &resources.CreateVirtualMachineRequest{
		Resources: vmResources,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(vm.CommonParameters.ResourceId)
	return nil
}

func virtualMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	vm, err := c.Client.ReadVirtualMachine(ctx, &resources.ReadVirtualMachineRequest{ResourceId: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

	for _, cloudR := range vm.Resources {
		err = d.Set(strings.ToLower(cloudR.CommonParameters.CloudProvider.String()), map[string]any{
			"name":                       cloudR.Name,
			"operating_system":           cloudR.OperatingSystem,
			"network_interface_ids":      cloudR.NetworkInterfaceIds,
			"network_security_group_ids": cloudR.NetworkSecurityGroupIds,
			"size":                       cloudR.VmSize,
			"user_data":                  cloudR.UserData,
			"subnet_id":                  cloudR.SubnetId,
			"ssh_key":                    cloudR.PublicSshKey,
			"public_ip_id":               cloudR.PublicIpId,
			"public_ip":                  cloudR.GeneratePublicIp,
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func virtualMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	clouds := c.GetClouds(d)

	vmResources := vmConvert(clouds, d, c)

	vm, err := c.Client.UpdateVirtualMachine(ctx, &resources.UpdateVirtualMachineRequest{
		ResourceId: d.Id(),
		Resources:  vmResources,
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(vm.CommonParameters.ResourceId)
	return nil
}

func virtualMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c := m.(*common.ProviderConfig)
	ctx = c.AddHeaders(ctx)

	_, err := c.Client.DeleteVirtualMachine(ctx, &resources.DeleteVirtualMachineRequest{ResourceId: d.Id()})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func vmConvert(clouds []common_proto.CloudProvider, d *schema.ResourceData, c *common.ProviderConfig) []*resources.CloudSpecificVirtualMachineArgs {
	var vmResources []*resources.CloudSpecificVirtualMachineArgs
	for _, cloud := range clouds {
		vmResources = append(vmResources, &resources.CloudSpecificVirtualMachineArgs{
			CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
				Location:      c.GetLocation(d),
				CloudProvider: cloud,
			},
			Name:                    d.Get("name").(string),
			OperatingSystem:         c.GetOperatingSystem(d),
			NetworkInterfaceIds:     common.InterfaceToStringMap(d.Get("network_interface_ids").([]interface{})),
			NetworkSecurityGroupIds: common.InterfaceToStringMap(d.Get("network_security_group_ids").([]interface{})),
			VmSize:                  c.GetVmSize(d),
			UserData:                d.Get("user_data").(string),
			SubnetId:                d.Get("subnet_id").(string),
			PublicSshKey:            d.Get("ssh_key").(string),
			PublicIpId:              d.Get("public_ip_id").(string),
			GeneratePublicIp:        d.Get("public_ip").(bool),
		})
	}
	return vmResources
}
