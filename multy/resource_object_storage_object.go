package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceObjectStorageObjectType struct{}

func (r ResourceObjectStorageObjectType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Object Storage Object resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of object storage object",
				Required:    true,
			},
			"object_storage_id": {
				Type:          types.StringType,
				Description:   "Id of object storage",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"content": {
				Type:        types.StringType,
				Description: "Content of the object",
				Required:    true,
			},
			"content_type": {
				Type:        types.StringType,
				Description: "Standard MIME type describing the format of the object data",
				Optional:    true,
				Validators:  []tfsdk.AttributeValidator{mtypes.NonEmptyStringValidator},
			},
			"acl": {
				Type:        mtypes.ObjectAclType,
				Description: "Access control for the given object. Can be public_read or private. Defaults to private.",
				Optional:    true,
				Computed:    true,
				Validators:  []tfsdk.AttributeValidator{validators.NewValidator(mtypes.ObjectAclType)},
			},
			// outputs
			"url": {
				Type:        types.StringType,
				Description: "URL of object",
				Computed:    true,
			},
		},
	}, nil
}

func (r ResourceObjectStorageObjectType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[ObjectStorageObject]{
		p:          *(p.(*Provider)),
		createFunc: createObjectStorageObject,
		updateFunc: updateObjectStorageObject,
		readFunc:   readObjectStorageObject,
		deleteFunc: deleteObjectStorageObject,
	}, nil
}

func createObjectStorageObject(ctx context.Context, p Provider, plan ObjectStorageObject) (ObjectStorageObject, error) {
	vn, err := p.Client.Client.CreateObjectStorageObject(ctx, &resourcespb.CreateObjectStorageObjectRequest{
		Resource: convertFromObjectStorageObject(plan),
	})
	if err != nil {
		return ObjectStorageObject{}, err
	}
	return convertToObjectStorageObject(vn), nil
}

func updateObjectStorageObject(ctx context.Context, p Provider, plan ObjectStorageObject) (ObjectStorageObject, error) {
	vn, err := p.Client.Client.UpdateObjectStorageObject(ctx, &resourcespb.UpdateObjectStorageObjectRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromObjectStorageObject(plan),
	})
	if err != nil {
		return ObjectStorageObject{}, err
	}
	return convertToObjectStorageObject(vn), nil
}

func readObjectStorageObject(ctx context.Context, p Provider, state ObjectStorageObject) (ObjectStorageObject, error) {
	vn, err := p.Client.Client.ReadObjectStorageObject(ctx, &resourcespb.ReadObjectStorageObjectRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return ObjectStorageObject{}, err
	}
	return convertToObjectStorageObject(vn), nil
}

func deleteObjectStorageObject(ctx context.Context, p Provider, state ObjectStorageObject) error {
	_, err := p.Client.Client.DeleteObjectStorageObject(ctx, &resourcespb.DeleteObjectStorageObjectRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type ObjectStorageObject struct {
	Id              types.String                                         `tfsdk:"id"`
	Name            string                                               `tfsdk:"name"`
	Acl             mtypes.EnumValue[resourcespb.ObjectStorageObjectAcl] `tfsdk:"acl"`
	ObjectStorageId string                                               `tfsdk:"object_storage_id"`
	Content         string                                               `tfsdk:"content"`
	ContentType     types.String                                         `tfsdk:"content_type"`
	Url             types.String                                         `tfsdk:"url"`
}

func convertToObjectStorageObject(res *resourcespb.ObjectStorageObjectResource) ObjectStorageObject {
	return ObjectStorageObject{
		Id:              types.String{Value: res.CommonParameters.ResourceId},
		Name:            res.Name,
		Acl:             mtypes.ObjectAclType.NewVal(res.Acl),
		ObjectStorageId: res.ObjectStorageId,
		Content:         res.Content,
		ContentType:     common.DefaultToNull[types.String](res.ContentType),
		Url:             types.String{Value: res.Url},
	}
}

func convertFromObjectStorageObject(plan ObjectStorageObject) *resourcespb.ObjectStorageObjectArgs {
	return &resourcespb.ObjectStorageObjectArgs{
		Name:            plan.Name,
		Acl:             plan.Acl.Value,
		ObjectStorageId: plan.ObjectStorageId,
		Content:         plan.Content,
		ContentType:     common.NullToDefault[string](plan.ContentType),
	}
}
