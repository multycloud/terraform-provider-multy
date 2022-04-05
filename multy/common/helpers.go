package common

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func StringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func InterfaceToStringMap(t []interface{}) []string {
	s := make([]string, len(t))
	for i, v := range t {
		s[i] = fmt.Sprint(v)
	}
	return s
}

func StringSliceToTypesString(t []types.String) []string {
	s := make([]string, len(t))
	for i, v := range t {
		s[i] = fmt.Sprint(v)
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
