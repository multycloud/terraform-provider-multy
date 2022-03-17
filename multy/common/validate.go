package common

import (
	"fmt"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"net"
)

func ValidateCidr(v interface{}, p cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	value := v.(string)
	_, _, err := net.ParseCIDR(value)

	if err != nil {
		diag := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "invalid cidr block",
			Detail:   fmt.Sprintf("%s is not a valid %+v", value, p),
		}
		diags = append(diags, diag)
	}
	return diags
}
