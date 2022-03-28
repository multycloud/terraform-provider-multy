package common

import (
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mproto "github.com/multycloud/multy/api/proto"
	common_proto "github.com/multycloud/multy/api/proto/common"
	"github.com/multycloud/multy/api/proto/creds"
	"google.golang.org/grpc/metadata"
)

type ProviderConfig struct {
	Client mproto.MultyResourceServiceClient
	ApiKey string
	Aws    *AwsConfig
	Azure  *AzureConfig
}

func (c *ProviderConfig) AddHeaders(ctx context.Context) (context.Context, error) {
	cloudCreds := &creds.CloudCredentials{
		AwsCreds: &creds.AwsCredentials{
			AccessKey: c.Aws.AccessKeyId,
			SecretKey: c.Aws.AccessKeySecret,
		},
		AzureCreds: &creds.AzureCredentials{
			SubscriptionId: c.Azure.SubscriptionId,
			TenantId:       c.Azure.TenantId,
			ClientId:       c.Azure.ClientId,
			ClientSecret:   c.Azure.ClientSecret,
		},
	}

	b, err := proto.Marshal(cloudCreds)
	if err != nil {
		return nil, err
	}
	fmt.Println(string(b))
	ctx = metadata.AppendToOutgoingContext(ctx, "cloud-creds-bin", string(b))
	// TODO: retrieve user id from api key
	return metadata.AppendToOutgoingContext(ctx, "user_id", c.ApiKey), nil
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

type AwsConfig struct {
	AccessKeyId     string
	AccessKeySecret string
}

type AzureConfig struct {
	SubscriptionId string
	ClientId       string
	ClientSecret   string
	TenantId       string
}
