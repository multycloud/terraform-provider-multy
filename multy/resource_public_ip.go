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

type ResourcePublicIpType struct{}

func (r ResourcePublicIpType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Public IP resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:          types.StringType,
				Description:   "Name of Public IP",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
			},
			"network_interface_id": {
				Type:        types.StringType,
				Description: "Associate Public IP to `network_interface` resource",
				Optional:    true,
			},
			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,
		},
	}, nil
}

func (r ResourcePublicIpType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[PublicIp]{
		p:          *(p.(*Provider)),
		createFunc: createPublicIp,
		updateFunc: updatePublicIp,
		readFunc:   readPublicIp,
		deleteFunc: deletePublicIp,
	}, nil
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
		ResourceId: plan.Id.Value,
		Resource:   convertFromPublicIp(plan),
	})
	if err != nil {
		return PublicIp{}, err
	}
	return convertToPublicIp(vn), nil
}

func readPublicIp(ctx context.Context, p Provider, state PublicIp) (PublicIp, error) {
	vn, err := p.Client.Client.ReadPublicIp(ctx, &resourcespb.ReadPublicIpRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return PublicIp{}, err
	}
	return convertToPublicIp(vn), nil
}

func deletePublicIp(ctx context.Context, p Provider, state PublicIp) error {
	_, err := p.Client.Client.DeletePublicIp(ctx, &resourcespb.DeletePublicIpRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type PublicIp struct {
	Id                 types.String                             `tfsdk:"id"`
	Name               types.String                             `tfsdk:"name"`
	NetworkInterfaceId types.String                             `tfsdk:"network_interface_id"`
	Cloud              mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location           mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
}

func convertToPublicIp(res *resourcespb.PublicIpResource) PublicIp {
	return PublicIp{
		Id:                 types.String{Value: res.CommonParameters.ResourceId},
		Name:               types.String{Value: res.Name},
		NetworkInterfaceId: types.String{Value: res.NetworkInterfaceId},
		Cloud:              mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:           mtypes.LocationType.NewVal(res.CommonParameters.Location),
	}
}

func convertFromPublicIp(plan PublicIp) *resourcespb.PublicIpArgs {
	return &resourcespb.PublicIpArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			Location:      plan.Location.Value,
			CloudProvider: plan.Cloud.Value,
		},
		Name:               plan.Name.Value,
		NetworkInterfaceId: plan.NetworkInterfaceId.Value,
	}
}
