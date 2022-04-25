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

type ResourceKubernetesClusterType struct{}

func (r ResourceKubernetesClusterType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Kubernetes Cluster resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of the cluster",
				Required:    true,
			},
			"subnet_ids": {
				Type:          types.ListType{ElemType: types.StringType},
				Description:   "Subnets associated with this cluster. At least one must be public.",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"default_node_pool": {
				Attributes:  tfsdk.SingleNestedAttributes(getKubernetesNodePoolAttrs()),
				Description: "Default node pool to associate with this cluster.",
				Required:    true,
			},
			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,
		},
	}, nil
}

func (r ResourceKubernetesClusterType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[KubernetesCluster]{
		p:          *(p.(*Provider)),
		createFunc: createKubernetesCluster,
		updateFunc: updateKubernetesCluster,
		readFunc:   readKubernetesCluster,
		deleteFunc: deleteKubernetesCluster,
	}, nil
}

func createKubernetesCluster(ctx context.Context, p Provider, plan KubernetesCluster) (KubernetesCluster, error) {
	vn, err := p.Client.Client.CreateKubernetesCluster(ctx, &resourcespb.CreateKubernetesClusterRequest{
		Resource: convertFromKubernetesCluster(plan),
	})
	if err != nil {
		return KubernetesCluster{}, err
	}
	return convertToKubernetesCluster(vn), nil
}

func updateKubernetesCluster(ctx context.Context, p Provider, plan KubernetesCluster) (KubernetesCluster, error) {
	vn, err := p.Client.Client.UpdateKubernetesCluster(ctx, &resourcespb.UpdateKubernetesClusterRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromKubernetesCluster(plan),
	})
	if err != nil {
		return KubernetesCluster{}, err
	}
	return convertToKubernetesCluster(vn), nil
}

func readKubernetesCluster(ctx context.Context, p Provider, state KubernetesCluster) (KubernetesCluster, error) {
	vn, err := p.Client.Client.ReadKubernetesCluster(ctx, &resourcespb.ReadKubernetesClusterRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return KubernetesCluster{}, err
	}
	return convertToKubernetesCluster(vn), nil
}

func deleteKubernetesCluster(ctx context.Context, p Provider, state KubernetesCluster) error {
	_, err := p.Client.Client.DeleteKubernetesCluster(ctx, &resourcespb.DeleteKubernetesClusterRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type KubernetesCluster struct {
	Id        types.String                             `tfsdk:"id"`
	Name      types.String                             `tfsdk:"name"`
	SubnetIds []types.String                           `tfsdk:"subnet_ids"`
	Cloud     mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location  mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`

	DefaultNodePool KubernetesNodePool `tfsdk:"default_node_pool"`
}

func convertToKubernetesCluster(res *resourcespb.KubernetesClusterResource) KubernetesCluster {
	return KubernetesCluster{
		Id:              types.String{Value: res.CommonParameters.ResourceId},
		Name:            types.String{Value: res.Name},
		SubnetIds:       common.TypesStringToStringSlice(res.SubnetIds),
		Cloud:           mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:        mtypes.LocationType.NewVal(res.CommonParameters.Location),
		DefaultNodePool: convertToKubernetesNodePool(res.GetDefaultNodePool()),
	}
}

func convertFromKubernetesCluster(plan KubernetesCluster) *resourcespb.KubernetesClusterArgs {
	return &resourcespb.KubernetesClusterArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			Location:      plan.Location.Value,
			CloudProvider: plan.Cloud.Value,
		},
		Name:            plan.Name.Value,
		SubnetIds:       common.StringSliceToTypesString(plan.SubnetIds),
		DefaultNodePool: convertFromKubernetesNodePool(plan.DefaultNodePool),
	}
}
