package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceKubernetesNodePoolType struct{}

func (r ResourceKubernetesNodePoolType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	attrs := getKubernetesNodePoolAttrs()
	attrs["cluster_id"] =
		tfsdk.Attribute{
			Type:          types.StringType,
			Description:   "Id of the multy kubernetes cluster",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
		}
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Object Storage Object resource",
		Attributes:          attrs,
	}, nil
}

func getKubernetesNodePoolAttrs() map[string]tfsdk.Attribute {
	return map[string]tfsdk.Attribute{
		"id": {
			Type:          types.StringType,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
		},
		"name": {
			Type:          types.StringType,
			Description:   "Name of kubernetes node pool",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
		},
		"subnet_ids": {
			Type:          types.ListType{ElemType: types.StringType},
			Description:   "Subnets associated with this cluster. At least one must be public.",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
		},
		"starting_node_count": {
			Type:          types.Int64Type,
			Description:   "Number of initial nodes. Defaults to the minimum number of nodes.",
			Computed:      true,
			Optional:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
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
			PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
		},
		"disk_size_gb": {
			Type:          types.Int64Type,
			Description:   "Disk size used for each node.",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
		},
		"labels": {
			Type:        types.MapType{ElemType: types.StringType},
			Description: "Labels to be applied to each node.",
			Optional:    true,
			Computed:    true,
		},

		"cluster_id": {
			Type:          types.StringType,
			Description:   "Id of the multy kubernetes cluster",
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
		},
	}
}

func (r ResourceKubernetesNodePoolType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[KubernetesNodePool]{
		p:          *(p.(*Provider)),
		createFunc: createKubernetesNodePool,
		updateFunc: updateKubernetesNodePool,
		readFunc:   readKubernetesNodePool,
		deleteFunc: deleteKubernetesNodePool,
	}, nil
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
		ResourceId: plan.Id.Value,
		Resource:   convertFromKubernetesNodePool(plan),
	})
	if err != nil {
		return KubernetesNodePool{}, err
	}
	return convertToKubernetesNodePool(vn), nil
}

func readKubernetesNodePool(ctx context.Context, p Provider, state KubernetesNodePool) (KubernetesNodePool, error) {
	vn, err := p.Client.Client.ReadKubernetesNodePool(ctx, &resourcespb.ReadKubernetesNodePoolRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return KubernetesNodePool{}, err
	}
	return convertToKubernetesNodePool(vn), nil
}

func deleteKubernetesNodePool(ctx context.Context, p Provider, state KubernetesNodePool) error {
	_, err := p.Client.Client.DeleteKubernetesNodePool(ctx, &resourcespb.DeleteKubernetesNodePoolRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type KubernetesNodePool struct {
	Id                types.String                           `tfsdk:"id"`
	ClusterId         types.String                           `tfsdk:"cluster_id"`
	Name              types.String                           `tfsdk:"name"`
	VmSize            mtypes.EnumValue[commonpb.VmSize_Enum] `tfsdk:"vm_size"`
	SubnetIds         []types.String                         `tfsdk:"subnet_ids"`
	StartingNodeCount types.Int64                            `tfsdk:"starting_node_count"`
	MinNodeCount      types.Int64                            `tfsdk:"min_node_count"`
	MaxNodeCount      types.Int64                            `tfsdk:"max_node_count"`
	DiskSizeGb        types.Int64                            `tfsdk:"disk_size_gb"`
	Labels            types.Map                              `tfsdk:"labels"`
}

func convertToKubernetesNodePool(res *resourcespb.KubernetesNodePoolResource) KubernetesNodePool {
	return KubernetesNodePool{
		Id:                types.String{Value: res.CommonParameters.ResourceId},
		Name:              types.String{Value: res.Name},
		VmSize:            mtypes.VmSizeType.NewVal(res.VmSize),
		SubnetIds:         common.TypesStringToStringSlice(res.SubnetIds),
		StartingNodeCount: common.DefaultToNull[types.Int64](int64(res.StartingNodeCount)),
		MinNodeCount:      types.Int64{Value: int64(res.MinNodeCount)},
		MaxNodeCount:      types.Int64{Value: int64(res.MaxNodeCount)},
		DiskSizeGb:        types.Int64{Value: res.DiskSizeGb},
		Labels:            common.GoMapToMapType(res.Labels),
		ClusterId:         types.String{Value: res.ClusterId},
	}
}

func convertFromKubernetesNodePool(plan KubernetesNodePool) *resourcespb.KubernetesNodePoolArgs {
	var clusterId string
	if !plan.ClusterId.Unknown {
		clusterId = plan.ClusterId.Value
	}
	return &resourcespb.KubernetesNodePoolArgs{
		Name:              plan.Name.Value,
		SubnetIds:         common.StringSliceToTypesString(plan.SubnetIds),
		ClusterId:         clusterId,
		StartingNodeCount: int32(plan.StartingNodeCount.Value),
		MinNodeCount:      int32(plan.MinNodeCount.Value),
		MaxNodeCount:      int32(plan.MaxNodeCount.Value),
		VmSize:            plan.VmSize.Value,
		DiskSizeGb:        plan.DiskSizeGb.Value,
		Labels:            common.MapTypeToGoMap(plan.Labels),
	}
}
