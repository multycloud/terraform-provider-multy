package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceKubernetesNodePoolType struct{}

var kubernetesNodePoolAwsOutputs = map[string]attr.Type{
	"eks_node_pool_id": types.StringType,
	"iam_role_arn":     types.StringType,
}

var kubernetesNodePoolAzureOutputs = map[string]attr.Type{
	"aks_node_pool_id": types.StringType,
}

var kubernetesNodePoolGcpOutputs = map[string]attr.Type{
	"gke_node_pool_id": types.StringType,
}

func getKubernetesNodePoolAttrs() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Type:          types.StringType,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		},
		"name": {
			Type:          types.StringType,
			Description:   "Name of kubernetes node pool",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"subnet_id": {
			Type:          types.StringType,
			Description:   "Subnet to place the node and pods in. Must have access to the Internet to connect with the control plane.",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"starting_node_count": {
			Type:          types.Int64Type,
			Description:   "Number of initial nodes. Defaults to the minimum number of nodes.",
			Computed:      true,
			Optional:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		},
		"min_node_count": {
			Type:        types.Int64Type,
			Description: "Minimum number of nodes.",
			Required:    true,
		},
		"max_node_count": {
			Type:        types.Int64Type,
			Description: "Maximum number of nodes.",
			Required:    true,
		},
		"vm_size": {
			Type:          mtypes.VmSizeType,
			Description:   fmt.Sprintf("Size of Virtual Machine used for the nodes. Accepted values are %s", common.StringSliceToDocsMarkdown(mtypes.VmSizeType.GetAllValues())),
			Required:      true,
			Validators:    []tfsdk.AttributeValidator{validators.NewValidator(mtypes.VmSizeType)},
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"disk_size_gb": {
			Type:          types.Int64Type,
			Description:   "Disk size used for each node.",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"labels": {
			Type:        types.MapType{ElemType: types.StringType},
			Description: "Labels to be applied to each node.",
			Optional:    true,
			Computed:    true,
		},

		"availability_zones": {
			Type:          types.ListType{ElemType: types.Int64Type},
			Description:   "Zones to place nodes in. If not set, they will be spread across multiple zones selected by the cloud provider.",
			Optional:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},

		"cluster_id": {
			Type:          types.StringType,
			Description:   "Id of the multy kubernetes cluster",
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"aws_overrides": {
			Description: "AWS-specific attributes that will be set if this resource is deployed in AWS",
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"instance_types": {
					Type:          types.ListType{ElemType: types.StringType},
					Description:   fmt.Sprintf("The instance type to use for nodes."),
					Optional:      true,
					PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
				},
			}),
			Optional: true,
			//PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("aws")},
		},
		"azure_overrides": {
			Description: "Azure-specific attributes that will be set if this resource is deployed in Azure",
			Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
				"vm_size": {
					Type:          types.StringType,
					Description:   fmt.Sprintf("The size to use for nodes."),
					Optional:      true,
					PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
					Validators:    []tfsdk.AttributeValidator{mtypes.NonEmptyStringValidator},
				},
			}),
			Optional: true,
			//PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
		},
		"aws": {
			Description: "AWS-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: kubernetesNodePoolAwsOutputs},
			Computed:    true,
		},
		"azure": {
			Description: "Azure-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: kubernetesNodePoolAzureOutputs},
			Computed:    true,
		},
		"gcp": {
			Description: "GCP-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: kubernetesNodePoolGcpOutputs},
			Computed:    true,
		},
		"resource_status": common.ResourceStatusSchema,
	}
}

func (r ResourceKubernetesNodePoolType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	attrs := getKubernetesNodePoolAttrs()
	attrs["cluster_id"] =
		tfsdk.Attribute{
			Type:          types.StringType,
			Description:   "Id of the multy kubernetes cluster",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		}

	return MultyResource[KubernetesNodePool]{
		p:          *(p.(*Provider)),
		createFunc: createKubernetesNodePool,
		updateFunc: updateKubernetesNodePool,
		readFunc:   readKubernetesNodePool,
		deleteFunc: deleteKubernetesNodePool,
		name:       "multy_kubernetes_node_pool",
		schema: tfsdk.Schema{
			MarkdownDescription: "Provides Multy Object Storage Object resource",
			Attributes:          attrs,
		},
	}
}

func createKubernetesNodePool(ctx context.Context, p Provider, plan KubernetesNodePool) (KubernetesNodePool, error) {
	vn, err := p.Client.Client.CreateKubernetesNodePool(ctx, &resourcespb.CreateKubernetesNodePoolRequest{
		Resource: convertFromKubernetesNodePool(plan),
	})
	if err != nil {
		return KubernetesNodePool{}, err
	}
	return convertToKubernetesNodePool(vn), nil
}

func updateKubernetesNodePool(ctx context.Context, p Provider, plan KubernetesNodePool) (KubernetesNodePool, error) {
	vn, err := p.Client.Client.UpdateKubernetesNodePool(ctx, &resourcespb.UpdateKubernetesNodePoolRequest{
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromKubernetesNodePool(plan),
	})
	if err != nil {
		return KubernetesNodePool{}, err
	}
	return convertToKubernetesNodePool(vn), nil
}

func readKubernetesNodePool(ctx context.Context, p Provider, state KubernetesNodePool) (KubernetesNodePool, error) {
	vn, err := p.Client.Client.ReadKubernetesNodePool(ctx, &resourcespb.ReadKubernetesNodePoolRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return KubernetesNodePool{}, err
	}
	return convertToKubernetesNodePool(vn), nil
}

func deleteKubernetesNodePool(ctx context.Context, p Provider, state KubernetesNodePool) error {
	_, err := p.Client.Client.DeleteKubernetesNodePool(ctx, &resourcespb.DeleteKubernetesNodePoolRequest{
		ResourceId: state.Id.ValueString(),
	})
	return err
}

type KubernetesNodePool struct {
	Id                types.String                           `tfsdk:"id"`
	ClusterId         types.String                           `tfsdk:"cluster_id"`
	Name              types.String                           `tfsdk:"name"`
	VmSize            mtypes.EnumValue[commonpb.VmSize_Enum] `tfsdk:"vm_size"`
	SubnetId          types.String                           `tfsdk:"subnet_id"`
	StartingNodeCount types.Int64                            `tfsdk:"starting_node_count"`
	MinNodeCount      types.Int64                            `tfsdk:"min_node_count"`
	MaxNodeCount      types.Int64                            `tfsdk:"max_node_count"`
	DiskSizeGb        types.Int64                            `tfsdk:"disk_size_gb"`
	Labels            types.Map                              `tfsdk:"labels"`
	AvailabilityZones []types.Int64                          `tfsdk:"availability_zones"`
	AwsOverrides      *KubernetesNodePoolAwsOverrides        `tfsdk:"aws_overrides"`
	AzureOverrides    *KubernetesNodePoolAzureOverrides      `tfsdk:"azure_overrides"`
	AwsOutputs        types.Object                           `tfsdk:"aws"`
	AzureOutputs      types.Object                           `tfsdk:"azure"`
	GcpOutputs        types.Object                           `tfsdk:"gcp"`
	ResourceStatus    types.Map                              `tfsdk:"resource_status"`
}

func convertToKubernetesNodePool(res *resourcespb.KubernetesNodePoolResource) KubernetesNodePool {
	return KubernetesNodePool{
		Id:                types.StringValue(res.CommonParameters.ResourceId),
		ClusterId:         types.StringValue(res.ClusterId),
		Name:              types.StringValue(res.Name),
		VmSize:            mtypes.VmSizeType.NewVal(res.VmSize),
		SubnetId:          types.StringValue(res.SubnetId),
		StartingNodeCount: common.DefaultToNull[types.Int64](int64(res.StartingNodeCount)),
		MinNodeCount:      types.Int64Value(int64(res.MinNodeCount)),
		MaxNodeCount:      types.Int64Value(int64(res.MaxNodeCount)),
		DiskSizeGb:        types.Int64Value(res.DiskSizeGb),
		Labels:            common.GoMapToMapType(res.Labels),
		AvailabilityZones: common.GoIntToTfInt(res.AvailabilityZone),
		AwsOverrides:      convertToKubernetesNodePoolAwsOverrides(res.AwsOverride),
		AzureOverrides:    convertToKubernetesNodePoolAzureOverrides(res.AzureOverride),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, kubernetesNodePoolAwsOutputs, map[string]attr.Value{
			"eks_node_pool_id": common.DefaultToNull[types.String](res.GetAwsOutputs().GetEksNodePoolId()),
			"iam_role_arn":     common.DefaultToNull[types.String](res.GetAwsOutputs().GetIamRoleArn()),
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, kubernetesNodePoolAzureOutputs, map[string]attr.Value{
			"aks_node_pool_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetAksNodePoolId()),
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, kubernetesNodePoolGcpOutputs, map[string]attr.Value{
			"gke_node_pool_id": common.DefaultToNull[types.String](res.GetGcpOutputs().GetGkeNodePoolId()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromKubernetesNodePool(plan KubernetesNodePool) *resourcespb.KubernetesNodePoolArgs {
	var clusterId string
	if !plan.ClusterId.IsUnknown() {
		clusterId = plan.ClusterId.ValueString()
	}
	return &resourcespb.KubernetesNodePoolArgs{
		Name:              plan.Name.ValueString(),
		SubnetId:          plan.SubnetId.ValueString(),
		ClusterId:         clusterId,
		StartingNodeCount: int32(plan.StartingNodeCount.ValueInt64()),
		MinNodeCount:      int32(plan.MinNodeCount.ValueInt64()),
		MaxNodeCount:      int32(plan.MaxNodeCount.ValueInt64()),
		VmSize:            plan.VmSize.Value,
		DiskSizeGb:        plan.DiskSizeGb.ValueInt64(),
		AwsOverride:       convertFromKubernetesNodePoolAwsOverrides(plan.AwsOverrides),
		AzureOverride:     convertFromKubernetesNodePoolAzureOverrides(plan.AzureOverrides),
		Labels:            common.MapTypeToGoMap(plan.Labels),
		AvailabilityZone:  common.TfIntToGoInt(plan.AvailabilityZones),
	}
}

type KubernetesNodePoolAwsOverrides struct {
	InstanceTypes []types.String `tfsdk:"instance_types"`
}
type KubernetesNodePoolAzureOverrides struct {
	VmSize types.String `tfsdk:"vm_size"`
}

func convertFromKubernetesNodePoolAwsOverrides(ref *KubernetesNodePoolAwsOverrides) *resourcespb.KubernetesNodePoolAwsOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.KubernetesNodePoolAwsOverride{
		InstanceTypes: common.StringSliceToTypesString(ref.InstanceTypes),
	}
}

func convertToKubernetesNodePoolAwsOverrides(ref *resourcespb.KubernetesNodePoolAwsOverride) *KubernetesNodePoolAwsOverrides {
	if ref == nil {
		return nil
	}

	return &KubernetesNodePoolAwsOverrides{InstanceTypes: common.TypesStringToStringSlice(ref.InstanceTypes)}
}
func convertFromKubernetesNodePoolAzureOverrides(ref *KubernetesNodePoolAzureOverrides) *resourcespb.KubernetesNodePoolAzureOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.KubernetesNodePoolAzureOverride{
		VmSize: ref.VmSize.ValueString(),
	}
}

func convertToKubernetesNodePoolAzureOverrides(ref *resourcespb.KubernetesNodePoolAzureOverride) *KubernetesNodePoolAzureOverrides {
	if ref == nil {
		return nil
	}

	return &KubernetesNodePoolAzureOverrides{VmSize: common.DefaultToNull[types.String](ref.VmSize)}
}
