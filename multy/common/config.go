package common

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/multycloud/multy/api/proto"
	common_proto "github.com/multycloud/multy/api/proto/common"
	"google.golang.org/grpc/metadata"
)

type ProviderConfig struct {
	Client proto.MultyResourceServiceClient
	ApiKey string
}

func (c *ProviderConfig) AddHeaders(ctx context.Context) context.Context {
	// TODO: retrieve user id from api key
	return metadata.AppendToOutgoingContext(ctx, "user_id", c.ApiKey)
}

//
//func (c *ProviderConfig) GetClouds(d *schema.ResourceData) common_proto.CloudProvider {
//	if clouds, check := d.GetOk("clouds"); check && len(clouds.([]interface{})) != 0 {
//		return ListToCloudList(InterfaceToStringMap(clouds.([]interface{})))
//	}
//	return c.Clouds
//}

func (c *ProviderConfig) GetOperatingSystem(d *schema.ResourceData) common_proto.OperatingSystem_Enum {
	if loc, check := d.GetOk("operating_system"); check {
		return StringToVmOperatingSystem(loc.(string))
	}
	return common_proto.OperatingSystem_UNKNOWN_OS
}

func (c *ProviderConfig) GetVmSize(d *schema.ResourceData) common_proto.VmSize_Enum {
	if loc, check := d.GetOk("size"); check {
		return StringToVmSize(loc.(string))
	}
	return common_proto.VmSize_UNKNOWN_VM_SIZE
}
