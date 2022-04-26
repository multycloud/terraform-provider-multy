package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type StringInSliceValidator struct {
	Values []string
}

func (v StringInSliceValidator) Description(_ context.Context) string {
	return fmt.Sprintf("string value must be one of %v", v.Values)
}

func (v StringInSliceValidator) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("string value must be one of %v", v.Values)
}
func (v StringInSliceValidator) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	var str types.String
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &str)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if str.Unknown || str.Null {
		return
	}

	for _, v := range v.Values {
		if v == str.Value {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.AttributePath,
		"Invalid value",
		fmt.Sprintf("expected %s to be one of %v", v.Values, str.Value),
	)
}
