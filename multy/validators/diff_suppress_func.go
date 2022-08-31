package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

type DiffSuppressFunc[T any] struct {
	isEqual func(T, T) bool
}

var IgnoringWhitespace = NewDiffSuppressFunc(func(val1 string, val2 string) bool {
	return strings.TrimSpace(val1) == strings.TrimSpace(val2)
})

func NewDiffSuppressFunc[T any](isEqual func(T, T) bool) DiffSuppressFunc[T] {
	return DiffSuppressFunc[T]{isEqual: isEqual}
}

func (m DiffSuppressFunc[T]) Description(_ context.Context) string {
	return fmt.Sprintf("Ignores diff based on a custom function.")
}

func (m DiffSuppressFunc[T]) MarkdownDescription(c context.Context) string {
	return m.Description(c)
}

func (m DiffSuppressFunc[T]) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeConfig == nil || req.AttributeConfig.IsNull() || req.AttributeConfig.IsUnknown() {
		return
	}
	if req.AttributeState == nil || req.AttributeState.IsNull() || req.AttributeState.IsUnknown() {
		return
	}

	if req.AttributeConfig.Equal(req.AttributeState) {
		return
	}

	stateVal, err := req.AttributeState.ToTerraformValue(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("DiffSupressFunc: Unable to convert state value to tf value, %s", err))
		return
	}

	configVal, err := req.AttributeConfig.ToTerraformValue(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("DiffSupressFunc: Unable to convert config value to tf value, %s", err))
		return
	}

	var c, s T

	err = stateVal.As(&s)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("DiffSupressFunc: Unable to convert state value, %s", err))
		return
	}

	err = configVal.As(&c)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("DiffSupressFunc: Unable to convert config value, %s", err))
		return
	}

	if m.isEqual(c, s) {
		tflog.Info(ctx, fmt.Sprintf("DiffSupressFunc: Diff has been suprressed."))
		resp.AttributePlan = req.AttributeState
	}
}
