package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
)

type ResourceNetworkInterfaceType struct{}

func (r ResourceNetworkInterfaceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Network Interface resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of Network Interface",
				Required:    true,
			},
			"subnet_id": {
				Type:        types.StringType,
				Description: "ID of `subnet` resource",
				Required:    true,
			},
			"cloud": common.CloudsSchema,
		},
	}, nil
}

func (r ResourceNetworkInterfaceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[NetworkInterface]{
		p:          *(p.(*Provider)),
		createFunc: createNetworkInterface,
		updateFunc: updateNetworkInterface,
		readFunc:   readNetworkInterface,
		deleteFunc: deleteNetworkInterface,
	}, nil
}

func createNetworkInterface(ctx context.Context, p Provider, plan NetworkInterface) (NetworkInterface, error) {
	vn, err := p.Client.Client.CreateNetworkInterface(ctx, &resourcespb.CreateNetworkInterfaceRequest{
		Resource: convertFromNetworkInterface(plan),
	})
	if err != nil {
		return NetworkInterface{}, err
	}
	return convertToNetworkInterface(vn), nil
}

func updateNetworkInterface(ctx context.Context, p Provider, plan NetworkInterface) (NetworkInterface, error) {
	vn, err := p.Client.Client.UpdateNetworkInterface(ctx, &resourcespb.UpdateNetworkInterfaceRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromNetworkInterface(plan),
	})
	if err != nil {
		return NetworkInterface{}, err
	}
	return convertToNetworkInterface(vn), nil
}

func readNetworkInterface(ctx context.Context, p Provider, state NetworkInterface) (NetworkInterface, error) {
	vn, err := p.Client.Client.ReadNetworkInterface(ctx, &resourcespb.ReadNetworkInterfaceRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return NetworkInterface{}, err
	}
	return convertToNetworkInterface(vn), nil
}

func deleteNetworkInterface(ctx context.Context, p Provider, state NetworkInterface) error {
	_, err := p.Client.Client.DeleteNetworkInterface(ctx, &resourcespb.DeleteNetworkInterfaceRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type NetworkInterface struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	SubnetId types.String `tfsdk:"subnet_id"`
}

func convertToNetworkInterface(res *resourcespb.NetworkInterfaceResource) NetworkInterface {
	return NetworkInterface{
		Id:       types.String{Value: res.CommonParameters.ResourceId},
		Name:     types.String{Value: res.Name},
		SubnetId: types.String{Value: res.SubnetId},
	}
}

func convertFromNetworkInterface(plan NetworkInterface) *resourcespb.NetworkInterfaceArgs {
	return &resourcespb.NetworkInterfaceArgs{
		Name:     plan.Name.Value,
		SubnetId: plan.SubnetId.Value,
	}
}
