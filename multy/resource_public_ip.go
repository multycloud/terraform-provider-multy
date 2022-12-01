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

type ResourcePublicIpType struct{}

var publicIpAwsOutputs = map[string]attr.Type{
	"public_ip_id": types.StringType,
}

var publicIpAzureOutputs = map[string]attr.Type{
	"public_ip_id": types.StringType,
}

var publicIpGcpOutputs = map[string]attr.Type{
	"compute_address_id": types.StringType,
}

var publicIpSchema = tfsdk.Schema{
	MarkdownDescription: "Provides Multy Public IP resource",
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
			Description:   "Name of Public IP",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
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
			Type:        types.ObjectType{AttrTypes: publicIpAwsOutputs},
			Computed:    true,
		},
		"azure": {
			Description: "Azure-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: publicIpAzureOutputs},
			Computed:    true,
		},
		"gcp": {
			Description: "GCP-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: publicIpGcpOutputs},
			Computed:    true,
		},
		"resource_status": common.ResourceStatusSchema,
	},
}

func (r ResourcePublicIpType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[PublicIp]{
		p:          *(p.(*Provider)),
		createFunc: createPublicIp,
		updateFunc: updatePublicIp,
		readFunc:   readPublicIp,
		deleteFunc: deletePublicIp,
		name:       "multy_public_ip",
		schema:     publicIpSchema,
	}
}

func createPublicIp(ctx context.Context, p Provider, plan PublicIp) (PublicIp, error) {
	vn, err := p.Client.Client.CreatePublicIp(ctx, &resourcespb.CreatePublicIpRequest{
		Resource: convertFromPublicIp(plan),
	})
	if err != nil {
		return PublicIp{}, err
	}
	return convertToPublicIp(vn), nil
}

func updatePublicIp(ctx context.Context, p Provider, plan PublicIp) (PublicIp, error) {
	vn, err := p.Client.Client.UpdatePublicIp(ctx, &resourcespb.UpdatePublicIpRequest{
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromPublicIp(plan),
	})
	if err != nil {
		return PublicIp{}, err
	}
	return convertToPublicIp(vn), nil
}

func readPublicIp(ctx context.Context, p Provider, state PublicIp) (PublicIp, error) {
	vn, err := p.Client.Client.ReadPublicIp(ctx, &resourcespb.ReadPublicIpRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return PublicIp{}, err
	}
	return convertToPublicIp(vn), nil
}

func deletePublicIp(ctx context.Context, p Provider, state PublicIp) error {
	_, err := p.Client.Client.DeletePublicIp(ctx, &resourcespb.DeletePublicIpRequest{
		ResourceId: state.Id.ValueString(),
	})
	return err
}

type PublicIp struct {
	Id              types.String                             `tfsdk:"id"`
	Name            types.String                             `tfsdk:"name"`
	Cloud           mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location        mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
	ResourceGroupId types.String                             `tfsdk:"resource_group_id"`

	GcpOverridesObject types.Object `tfsdk:"gcp_overrides"`
	AwsOutputs         types.Object `tfsdk:"aws"`
	AzureOutputs       types.Object `tfsdk:"azure"`
	GcpOutputs         types.Object `tfsdk:"gcp"`
	ResourceStatus     types.Map    `tfsdk:"resource_status"`
}

func convertToPublicIp(res *resourcespb.PublicIpResource) PublicIp {
	return PublicIp{
		Id:                 types.StringValue(res.CommonParameters.ResourceId),
		Name:               types.StringValue(res.Name),
		Cloud:              mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:           mtypes.LocationType.NewVal(res.CommonParameters.Location),
		ResourceGroupId:    types.StringValue(res.CommonParameters.ResourceGroupId),
		GcpOverridesObject: convertToPublicIpGcpOverrides(res.GcpOverride).GcpOverridesToObj(),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, publicIpAwsOutputs, map[string]attr.Value{
			"public_ip_id": common.DefaultToNull[types.String](res.GetAwsOutputs().GetPublicIpId()),
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, publicIpAzureOutputs, map[string]attr.Value{
			"public_ip_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetPublicIpId()),
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, publicIpGcpOutputs, map[string]attr.Value{
			"compute_address_id": common.DefaultToNull[types.String](res.GetGcpOutputs().GetComputeAddressId()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromPublicIp(plan PublicIp) *resourcespb.PublicIpArgs {
	return &resourcespb.PublicIpArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			ResourceGroupId: plan.ResourceGroupId.ValueString(),
			Location:        plan.Location.Value,
			CloudProvider:   plan.Cloud.Value,
		},
		Name:        plan.Name.ValueString(),
		GcpOverride: convertFromPublicIpGcpOverrides(plan.GetGcpOverrides()),
	}
}

func convertFromPublicIpGcpOverrides(ref *PublicIpGcpOverrides) *resourcespb.PublicIpGcpOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.PublicIpGcpOverride{Project: ref.Project.ValueString()}
}

func convertToPublicIpGcpOverrides(ref *resourcespb.PublicIpGcpOverride) *PublicIpGcpOverrides {
	if ref == nil {
		return nil
	}

	return &PublicIpGcpOverrides{Project: common.DefaultToNull[types.String](ref.Project)}
}

func (v PublicIp) GetGcpOverrides() (o *PublicIpGcpOverrides) {
	if v.GcpOverridesObject.IsNull() || v.GcpOverridesObject.IsUnknown() {
		return
	}
	o = &PublicIpGcpOverrides{
		Project: v.GcpOverridesObject.Attributes()["project"].(types.String),
	}
	return
}

func (o *PublicIpGcpOverrides) GcpOverridesToObj() types.Object {
	attrTypes := map[string]attr.Type{
		"project": types.StringType,
	}
	if o == nil {
		return types.ObjectNull(attrTypes)
	}
	result, _ := types.ObjectValue(attrTypes, map[string]attr.Value{"project": o.Project})
	return result
}

type PublicIpGcpOverrides struct {
	Project types.String
}

func (v PublicIp) UpdatePlan(_ context.Context, config PublicIp, p Provider) (PublicIp, []path.Path) {
	if config.Cloud.Value != commonpb.CloudProvider_GCP || p.Client.Gcp == nil {
		return v, nil
	}
	var requiresReplace []path.Path
	gcpOverrides := v.GetGcpOverrides()
	if o := config.GetGcpOverrides(); o == nil || o.Project.IsUnknown() {
		if gcpOverrides == nil {
			gcpOverrides = &PublicIpGcpOverrides{}
		}

		gcpOverrides.Project = types.StringValue(p.Client.Gcp.Project)

		v.GcpOverridesObject = gcpOverrides.GcpOverridesToObj()
		requiresReplace = append(requiresReplace, path.Root("gcp_overrides").AtName("project"))
	}
	return v, requiresReplace
}
