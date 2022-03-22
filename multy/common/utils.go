package common

import (
	common_proto "github.com/multycloud/multy/api/proto/common"
	"strings"
)

func GetLocationNames() []string {
	var keys []string
	for l, _ := range common_proto.Location_value {
		keys = append(keys, strings.ToLower(l))
	}
	return keys
}

func GetCloudNames() []string {
	var keys []string
	for l, _ := range common_proto.CloudProvider_value {
		keys = append(keys, strings.ToLower(l))
	}
	return keys
}

func GetVmOperatingSystem() []string {
	var keys []string
	for l, _ := range common_proto.OperatingSystem_Enum_value {
		keys = append(keys, strings.ToLower(l))
	}
	return keys
}

func GetVmSize() []string {
	var keys []string
	for l, _ := range common_proto.VmSize_Enum_value {
		keys = append(keys, strings.ToLower(l))
	}
	return keys
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

func ListToCloudList(clouds []string) []common_proto.CloudProvider {
	var cloudList []common_proto.CloudProvider
	for _, c := range clouds {
		cloudList = append(cloudList, common_proto.CloudProvider(common_proto.CloudProvider_value[strings.ToUpper(c)]))
	}
	return cloudList
}
