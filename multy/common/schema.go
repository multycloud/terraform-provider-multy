package common

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"golang.org/x/exp/slices"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

//var RgVarsSchema = &schema.Schema{
//	Type:     schema.TypeMap,
//	Optional: true,
//	Elem: &schema.Schema{
//		Type: schema.TypeString,
//	},
//}

var CloudsSchema = tfsdk.Attribute{
	Type:          mtypes.CloudType,
	Description:   fmt.Sprintf("Cloud provider to deploy resource into. Accepted values are %s", StringSliceToDocsMarkdown(mtypes.CloudType.GetAllValues())),
	Required:      true,
	Validators:    []tfsdk.AttributeValidator{validators.NewValidator(mtypes.CloudType)},
	PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
}

var LocationSchema = tfsdk.Attribute{
	Type:          mtypes.LocationType,
	Description:   "Location to deploy resource into. Read more about regions in [documentation](https://docs.multy.dev/regions)",
	Required:      true,
	Validators:    []tfsdk.AttributeValidator{validators.NewValidator(mtypes.LocationType)},
	PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
}

var ResourceStatusSchema = tfsdk.Attribute{
	Description:   "Statuses of underlying created resources",
	Type:          types.MapType{ElemType: types.StringType},
	Computed:      true,
	PlanModifiers: []tfsdk.AttributePlanModifier{validators.ResourceStatusModifier{}},
}

var AwsSchema = tfsdk.Attribute{
	Type:     types.MapType{},
	Computed: true,
}

var AzureSchema = tfsdk.Attribute{
	Type:     types.MapType{},
	Computed: true,
}

func RequiresReplaceIfCloudEq(replaceIfCloud ...string) tfsdk.AttributePlanModifier {
	return validators.RequiresReplaceIf(func(ctx context.Context, state, config attr.Value, plan tfsdk.Plan) (bool, diag.Diagnostics) {
		return requiresReplaceIfCloudEq(plan, replaceIfCloud...)
	}, fmt.Sprintf("Resource is replaced if cloud is %s", replaceIfCloud), fmt.Sprintf("Resource is replaced if cloud is %s", replaceIfCloud))
}

func requiresReplaceIfCloudEq(plan tfsdk.Plan, replaceIfCloud ...string) (bool, diag.Diagnostics) {
	cloud := ""
	diags := plan.GetAttribute(context.Background(), tftypes.NewAttributePath().WithAttributeName("cloud"), &cloud)
	if diags.HasError() {
		return false, diags
	}
	return slices.Contains(replaceIfCloud, cloud), nil
}
