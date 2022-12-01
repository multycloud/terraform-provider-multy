package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

var rtaSchema = tfsdk.Schema{
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
}

func (r ResourceRouteTableAssociationType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[RouteTableAssociation]{
		p:          *(p.(*Provider)),
		createFunc: createRouteTableAssociation,
		updateFunc: updateRouteTableAssociation,
		readFunc:   readRouteTableAssociation,
		deleteFunc: deleteRouteTableAssociation,
		name:       "multy_route_table_association",
		schema:     rtaSchema,
	}
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
		Id:           types.StringValue(res.CommonParameters.ResourceId),
		SubnetId:     types.StringValue(res.SubnetId),
		RouteTableId: types.StringValue(res.RouteTableId),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, routeTableAssociationAwsOutputs, map[string]attr.Value{
			"route_table_association_id_by_availability_zone": common.GoMapToMapType(res.GetAwsOutputs().GetRouteTableAssociationIdByAvailabilityZone()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromRouteTableAssociation(plan RouteTableAssociation) *resourcespb.RouteTableAssociationArgs {
	return &resourcespb.RouteTableAssociationArgs{
		SubnetId:     plan.SubnetId.ValueString(),
		RouteTableId: plan.RouteTableId.ValueString(),
	}
}
