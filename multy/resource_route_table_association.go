package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
)

type ResourceRouteTableAssociationType struct{}

func (r ResourceRouteTableAssociationType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Route Table Association resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"subnet_id": {
				Type:        types.StringType,
				Description: "ID of `subnet` resource",
				Required:    true,
			},
			"route_table_id": {
				Type:        types.StringType,
				Description: "ID of `route_table` resource",
				Required:    true,
			},
		},
	}, nil
}

func (r ResourceRouteTableAssociationType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[RouteTableAssociation]{
		p:          *(p.(*Provider)),
		createFunc: createRouteTableAssociation,
		updateFunc: updateRouteTableAssociation,
		readFunc:   readRouteTableAssociation,
		deleteFunc: deleteRouteTableAssociation,
	}, nil
}

func createRouteTableAssociation(ctx context.Context, p Provider, plan RouteTableAssociation) (RouteTableAssociation, error) {
	vn, err := p.Client.Client.CreateRouteTableAssociation(ctx, &resourcespb.CreateRouteTableAssociationRequest{
		Resource: convertFromRouteTableAssociation(plan),
	})
	if err != nil {
		return RouteTableAssociation{}, err
	}
	return convertToRouteTableAssociation(vn), nil
}

func updateRouteTableAssociation(ctx context.Context, p Provider, plan RouteTableAssociation) (RouteTableAssociation, error) {
	vn, err := p.Client.Client.UpdateRouteTableAssociation(ctx, &resourcespb.UpdateRouteTableAssociationRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromRouteTableAssociation(plan),
	})
	if err != nil {
		return RouteTableAssociation{}, err
	}
	return convertToRouteTableAssociation(vn), nil
}

func readRouteTableAssociation(ctx context.Context, p Provider, state RouteTableAssociation) (RouteTableAssociation, error) {
	vn, err := p.Client.Client.ReadRouteTableAssociation(ctx, &resourcespb.ReadRouteTableAssociationRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return RouteTableAssociation{}, err
	}
	return convertToRouteTableAssociation(vn), nil
}

func deleteRouteTableAssociation(ctx context.Context, p Provider, state RouteTableAssociation) error {
	_, err := p.Client.Client.DeleteRouteTableAssociation(ctx, &resourcespb.DeleteRouteTableAssociationRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type RouteTableAssociation struct {
	Id           types.String `tfsdk:"id"`
	SubnetId     types.String `tfsdk:"subnet_id"`
	RouteTableId types.String `tfsdk:"route_table_id"`
}

func convertToRouteTableAssociation(res *resourcespb.RouteTableAssociationResource) RouteTableAssociation {
	return RouteTableAssociation{
		Id:           types.String{Value: res.CommonParameters.ResourceId},
		SubnetId:     types.String{Value: res.SubnetId},
		RouteTableId: types.String{Value: res.RouteTableId},
	}
}

func convertFromRouteTableAssociation(plan RouteTableAssociation) *resourcespb.RouteTableAssociationArgs {
	return &resourcespb.RouteTableAssociationArgs{
		SubnetId:     plan.SubnetId.Value,
		RouteTableId: plan.RouteTableId.Value,
	}
}
