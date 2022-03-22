package common

import (
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"terraform-provider-multy/multy/validators"
)

var RgVarsSchema = &schema.Schema{
	Type:     schema.TypeMap,
	Optional: true,
	Elem: &schema.Schema{
		Type: schema.TypeString,
	},
}

var CloudsSchema = tfsdk.Attribute{
	Type:       types.StringType,
	Required:   true,
	Validators: []tfsdk.AttributeValidator{validators.StringInSliceValidator{Enum: GetCloudNames()}},
}

var LocationSchema = tfsdk.Attribute{
	Type:       types.StringType,
	Optional:   true,
	Validators: []tfsdk.AttributeValidator{validators.StringInSliceValidator{Enum: GetLocationNames()}},
}

var AwsSchema = tfsdk.Attribute{
	Type:     types.MapType{},
	Computed: true,
}

var AzureSchema = tfsdk.Attribute{
	Type:     types.MapType{},
	Computed: true,
}
