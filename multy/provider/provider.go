package provider

import (
	"context"
	"fmt"
	"terraform-provider-multy/multy/resources"

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
		},
		ResourcesMap: map[string]*schema.Resource{
			"multy_virtual_network": resources.VirtualMachine(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	//c, err := hashicups.NewClient(apiKey)
	//if err != nil {
	//	diags = append(diags, diag.Diagnostic{
	//		Severity: diag.Error,
	//		Summary:  "Unable to create HashiCups client",
	//		Detail:   "Unable to create anonymous HashiCups client",
	//	})
	//	return nil, diags
	//}

	//return c, diags
	fmt.Sprintf("LOGIN: %s", apiKey)
	return nil, diags
}
