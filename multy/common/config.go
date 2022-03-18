package common

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/multycloud/multy/api/proto"
	common_proto "github.com/multycloud/multy/api/proto/common"
	"google.golang.org/grpc/metadata"
)

type ProviderConfig struct {
	Client   proto.MultyResourceServiceClient
	ApiKey   string
	Clouds   []common_proto.CloudProvider
	Location common_proto.Location
}

func (c *ProviderConfig) AddHeaders(ctx context.Context) context.Context {
	// TODO: retrieve user id from api key
	return metadata.AppendToOutgoingContext(ctx, "user_id", c.ApiKey)
}

func (c *ProviderConfig) GetLocation(d *schema.ResourceData) common_proto.Location {
	if loc, check := d.GetOk("location"); check {
		return StringToLocation(loc.(string))
	}
	return c.Location
}

func (c *ProviderConfig) GetClouds(d *schema.ResourceData) []common_proto.CloudProvider {
	if clouds, check := d.GetOk("clouds"); check && len(clouds.([]interface{})) != 0 {
		return ListToCloudList(InterfaceToStringMap(clouds.([]interface{})))
	}
	return c.Clouds
}
