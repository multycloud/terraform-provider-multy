package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"terraform-provider-multy/multy/mtypes"
)

type EnumValidator[T mtypes.ProtoEnum] struct {
	Typ mtypes.EnumType[T]
}

func (v EnumValidator[T]) Description(_ context.Context) string {
	return fmt.Sprintf("string value must be one of %v", v.Typ.GetAllValues())
}

func (v EnumValidator[T]) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("string value must be one of %v", v.Typ.GetAllValues())
}

func (v EnumValidator[T]) Validate(ctx context.Context, req tfsdk.ValidateAttributeRequest, resp *tfsdk.ValidateAttributeResponse) {
	val := v.Typ.ZeroVal()
	diags := tfsdk.ValueAs(ctx, req.AttributeConfig, &val)
	val.Typ = v.Typ
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := val.Validate()
	if err != nil {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid value",
			err.Error(),
		)
		return
	}
}

func NewValidator[T mtypes.ProtoEnum](n mtypes.EnumType[T]) tfsdk.AttributeValidator {
	return EnumValidator[T]{
		Typ: n,
	}
}
