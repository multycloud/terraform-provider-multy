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

func GetVmOperatingSystem() []string {
	return GetEnumNames(commonpb.OperatingSystem_Enum_value)
}

func GetVmSize() []string {
	return GetEnumNames(commonpb.VmSize_Enum_value)
}

func GetRouteDestinations() []string {
	return GetEnumNames(resourcespb.RouteDestination_value)
}

func StringToVmOperatingSystem(os string) commonpb.OperatingSystem_Enum {
	return StringToEnum[commonpb.OperatingSystem_Enum](commonpb.OperatingSystem_Enum_value, os)
}

func StringToVmSize(os string) commonpb.VmSize_Enum {
	return StringToEnum[commonpb.VmSize_Enum](commonpb.VmSize_Enum_value, os)
}

func StringToEnum[T ~int32](values map[string]int32, value string) T {
	if value == "" {
		return T(0)
	}
	return T(values[strings.ToUpper(value)])
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

func NullToDefault[OutT any, T attr.Value](t T) OutT {
	returnVal := new(OutT)
	value, err := t.ToTerraformValue(nil)
	if err != nil {
		panic(err)
	}
	if value.IsNull() {
		return *returnVal
	}

	err = value.As(returnVal)
	if err != nil {
		panic(err)
	}

	return *returnVal
}

type protoEnum interface {
	String() string
	Number() protoreflect.EnumNumber
}

func DefaultEnumToNullString(p protoEnum) types.String {
	return types.String{
		Null:  p.Number() == 0,
		Value: strings.ToLower(p.String()),
	}
}

func EnumToString(p protoEnum) types.String {
	return types.String{
		Value: strings.ToLower(p.String()),
	}
}

func DefaultSliceToNull[T attr.Value](t []T) []T {
	if len(t) == 0 {
		return nil
	}
	return t
}
