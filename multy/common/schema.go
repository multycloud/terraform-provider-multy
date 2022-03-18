package common

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var RgVarsSchema = &schema.Schema{
	Type:     schema.TypeMap,
	Optional: true,
	Elem: &schema.Schema{
		Type: schema.TypeString,
	},
}

var CloudsSchema = &schema.Schema{
	Type:     schema.TypeList,
	Optional: true,
	Elem: &schema.Schema{
		Type:         schema.TypeString,
		ValidateFunc: validation.StringInSlice(GetCloudNames(), true),
	},
}

var LocationSchema = &schema.Schema{
	Type:         schema.TypeString,
	Optional:     true,
	ValidateFunc: validation.StringInSlice(GetLocationNames(), true),
}
