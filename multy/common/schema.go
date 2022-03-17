package common

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		Type: schema.TypeString,
	},
	//ValidateFunc: validation.StringInSlice(Clouds, true),
}
