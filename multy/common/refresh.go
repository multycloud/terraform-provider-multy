package common

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	mproto "github.com/multycloud/multy/api/proto"
	"github.com/multycloud/multy/api/proto/commonpb"
	"golang.org/x/sync/errgroup"
	"sync"
)

type RefreshCache struct {
	cache sync.Map
}

type refreshResult struct {
	sync.Mutex
	done   bool
	result error
}

func (r *RefreshCache) Refresh(ctx context.Context, apiKey string, provider *ProviderConfig) error {
	var wg errgroup.Group
	if provider.Aws != nil {
		wg.Go(func() error {
			return r.refresh(ctx, apiKey, provider, commonpb.CloudProvider_AWS)
		})
	}
	if provider.Azure != nil {
		wg.Go(func() error {
			return r.refresh(ctx, apiKey, provider, commonpb.CloudProvider_AZURE)
		})
	}
	if provider.Gcp != nil {
		wg.Go(func() error {
			return r.refresh(ctx, apiKey, provider, commonpb.CloudProvider_GCP)
		})
	}

	err := wg.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (r *RefreshCache) refresh(ctx context.Context, apiKey string, provider *ProviderConfig, cloud commonpb.CloudProvider) error {
	value, _ := r.cache.LoadOrStore(fmt.Sprintf("%s/%s", apiKey, cloud.String()), &refreshResult{})
	result := value.(*refreshResult)
	result.Lock()
	defer result.Unlock()
	if !result.done {
		tflog.Info(ctx, fmt.Sprintf("refreshing state for %s/%s", apiKey, cloud.String()))
		_, err := provider.Client.RefreshState(ctx, &mproto.RefreshStateRequest{Cloud: cloud})
		result.done = true
		result.result = err
	} else {
		tflog.Info(ctx, fmt.Sprintf("skipping refreshing state for %s/%s", apiKey, cloud.String()))
	}

	return result.result
}
