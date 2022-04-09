package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
)

type ResourceObjectStorageType struct{}

func (r ResourceObjectStorageType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Object Storage resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of Virtual Network",
				Required:    true,
			},
			"versioning": {
				Type:        types.BoolType,
				Description: "If true, versioning will be enabled to `object_storage_object`",
				Optional:    true,
				Computed:    true,
			},
			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,
		},
	}, nil
}

func (r ResourceObjectStorageType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[ObjectStorage]{
		p:          *(p.(*Provider)),
		createFunc: createObjectStorage,
		updateFunc: updateObjectStorage,
		readFunc:   readObjectStorage,
		deleteFunc: deleteObjectStorage,
	}, nil
}

func createObjectStorage(ctx context.Context, p Provider, plan ObjectStorage) (ObjectStorage, error) {
	vn, err := p.Client.Client.CreateObjectStorage(ctx, &resourcespb.CreateObjectStorageRequest{
		Resource: convertFromObjectStorage(plan),
	})
	if err != nil {
		return ObjectStorage{}, err
	}
	return convertToObjectStorage(vn), nil
}

func updateObjectStorage(ctx context.Context, p Provider, plan ObjectStorage) (ObjectStorage, error) {
	vn, err := p.Client.Client.UpdateObjectStorage(ctx, &resourcespb.UpdateObjectStorageRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromObjectStorage(plan),
	})
	if err != nil {
		return ObjectStorage{}, err
	}
	return convertToObjectStorage(vn), nil
}

func readObjectStorage(ctx context.Context, p Provider, state ObjectStorage) (ObjectStorage, error) {
	vn, err := p.Client.Client.ReadObjectStorage(ctx, &resourcespb.ReadObjectStorageRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return ObjectStorage{}, err
	}
	return convertToObjectStorage(vn), nil
}

func deleteObjectStorage(ctx context.Context, p Provider, state ObjectStorage) error {
	_, err := p.Client.Client.DeleteObjectStorage(ctx, &resourcespb.DeleteObjectStorageRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type ObjectStorage struct {
	Id         types.String                             `tfsdk:"id"`
	Name       types.String                             `tfsdk:"name"`
	Versioning types.Bool                               `tfsdk:"versioning"`
	Cloud      mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location   mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
}

func convertToObjectStorage(res *resourcespb.ObjectStorageResource) ObjectStorage {
	return ObjectStorage{
		Id:         types.String{Value: res.CommonParameters.ResourceId},
		Name:       types.String{Value: res.Name},
		Versioning: types.Bool{Value: res.Versioning},
		Cloud:      mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:   mtypes.LocationType.NewVal(res.CommonParameters.Location),
	}
}

func convertFromObjectStorage(plan ObjectStorage) *resourcespb.ObjectStorageArgs {
	return &resourcespb.ObjectStorageArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			Location:      plan.Location.Value,
			CloudProvider: plan.Cloud.Value,
		},
		Name:       plan.Name.Value,
		Versioning: plan.Versioning.Value,
	}
}
