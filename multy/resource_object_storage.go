package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
)

type ResourceObjectStorageType struct{}

var objectStorageAwsOutputs = map[string]attr.Type{
	"s3_bucket_arn": types.StringType,
}

var objectStorageAzureOutputs = map[string]attr.Type{
	"storage_account_id":           types.StringType,
	"public_storage_container_id":  types.StringType,
	"private_storage_container_id": types.StringType,
}

var objectStorageGcpOutputs = map[string]attr.Type{
	"storage_bucket_id": types.StringType,
}

var objectStorageSchema = tfsdk.Schema{
	MarkdownDescription: "Provides Multy Object Storage resource",
	Attributes: map[string]tfsdk.Attribute{
		"id": {
			Type:          types.StringType,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		},
		"resource_group_id": {
			Type:          types.StringType,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		},
		"name": {
			Type:          types.StringType,
			Description:   "Name of Object Storage",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"versioning": {
			Type:        types.BoolType,
			Description: "If true, versioning will be enabled to `object_storage_object`",
			Optional:    true,
			Computed:    true,
		},
		"cloud":    common.CloudsSchema,
		"location": common.LocationSchema,
		"gcp_overrides": {
			Description: "GCP-specific attributes that will be set if this resource is deployed in GCP",
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"project": {
					Type:          types.StringType,
					Description:   fmt.Sprintf("The project to use for this resource."),
					Optional:      true,
					Computed:      true,
					PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("gcp"), resource.UseStateForUnknown()},
					Validators:    []tfsdk.AttributeValidator{mtypes.NonEmptyStringValidator},
				},
			}),
			Optional: true,
			Computed: true,
		},
		"aws": {
			Description: "AWS-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: objectStorageAwsOutputs},
			Computed:    true,
		},
		"azure": {
			Description: "Azure-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: objectStorageAzureOutputs},
			Computed:    true,
		},
		"gcp": {
			Description: "GCP-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: objectStorageGcpOutputs},
			Computed:    true,
		},
		"resource_status": common.ResourceStatusSchema,
	},
}

func (r ResourceObjectStorageType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[ObjectStorage]{
		p:          *(p.(*Provider)),
		createFunc: createObjectStorage,
		updateFunc: updateObjectStorage,
		readFunc:   readObjectStorage,
		deleteFunc: deleteObjectStorage,
		name:       "multy_object_storage",
		schema:     objectStorageSchema,
	}
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
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromObjectStorage(plan),
	})
	if err != nil {
		return ObjectStorage{}, err
	}
	return convertToObjectStorage(vn), nil
}

func readObjectStorage(ctx context.Context, p Provider, state ObjectStorage) (ObjectStorage, error) {
	vn, err := p.Client.Client.ReadObjectStorage(ctx, &resourcespb.ReadObjectStorageRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return ObjectStorage{}, err
	}
	return convertToObjectStorage(vn), nil
}

func deleteObjectStorage(ctx context.Context, p Provider, state ObjectStorage) error {
	_, err := p.Client.Client.DeleteObjectStorage(ctx, &resourcespb.DeleteObjectStorageRequest{
		ResourceId: state.Id.ValueString(),
	})
	return err
}

type ObjectStorage struct {
	Id              types.String                             `tfsdk:"id"`
	Name            types.String                             `tfsdk:"name"`
	Versioning      types.Bool                               `tfsdk:"versioning"`
	Cloud           mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location        mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
	ResourceGroupId types.String                             `tfsdk:"resource_group_id"`

	GcpOverridesObject types.Object `tfsdk:"gcp_overrides"`
	AwsOutputs         types.Object `tfsdk:"aws"`
	AzureOutputs       types.Object `tfsdk:"azure"`
	GcpOutputs         types.Object `tfsdk:"gcp"`
	ResourceStatus     types.Map    `tfsdk:"resource_status"`
}

func convertToObjectStorage(res *resourcespb.ObjectStorageResource) ObjectStorage {
	return ObjectStorage{
		Id:                 types.StringValue(res.CommonParameters.ResourceId),
		Name:               types.StringValue(res.Name),
		Versioning:         types.BoolValue(res.Versioning),
		Cloud:              mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:           mtypes.LocationType.NewVal(res.CommonParameters.Location),
		ResourceGroupId:    types.StringValue(res.CommonParameters.ResourceGroupId),
		GcpOverridesObject: convertToObjectStorageGcpOverrides(res.GcpOverride).GcpOverridesToObj(),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, objectStorageAwsOutputs, map[string]attr.Value{
			"s3_bucket_arn": common.DefaultToNull[types.String](res.GetAwsOutputs().GetS3BucketArn()),
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, objectStorageAzureOutputs, map[string]attr.Value{
			"storage_account_id":           common.DefaultToNull[types.String](res.GetAzureOutputs().GetStorageAccountId()),
			"public_storage_container_id":  common.DefaultToNull[types.String](res.GetAzureOutputs().GetPublicStorageContainerId()),
			"private_storage_container_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetPrivateStorageContainerId()),
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, objectStorageGcpOutputs, map[string]attr.Value{
			"storage_bucket_id": common.DefaultToNull[types.String](res.GetGcpOutputs().GetStorageBucketId()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromObjectStorage(plan ObjectStorage) *resourcespb.ObjectStorageArgs {
	return &resourcespb.ObjectStorageArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			ResourceGroupId: plan.ResourceGroupId.ValueString(),
			Location:        plan.Location.Value,
			CloudProvider:   plan.Cloud.Value,
		},
		Name:        plan.Name.ValueString(),
		Versioning:  plan.Versioning.ValueBool(),
		GcpOverride: convertFromObjectStorageGcpOverrides(plan.GetGcpOverrides()),
	}
}

func (v ObjectStorage) UpdatePlan(_ context.Context, config ObjectStorage, p Provider) (ObjectStorage, []path.Path) {
	if config.Cloud.Value != commonpb.CloudProvider_GCP || p.Client.Gcp == nil {
		return v, nil
	}
	var requiresReplace []path.Path
	gcpOverrides := v.GetGcpOverrides()
	if o := config.GetGcpOverrides(); o == nil || o.Project.IsUnknown() {
		if gcpOverrides == nil {
			gcpOverrides = &ObjectStorageGcpOverrides{}
		}

		gcpOverrides.Project = types.StringValue(p.Client.Gcp.Project)

		v.GcpOverridesObject = gcpOverrides.GcpOverridesToObj()
		requiresReplace = append(requiresReplace, path.Root("gcp_overrides").AtName("project"))
	}
	return v, requiresReplace
}

func (v ObjectStorage) GetGcpOverrides() (o *ObjectStorageGcpOverrides) {
	if v.GcpOverridesObject.IsNull() || v.GcpOverridesObject.IsUnknown() {
		return
	}
	o = &ObjectStorageGcpOverrides{
		Project: v.GcpOverridesObject.Attributes()["project"].(types.String),
	}
	return
}

func (o *ObjectStorageGcpOverrides) GcpOverridesToObj() types.Object {
	attrTypes := map[string]attr.Type{
		"project": types.StringType,
	}
	if o == nil {
		return types.ObjectNull(attrTypes)
	}
	result, _ := types.ObjectValue(attrTypes, map[string]attr.Value{"project": o.Project})
	return result
}

type ObjectStorageGcpOverrides struct {
	Project types.String
}

func convertFromObjectStorageGcpOverrides(ref *ObjectStorageGcpOverrides) *resourcespb.ObjectStorageGcpOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.ObjectStorageGcpOverride{Project: ref.Project.ValueString()}
}

func convertToObjectStorageGcpOverrides(ref *resourcespb.ObjectStorageGcpOverride) *ObjectStorageGcpOverrides {
	if ref == nil {
		return nil
	}

	return &ObjectStorageGcpOverrides{Project: common.DefaultToNull[types.String](ref.Project)}
}
