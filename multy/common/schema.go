package common

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	Description:   fmt.Sprintf("Location to deploy resource into. Accepted values are %s", StringSliceToDocsMarkdown(mtypes.LocationType.GetAllValues())),
	Required:      true,
	Validators:    []tfsdk.AttributeValidator{validators.NewValidator(mtypes.LocationType)},
	PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
}

var AwsSchema = tfsdk.Attribute{
	Type:     types.MapType{},
	Computed: true,
}

var AzureSchema = tfsdk.Attribute{
	Type:     types.MapType{},
	Computed: true,
}
