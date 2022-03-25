package common

import (
	common_proto "github.com/multycloud/multy/api/proto/common"
	"github.com/multycloud/multy/api/proto/errors"
	"github.com/multycloud/multy/api/proto/resources"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func StringToRuleDirection(dir string) resources.Direction {
	if strings.EqualFold(dir, "both") {
		dir = "both_directions"
	}
	return resources.Direction(resources.Direction_value[strings.ToUpper(dir)])
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
