package common

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/errorspb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
)

const DebugMode = true

func GetEnumNames(vals map[string]int32) []string {
	var keys []string
	for l := range vals {
		keys = append(keys, strings.ToLower(l))
	}
	return keys
}

func GetLocationNames() []string {
	return GetEnumNames(commonpb.Location_value)
}

func GetCloudNames() []string {
	return GetEnumNames(commonpb.CloudProvider_value)
}

func GetVmOperatingSystem() []string {
	return GetEnumNames(commonpb.OperatingSystem_Enum_value)
}

func GetVmSize() []string {
	return GetEnumNames(commonpb.VmSize_Enum_value)
}

func GetRouteDestinations() []string {
	return GetEnumNames(resourcespb.RouteDestination_value)
}

func StringToLocation(loc string) commonpb.Location {
	return commonpb.Location(commonpb.Location_value[strings.ToUpper(loc)])
}

func StringToVmOperatingSystem(os string) commonpb.OperatingSystem_Enum {
	return commonpb.OperatingSystem_Enum(commonpb.OperatingSystem_Enum_value[strings.ToUpper(os)])
}

func StringToVmSize(os string) commonpb.VmSize_Enum {
	return commonpb.VmSize_Enum(commonpb.VmSize_Enum_value[strings.ToUpper(os)])
}

func StringToCloud(cloud string) commonpb.CloudProvider {
	return commonpb.CloudProvider(commonpb.CloudProvider_value[strings.ToUpper(cloud)])
}

func StringToRouteDestination(route string) resourcespb.RouteDestination {
	return resourcespb.RouteDestination(resourcespb.RouteDestination_value[strings.ToUpper(route)])
}

func StringToRuleDirection(dir string) resourcespb.Direction {
	if strings.EqualFold(dir, "both") {
		dir = "both_directions"
	}
	return resourcespb.Direction(resourcespb.Direction_value[strings.ToUpper(dir)])
}

func RuleDirectionToString(dir resourcespb.Direction) string {
	if strings.EqualFold(dir.String(), "BOTH_DIRECTIONS") {
		return "both"
	}
	return strings.ToLower(dir.String())
}

func ListToCloudList(clouds []string) []commonpb.CloudProvider {
	var cloudList []commonpb.CloudProvider
	for _, c := range clouds {
		cloudList = append(cloudList, commonpb.CloudProvider(commonpb.CloudProvider_value[strings.ToUpper(c)]))
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
			if e, ok := detail.(*errorspb.ResourceValidationError); ok {
				str += "\n" + e.ErrorMessage
			}
		}
	} else if s.Code() == codes.Internal {
		str += "something went wrong: " + s.Message()
		if DebugMode {
			for _, detail := range s.Details() {
				if e, ok := detail.(*errorspb.InternalErrorDetails); ok {
					str += "\n" + e.ErrorMessage
				}
			}
		}
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
