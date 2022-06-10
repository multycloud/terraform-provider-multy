package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
)

type ResourceNetworkInterfaceSecurityGroupAssociationType struct{}

func (r ResourceNetworkInterfaceSecurityGroupAssociationType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Network Interface Security Group Association resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"network_interface_id": {
				Type:          types.StringType,
				Description:   "ID of `network_interface` resource",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"security_group_id": {
				Type:          types.StringType,
				Description:   "ID of `security_group` resource",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
		},
	}, nil
}

func (r ResourceNetworkInterfaceSecurityGroupAssociationType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[NetworkInterfaceSecurityGroupAssociation]{
		p:          *(p.(*Provider)),
		createFunc: createNetworkInterfaceSecurityGroupAssociation,
		updateFunc: updateNetworkInterfaceSecurityGroupAssociation,
		readFunc:   readNetworkInterfaceSecurityGroupAssociation,
		deleteFunc: deleteNetworkInterfaceSecurityGroupAssociation,
	}, nil
}

func createNetworkInterfaceSecurityGroupAssociation(ctx context.Context, p Provider, plan NetworkInterfaceSecurityGroupAssociation) (NetworkInterfaceSecurityGroupAssociation, error) {
	vn, err := p.Client.Client.CreateNetworkInterfaceSecurityGroupAssociation(ctx, &resourcespb.CreateNetworkInterfaceSecurityGroupAssociationRequest{
		Resource: convertFromNetworkInterfaceSecurityGroupAssociation(plan),
	})
	if err != nil {
		return NetworkInterfaceSecurityGroupAssociation{}, err
	}
	return convertToNetworkInterfaceSecurityGroupAssociation(vn), nil
}

func updateNetworkInterfaceSecurityGroupAssociation(ctx context.Context, p Provider, plan NetworkInterfaceSecurityGroupAssociation) (NetworkInterfaceSecurityGroupAssociation, error) {
	vn, err := p.Client.Client.UpdateNetworkInterfaceSecurityGroupAssociation(ctx, &resourcespb.UpdateNetworkInterfaceSecurityGroupAssociationRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromNetworkInterfaceSecurityGroupAssociation(plan),
	})
	if err != nil {
		return NetworkInterfaceSecurityGroupAssociation{}, err
	}
	return convertToNetworkInterfaceSecurityGroupAssociation(vn), nil
}

func readNetworkInterfaceSecurityGroupAssociation(ctx context.Context, p Provider, state NetworkInterfaceSecurityGroupAssociation) (NetworkInterfaceSecurityGroupAssociation, error) {
	vn, err := p.Client.Client.ReadNetworkInterfaceSecurityGroupAssociation(ctx, &resourcespb.ReadNetworkInterfaceSecurityGroupAssociationRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return NetworkInterfaceSecurityGroupAssociation{}, err
	}
	return convertToNetworkInterfaceSecurityGroupAssociation(vn), nil
}

func deleteNetworkInterfaceSecurityGroupAssociation(ctx context.Context, p Provider, state NetworkInterfaceSecurityGroupAssociation) error {
	_, err := p.Client.Client.DeleteNetworkInterfaceSecurityGroupAssociation(ctx, &resourcespb.DeleteNetworkInterfaceSecurityGroupAssociationRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type NetworkInterfaceSecurityGroupAssociation struct {
	Id                 types.String `tfsdk:"id"`
	NetworkInterfaceId types.String `tfsdk:"network_interface_id"`
	SecurityGroupId    types.String `tfsdk:"security_group_id"`
}

func convertToNetworkInterfaceSecurityGroupAssociation(res *resourcespb.NetworkInterfaceSecurityGroupAssociationResource) NetworkInterfaceSecurityGroupAssociation {
	return NetworkInterfaceSecurityGroupAssociation{
		Id:                 types.String{Value: res.CommonParameters.ResourceId},
		NetworkInterfaceId: types.String{Value: res.NetworkInterfaceId},
		SecurityGroupId:    types.String{Value: res.SecurityGroupId},
	}
}

func convertFromNetworkInterfaceSecurityGroupAssociation(plan NetworkInterfaceSecurityGroupAssociation) *resourcespb.NetworkInterfaceSecurityGroupAssociationArgs {
	return &resourcespb.NetworkInterfaceSecurityGroupAssociationArgs{
		NetworkInterfaceId: plan.NetworkInterfaceId.Value,
		SecurityGroupId:    plan.SecurityGroupId.Value,
	}
}
