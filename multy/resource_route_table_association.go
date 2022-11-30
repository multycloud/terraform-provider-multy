package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
)

type ResourceRouteTableAssociationType struct{}

var routeTableAssociationAwsOutputs = map[string]attr.Type{
	"route_table_association_id_by_availability_zone": types.MapType{ElemType: types.StringType},
}

func (r ResourceRouteTableAssociationType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Route Table Association resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
			},
			"subnet_id": {
				Type:          types.StringType,
				Description:   "ID of `subnet` resource",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
			},
			"route_table_id": {
				Type:          types.StringType,
				Description:   "ID of `route_table` resource",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
			},
			"aws": {
				Description: "AWS-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: routeTableAssociationAwsOutputs},
				Computed:    true,
			},
			"resource_status": common.ResourceStatusSchema,
		},
	}, nil
}

func (r ResourceRouteTableAssociationType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
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
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromRouteTableAssociation(plan),
	})
	if err != nil {
		return RouteTableAssociation{}, err
	}
	return convertToRouteTableAssociation(vn), nil
}

func readRouteTableAssociation(ctx context.Context, p Provider, state RouteTableAssociation) (RouteTableAssociation, error) {
	vn, err := p.Client.Client.ReadRouteTableAssociation(ctx, &resourcespb.ReadRouteTableAssociationRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return RouteTableAssociation{}, err
	}
	return convertToRouteTableAssociation(vn), nil
}

func deleteRouteTableAssociation(ctx context.Context, p Provider, state RouteTableAssociation) error {
	_, err := p.Client.Client.DeleteRouteTableAssociation(ctx, &resourcespb.DeleteRouteTableAssociationRequest{
		ResourceId: state.Id.ValueString(),
	})
	return err
}

type RouteTableAssociation struct {
	Id             types.String `tfsdk:"id"`
	SubnetId       types.String `tfsdk:"subnet_id"`
	RouteTableId   types.String `tfsdk:"route_table_id"`
	AwsOutputs     types.Object `tfsdk:"aws"`
	ResourceStatus types.Map    `tfsdk:"resource_status"`
}

func convertToRouteTableAssociation(res *resourcespb.RouteTableAssociationResource) RouteTableAssociation {
	return RouteTableAssociation{
		Id:           types.String{Value: res.CommonParameters.ResourceId},
		SubnetId:     types.String{Value: res.SubnetId},
		RouteTableId: types.String{Value: res.RouteTableId},
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"route_table_association_id_by_availability_zone": common.GoMapToMapType(res.GetAwsOutputs().GetRouteTableAssociationIdByAvailabilityZone()),
			},
			AttrTypes: routeTableAssociationAwsOutputs,
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromRouteTableAssociation(plan RouteTableAssociation) *resourcespb.RouteTableAssociationArgs {
	return &resourcespb.RouteTableAssociationArgs{
		SubnetId:     plan.SubnetId.Value,
		RouteTableId: plan.RouteTableId.Value,
	}
}
