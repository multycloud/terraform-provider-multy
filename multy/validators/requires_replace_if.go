package validators

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-log/tfsdklog"
)

func RequiresReplaceIf(f RequiresReplaceIfFunc, description, markdownDescription string) tfsdk.AttributePlanModifier {
	return RequiresReplaceIfModifier{
		f:                   f,
		description:         description,
		markdownDescription: markdownDescription,
	}
}

// RequiresReplaceIfFunc is a conditional function used in the RequiresReplaceIf
// plan modifier to determine whether the attribute requires replacement.
type RequiresReplaceIfFunc func(ctx context.Context, state, config attr.Value, plan tfsdk.Plan) (bool, diag.Diagnostics)

// RequiresReplaceIfModifier is an AttributePlanModifier that sets RequiresReplace
// on the attribute if the conditional function returns true.
type RequiresReplaceIfModifier struct {
	f                   RequiresReplaceIfFunc
	description         string
	markdownDescription string
}

func (r RequiresReplaceIfModifier) Modify(ctx context.Context, req tfsdk.ModifyAttributePlanRequest, resp *tfsdk.ModifyAttributePlanResponse) {
	if req.AttributeConfig == nil || req.AttributePlan == nil || req.AttributeState == nil {
		// shouldn't happen, but let's not panic if it does
		return
	}

	if req.State.Raw.IsNull() {
		// if we're creating the resource, no need to delete and
		// recreate it
		return
	}

	if req.Plan.Raw.IsNull() {
		// if we're deleting the resource, no need to delete and
		// recreate it
		return
	}

	if req.AttributePlan.Equal(req.AttributeState) {
		// if the plan and the state are in agreement, this attribute
		// isn't changing, don't require replace
		return
	}

	res, diags := r.f(ctx, req.AttributeState, req.AttributeConfig, req.Plan)
	resp.Diagnostics.Append(diags...)

	// If the function says to require replacing, we require replacing.
	// If the function says not to, we don't change the value that prior
	// plan modifiers may have set.
	if res {
		resp.RequiresReplace = true
	} else if resp.RequiresReplace {
		tfsdklog.Debug(ctx, "Keeping previous attribute replacement requirement", map[string]interface{}{"attribute_path": req.AttributePath.String()})
	}
}

// Description returns a human-readable description of the plan modifier.
func (r RequiresReplaceIfModifier) Description(ctx context.Context) string {
	return r.description
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (r RequiresReplaceIfModifier) MarkdownDescription(ctx context.Context) string {
	return r.markdownDescription
}
