package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/commonpb"
)

func GetResourceStatus(params *commonpb.ResourceStatus) types.Map {
	if params == nil || len(params.GetStatuses()) == 0 {
		return types.MapNull(types.StringType)
	}

	elems := map[string]attr.Value{}
	for k, v := range params.GetStatuses() {
		elems[k] = types.StringValue(v.String())
	}

	value, _ := types.MapValue(types.StringType, elems)

	return value
}
