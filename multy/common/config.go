package common

import (
	"context"
	mproto "github.com/multycloud/multy/api/proto"
	"github.com/multycloud/multy/api/proto/credspb"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type ProviderConfig struct {
	Client mproto.MultyResourceServiceClient
	ApiKey string
	Aws    *AwsConfig
	Azure  *AzureConfig
	Gcp    *GcpConfig
}

func (c *ProviderConfig) AddHeaders(ctx context.Context) (context.Context, error) {
	var cloudCreds credspb.CloudCredentials
	if c.Aws != nil {
		cloudCreds.AwsCreds = &credspb.AwsCredentials{
			AccessKey: c.Aws.AccessKeyId,
			SecretKey: c.Aws.AccessKeySecret,

			SessionToken: c.Aws.SessionToken,
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

	if c.Gcp != nil {
		cloudCreds.GcpCreds = &credspb.GCPCredentials{
			Credentials: c.Gcp.Credentials,
			Project:     c.Gcp.Project,
		}
	}

	b, err := proto.Marshal(&cloudCreds)
	if err != nil {
		return nil, err
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "cloud-creds-bin", string(b))
	// TODO: retrieve user id from api key
	return metadata.AppendToOutgoingContext(ctx, "api_key", c.ApiKey), nil
}

//
//func (c *ProviderConfig) GetClouds(d *schema.ResourceData) common_proto.CloudProvider {
//	if clouds, check := d.GetOk("clouds"); check && len(clouds.([]interface{})) != 0 {
//		return ListToCloudList(InterfaceToStringMap(clouds.([]interface{})))
//	}
//	return c.Clouds
//}

type AwsConfig struct {
	AccessKeyId     string
	AccessKeySecret string

	SessionToken string
}

type AzureConfig struct {
	SubscriptionId string
	ClientId       string
	ClientSecret   string
	TenantId       string
}

type GcpConfig struct {
	Project     string
	Credentials string
}
