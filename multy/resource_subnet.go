package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/validators"
)

type ResourceSubnetType struct{}

func (r ResourceSubnetType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Subnet resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of Subnet",
				Required:    true,
				//PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"cidr_block": {
				Type:        types.StringType,
				Description: "CIDR block of Subnet",
				Required:    true,
				Validators:  []tfsdk.AttributeValidator{validators.IsCidrValidator{}},
				//PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("aws")},
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"virtual_network_id": {
				Type:          types.StringType,
				Description:   "ID of `virtual_network` resource",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
		},
	}, nil
}

func (r ResourceSubnetType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[Subnet]{
		p:          *(p.(*Provider)),
		createFunc: createSubnet,
		updateFunc: updateSubnet,
		readFunc:   readSubnet,
		deleteFunc: deleteSubnet,
	}, nil
}

func createSubnet(ctx context.Context, p Provider, plan Subnet) (Subnet, error) {
	vn, err := p.Client.Client.CreateSubnet(ctx, &resourcespb.CreateSubnetRequest{
		Resource: convertFromSubnet(plan),
	})
	if err != nil {
		return Subnet{}, err
	}
	return convertToSubnet(vn), nil
}

func updateSubnet(ctx context.Context, p Provider, plan Subnet) (Subnet, error) {
	vn, err := p.Client.Client.UpdateSubnet(ctx, &resourcespb.UpdateSubnetRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromSubnet(plan),
	})
	if err != nil {
		return Subnet{}, err
	}
	return convertToSubnet(vn), nil
}

func readSubnet(ctx context.Context, p Provider, state Subnet) (Subnet, error) {
	vn, err := p.Client.Client.ReadSubnet(ctx, &resourcespb.ReadSubnetRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return Subnet{}, err
	}
	return convertToSubnet(vn), nil
}

func deleteSubnet(ctx context.Context, p Provider, state Subnet) error {
	_, err := p.Client.Client.DeleteSubnet(ctx, &resourcespb.DeleteSubnetRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type Subnet struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	CidrBlock        types.String `tfsdk:"cidr_block"`
	VirtualNetworkId types.String `tfsdk:"virtual_network_id"`
}

func convertToSubnet(res *resourcespb.SubnetResource) Subnet {
	result := Subnet{
		Id:               types.String{Value: res.CommonParameters.ResourceId},
		Name:             types.String{Value: res.Name},
		CidrBlock:        types.String{Value: res.CidrBlock},
		VirtualNetworkId: types.String{Value: res.VirtualNetworkId},
	}

	return result
}

func convertFromSubnet(plan Subnet) *resourcespb.SubnetArgs {
	return &resourcespb.SubnetArgs{
		Name:             plan.Name.Value,
		CidrBlock:        plan.CidrBlock.Value,
		VirtualNetworkId: plan.VirtualNetworkId.Value,
	}
}
