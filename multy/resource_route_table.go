package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceRouteTableType struct{}

func (r ResourceRouteTableType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Route Table resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of Route Table",
				Required:    true,
				// todo: if not aws
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"virtual_network_id": {
				Type:          types.StringType,
				Description:   "ID of `virtual_network` resource",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
		},
		Blocks: map[string]tfsdk.Block{
			"route": {
				Description: "Route block definition",
				Attributes: map[string]tfsdk.Attribute{
					"cidr_block": {
						Type:        types.StringType,
						Description: "CIDR block of network rule",
						Required:    true,
						Validators:  []tfsdk.AttributeValidator{validators.IsCidrValidator{}},
					},
					"destination": {
						Type:        mtypes.RouteDestinationType,
						Description: fmt.Sprintf("Destination of route. Accepted values are %s", common.StringSliceToDocsMarkdown(mtypes.RouteDestinationType.GetAllValues())),
						Required:    true,
						Validators:  []tfsdk.AttributeValidator{validators.NewValidator(mtypes.RouteDestinationType)},
					},
				},
				NestingMode: tfsdk.BlockNestingModeSet,
			},
		},
	}, nil
}

func (r ResourceRouteTableType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[RouteTable]{
		p:          *(p.(*Provider)),
		createFunc: createRouteTable,
		updateFunc: updateRouteTable,
		readFunc:   readRouteTable,
		deleteFunc: deleteRouteTable,
	}, nil
}

func createRouteTable(ctx context.Context, p Provider, plan RouteTable) (RouteTable, error) {
	vn, err := p.Client.Client.CreateRouteTable(ctx, &resourcespb.CreateRouteTableRequest{
		Resource: convertFromRouteTable(plan),
	})
	if err != nil {
		return RouteTable{}, err
	}
	return convertToRouteTable(vn), nil
}

func updateRouteTable(ctx context.Context, p Provider, plan RouteTable) (RouteTable, error) {
	vn, err := p.Client.Client.UpdateRouteTable(ctx, &resourcespb.UpdateRouteTableRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromRouteTable(plan),
	})
	if err != nil {
		return RouteTable{}, err
	}
	return convertToRouteTable(vn), nil
}

func readRouteTable(ctx context.Context, p Provider, state RouteTable) (RouteTable, error) {
	vn, err := p.Client.Client.ReadRouteTable(ctx, &resourcespb.ReadRouteTableRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return RouteTable{}, err
	}
	return convertToRouteTable(vn), nil
}

func deleteRouteTable(ctx context.Context, p Provider, state RouteTable) error {
	_, err := p.Client.Client.DeleteRouteTable(ctx, &resourcespb.DeleteRouteTableRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type RouteTable struct {
	Id               types.String      `tfsdk:"id"`
	Name             types.String      `tfsdk:"name"`
	VirtualNetworkId types.String      `tfsdk:"virtual_network_id"`
	Routes           []RouteTableRoute `tfsdk:"route"`
}

type RouteTableRoute struct {
	CidrBlock   types.String                                   `tfsdk:"cidr_block"`
	Destination mtypes.EnumValue[resourcespb.RouteDestination] `tfsdk:"destination"`
}

func convertToRouteTable(res *resourcespb.RouteTableResource) RouteTable {
	var routes []RouteTableRoute
	for _, i := range res.Routes {
		routes = append(routes, RouteTableRoute{
			CidrBlock:   types.String{Value: i.CidrBlock},
			Destination: mtypes.RouteDestinationType.NewVal(i.Destination),
		})
	}

	result := RouteTable{
		Id:               types.String{Value: res.CommonParameters.ResourceId},
		Name:             types.String{Value: res.Name},
		Routes:           routes,
		VirtualNetworkId: types.String{Value: res.VirtualNetworkId},
	}

	return result
}

func convertFromRouteTable(plan RouteTable) *resourcespb.RouteTableArgs {
	var routes []*resourcespb.Route
	for _, i := range plan.Routes {
		routes = append(routes, &resourcespb.Route{
			CidrBlock:   i.CidrBlock.Value,
			Destination: i.Destination.Value,
		})
	}

	return &resourcespb.RouteTableArgs{
		Name:             plan.Name.Value,
		Routes:           routes,
		VirtualNetworkId: plan.VirtualNetworkId.Value,
	}
}
