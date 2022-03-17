package common

import (
	"context"
	"github.com/multycloud/multy/api/proto"
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
