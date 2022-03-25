package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	common_proto "github.com/multycloud/multy/api/proto/common"
	"github.com/multycloud/multy/api/proto/errors"
	"github.com/multycloud/multy/api/proto/resources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

func GetEnumNames(vals map[string]int32) []string {
	var keys []string
	for l := range vals {
		keys = append(keys, strings.ToLower(l))
	}
	return keys
}

func GetLocationNames() []string {
	return GetEnumNames(common_proto.Location_value)
}

func GetCloudNames() []string {
	return GetEnumNames(common_proto.CloudProvider_value)
}

func GetVmOperatingSystem() []string {
	return GetEnumNames(common_proto.OperatingSystem_Enum_value)
}

func GetVmSize() []string {
	return GetEnumNames(common_proto.VmSize_Enum_value)
}

func GetRouteDestinations() []string {
	return GetEnumNames(resources.RouteDestination_value)
}

func StringToLocation(loc string) common_proto.Location {
	return common_proto.Location(common_proto.Location_value[strings.ToUpper(loc)])
}

func StringToVmOperatingSystem(os string) common_proto.OperatingSystem_Enum {
	return common_proto.OperatingSystem_Enum(common_proto.OperatingSystem_Enum_value[strings.ToUpper(os)])
}

func StringToVmSize(os string) common_proto.VmSize_Enum {
	return common_proto.VmSize_Enum(common_proto.OperatingSystem_Enum_value[strings.ToUpper(os)])
}

func StringToCloud(cloud string) common_proto.CloudProvider {
	return common_proto.CloudProvider(common_proto.CloudProvider_value[strings.ToUpper(cloud)])
}

func StringToRouteDestination(route string) resources.RouteDestination {
	return resources.RouteDestination(resources.RouteDestination_value[strings.ToUpper(route)])
}

func StringToRuleDirection(dir string) resources.Direction {
	if strings.EqualFold(dir, "both") {
		dir = "both_directions"
	}
	return resources.Direction(resources.Direction_value[strings.ToUpper(dir)])
}

func RuleDirectionToString(dir resources.Direction) string {
	if strings.EqualFold(dir.String(), "BOTH_DIRECTIONS") {
		return "both"
	}
	return strings.ToLower(dir.String())
}

func ListToCloudList(clouds []string) []common_proto.CloudProvider {
	var cloudList []common_proto.CloudProvider
	for _, c := range clouds {
		cloudList = append(cloudList, common_proto.CloudProvider(common_proto.CloudProvider_value[strings.ToUpper(c)]))
	}
	return cloudList
}

func ParseGrpcErrors(err error) string {
	s, ok := status.FromError(err)
	if !ok {
		return err.Error()
	}

	str := ""

	// user screwed up
	if s.Code() == codes.InvalidArgument {
		str += s.Message()
		for _, detail := range s.Details() {
			if e, ok := detail.(*errors.ResourceValidationError); ok {
				str += "\n" + e.ErrorMessage
			}
		}
	} else if s.Code() == codes.Internal {
		str += "something went wrong: " + s.Message()
	} else {
		str += s.String()
	}
	return str

}

func DefaultToNull[OutT attr.Value](t any) OutT {
	var s attr.Value
	switch t.(type) {
	case string:
		s = types.String{
			Null:  t.(string) == "",
			Value: t.(string),
		}
	case int, int64, int32:
		s = types.Int64{
			Null:  t.(int64) == 0,
			Value: t.(int64),
		}
	}
	return s.(OutT)
}

type protoEnum interface {
	String() string
	Number() protoreflect.EnumNumber
}

func DefaultEnumToNull(p protoEnum) types.String {
	return types.String{
		Null:  p.Number() == 0,
		Value: strings.ToLower(p.String()),
	}
}

func DefaultSliceToNull[T attr.Value](t []T) []T {
	if len(t) == 0 {
		return nil
	}
	return t
}
