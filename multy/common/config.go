package common

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	mproto "github.com/multycloud/multy/api/proto"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/credspb"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type ProviderConfig struct {
	Client mproto.MultyResourceServiceClient
	ApiKey string
	Aws    *AwsConfig
	Azure  *AzureConfig
}

func (c *ProviderConfig) AddHeaders(ctx context.Context) (context.Context, error) {
	var cloudCreds credspb.CloudCredentials
	if c.Aws != nil {
		cloudCreds.AwsCreds = &credspb.AwsCredentials{
			AccessKey: c.Aws.AccessKeyId,
			SecretKey: c.Aws.AccessKeySecret,
		}
	}
	if c.Azure != nil {
		cloudCreds.AzureCreds = &credspb.AzureCredentials{
			SubscriptionId: c.Azure.SubscriptionId,
			TenantId:       c.Azure.TenantId,
			ClientId:       c.Azure.ClientId,
			ClientSecret:   c.Azure.ClientSecret,
		}
	}

	b, err := proto.Marshal(&cloudCreds)
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

func (c *ProviderConfig) GetOperatingSystem(d *schema.ResourceData) commonpb.OperatingSystem_Enum {
	if loc, check := d.GetOk("operating_system"); check {
		return StringToVmOperatingSystem(loc.(string))
	}
	return commonpb.OperatingSystem_UNKNOWN_OS
}

func (c *ProviderConfig) GetVmSize(d *schema.ResourceData) commonpb.VmSize_Enum {
	if loc, check := d.GetOk("size"); check {
		return StringToVmSize(loc.(string))
	}
	return commonpb.VmSize_UNKNOWN_VM_SIZE
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
