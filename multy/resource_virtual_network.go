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
	"terraform-provider-multy/multy/validators"
)

type ResourceVirtualNetworkType struct{}

func (r ResourceVirtualNetworkType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Virtual Network resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"resource_group_id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:          types.StringType,
				Description:   "Name of Virtual Network",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
			},
			"cidr_block": {
				Type:          types.StringType,
				Description:   "CIDR Block of Virtual Network",
				Required:      true,
				Validators:    []tfsdk.AttributeValidator{validators.IsCidrValidator{}},
				PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("aws")},
			},
			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,
		},
	}, nil
}

func (r ResourceVirtualNetworkType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[VirtualNetwork]{
		p:          *(p.(*Provider)),
		createFunc: createVirtualNetwork,
		updateFunc: updateVirtualNetwork,
		readFunc:   readVirtualNetwork,
		deleteFunc: deleteVirtualNetwork,
	}, nil
}

func createVirtualNetwork(ctx context.Context, p Provider, plan VirtualNetwork) (VirtualNetwork, error) {
	vn, err := p.Client.Client.CreateVirtualNetwork(ctx, &resourcespb.CreateVirtualNetworkRequest{
		Resource: convertFromVirtualNetwork(plan),
	})
	if err != nil {
		return VirtualNetwork{}, err
	}
	return convertToVirtualNetwork(vn), nil
}

func updateVirtualNetwork(ctx context.Context, p Provider, plan VirtualNetwork) (VirtualNetwork, error) {
	vn, err := p.Client.Client.UpdateVirtualNetwork(ctx, &resourcespb.UpdateVirtualNetworkRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromVirtualNetwork(plan),
	})
	if err != nil {
		return VirtualNetwork{}, err
	}
	return convertToVirtualNetwork(vn), nil
}

func readVirtualNetwork(ctx context.Context, p Provider, state VirtualNetwork) (VirtualNetwork, error) {
	vn, err := p.Client.Client.ReadVirtualNetwork(ctx, &resourcespb.ReadVirtualNetworkRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return VirtualNetwork{}, err
	}
	return convertToVirtualNetwork(vn), nil
}

func deleteVirtualNetwork(ctx context.Context, p Provider, state VirtualNetwork) error {
	_, err := p.Client.Client.DeleteVirtualNetwork(ctx, &resourcespb.DeleteVirtualNetworkRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type VirtualNetwork struct {
	Id              types.String                             `tfsdk:"id"`
	ResourceGroupId types.String                             `tfsdk:"resource_group_id"`
	Name            types.String                             `tfsdk:"name"`
	CidrBlock       types.String                             `tfsdk:"cidr_block"`
	Cloud           mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location        mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
}

func convertToVirtualNetwork(res *resourcespb.VirtualNetworkResource) VirtualNetwork {
	return VirtualNetwork{
		Id:              types.String{Value: res.CommonParameters.ResourceId},
		ResourceGroupId: types.String{Value: res.CommonParameters.ResourceGroupId},
		Name:            types.String{Value: res.Name},
		CidrBlock:       types.String{Value: res.CidrBlock},
		Cloud:           mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:        mtypes.LocationType.NewVal(res.CommonParameters.Location),
	}
}

func convertFromVirtualNetwork(plan VirtualNetwork) *resourcespb.VirtualNetworkArgs {
	return &resourcespb.VirtualNetworkArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			ResourceGroupId: plan.ResourceGroupId.Value,
			Location:        plan.Location.Value,
			CloudProvider:   plan.Cloud.Value,
		},
		Name:      plan.Name.Value,
		CidrBlock: plan.CidrBlock.Value,
	}
}
