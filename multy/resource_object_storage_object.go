package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceObjectStorageObjectType struct{}

var objectStorageObjectAwsOutputs = map[string]attr.Type{
	"s3_bucket_object_id": types.StringType,
}

var objectStorageObjectAzureOutputs = map[string]attr.Type{
	"storage_blob_id": types.StringType,
}

var objectStorageObjectGcpOutputs = map[string]attr.Type{
	"storage_bucket_object_id":      types.StringType,
	"storage_object_access_control": types.StringType,
}

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
				Type:          types.StringType,
				Description:   "Name of object storage object",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"object_storage_id": {
				Type:          types.StringType,
				Description:   "Id of object storage",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"content_base64": {
				Type:        types.StringType,
				Description: "Content of the object",
				Required:    true,
				Validators:  []tfsdk.AttributeValidator{mtypes.NonEmptyStringValidator},
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
			"aws": {
				Description: "AWS-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: objectStorageObjectAwsOutputs},
				Computed:    true,
			},
			"azure": {
				Description: "Azure-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: objectStorageObjectAzureOutputs},
				Computed:    true,
			},
			"gcp": {
				Description: "GCP-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: objectStorageObjectGcpOutputs},
				Computed:    true,
			},
			"resource_status": common.ResourceStatusSchema,
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
	Name            types.String                                         `tfsdk:"name"`
	Acl             mtypes.EnumValue[resourcespb.ObjectStorageObjectAcl] `tfsdk:"acl"`
	ObjectStorageId types.String                                         `tfsdk:"object_storage_id"`
	ContentBase64   types.String                                         `tfsdk:"content_base64"`
	ContentType     types.String                                         `tfsdk:"content_type"`
	Url             types.String                                         `tfsdk:"url"`
	AwsOutputs      types.Object                                         `tfsdk:"aws"`
	AzureOutputs    types.Object                                         `tfsdk:"azure"`
	GcpOutputs      types.Object                                         `tfsdk:"gcp"`
	ResourceStatus  types.Map                                            `tfsdk:"resource_status"`
}

func convertToObjectStorageObject(res *resourcespb.ObjectStorageObjectResource) ObjectStorageObject {
	return ObjectStorageObject{
		Id:              types.String{Value: res.CommonParameters.ResourceId},
		Name:            types.String{Value: res.Name},
		Acl:             mtypes.ObjectAclType.NewVal(res.Acl),
		ObjectStorageId: types.String{Value: res.ObjectStorageId},
		ContentBase64:   common.DefaultToNull[types.String](res.ContentBase64),
		ContentType:     common.DefaultToNull[types.String](res.ContentType),
		Url:             types.String{Value: res.Url},
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"s3_bucket_object_id": common.DefaultToNull[types.String](res.GetAwsOutputs().GetS3BucketObjectId()),
			},
			AttrTypes: objectStorageObjectAwsOutputs,
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"storage_blob_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetStorageBlobId()),
			},
			AttrTypes: objectStorageObjectAzureOutputs,
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"storage_bucket_object_id":      common.DefaultToNull[types.String](res.GetGcpOutputs().GetStorageBucketObjectId()),
				"storage_object_access_control": common.DefaultToNull[types.String](res.GetGcpOutputs().GetStorageObjectAccessControl()),
			},
			AttrTypes: objectStorageObjectGcpOutputs,
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromObjectStorageObject(plan ObjectStorageObject) *resourcespb.ObjectStorageObjectArgs {
	return &resourcespb.ObjectStorageObjectArgs{
		Name:            plan.Name.Value,
		Acl:             plan.Acl.Value,
		ObjectStorageId: plan.ObjectStorageId.Value,
		ContentBase64:   plan.ContentBase64.Value,
		ContentType:     common.NullToDefault[string](plan.ContentType),
	}
}
