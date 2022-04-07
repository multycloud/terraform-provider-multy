package common

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func StringSliceToTypesString(t []types.String) []string {
	s := make([]string, len(t))
	for i, v := range t {
		if v.Null {
			panic("unexpected null value")
		}
		s[i] = v.Value
	}
	return s
}

func TypesStringToStringSlice(t []string) []types.String {
	s := make([]types.String, len(t))
	for i, v := range t {
		s[i] = types.String{Value: v}
	}
	return s
}
