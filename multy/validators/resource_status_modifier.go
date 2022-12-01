package validators

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ResourceStatusModifier struct {
}

func (m ResourceStatusModifier) Description(_ context.Context) string {
	return fmt.Sprintf("Sets value to always be null, as resource status should be empty when no drift is detected.")
}

func (m ResourceStatusModifier) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("Sets value to always be null, as resource status should be empty when no drift is detected.")
}

func (m ResourceStatusModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	resp.AttributePlan = types.MapNull(types.StringType)
}
