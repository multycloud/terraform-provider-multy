package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"net"
)

type IsCidrValidator struct {
}

func (v IsCidrValidator) Description(_ context.Context) string {
	return "string value must be a valid CIDR"
}

func (v IsCidrValidator) MarkdownDescription(_ context.Context) string {
	return "string value must be a valid CIDR"
}

func (v IsCidrValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if str.Unknown || str.Null {
		return
	}

	if _, _, err := net.ParseCIDR(str.Value); err == nil {
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.AttributePath,
		"Invalid value",
		fmt.Sprintf("%s is not a valid CIDR", str.Value),
	)
}
