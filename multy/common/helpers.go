package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

func MapTypeToGoMap(t types.Map) map[string]string {
	if t.Unknown || t.Null {
		return nil
	}
	res := map[string]string{}
	for k, elem := range t.Elems {
		res[k] = elem.(types.String).Value
	}

	return res
}

func GoMapToMapType(t map[string]string) types.Map {
	if t == nil || len(t) == 0 {
		return types.Map{
			Unknown:  false,
			Null:     true,
			Elems:    nil,
			ElemType: types.StringType,
		}
	}

	elems := map[string]attr.Value{}

	for k, v := range t {
		elems[k] = types.String{Value: v}
	}

	return types.Map{
		Unknown:  false,
		Null:     false,
		Elems:    elems,
		ElemType: types.StringType,
	}
}
