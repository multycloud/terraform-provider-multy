package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceRouteTableType struct{}

var routeTableAwsOutputs = map[string]attr.Type{
	"route_table_id": types.StringType,
}

var routeTableAzureOutputs = map[string]attr.Type{
	"route_table_id": types.StringType,
}

var routeTableGcpOutputs = map[string]attr.Type{
	"compute_route_ids": types.ListType{ElemType: types.StringType},
}

var routeTableSchema = tfsdk.Schema{
	MarkdownDescription: "Provides Multy Route Table resource",
	Attributes: map[string]tfsdk.Attribute{
		"id": {
			Type:          types.StringType,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		},
		"name": {
			Type:        types.StringType,
			Description: "Name of Route Table",
			Required:    true,
			// todo: if not aws
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"virtual_network_id": {
			Type:          types.StringType,
			Description:   "ID of `virtual_network` resource",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"aws": {
			Description: "AWS-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: routeTableAwsOutputs},
			Computed:    true,
		},
		"azure": {
			Description: "Azure-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: routeTableAzureOutputs},
			Computed:    true,
		},
		"gcp": {
			Description: "GCP-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: routeTableGcpOutputs},
			Computed:    true,
		},
		"resource_status": common.ResourceStatusSchema,
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
}

func (r ResourceRouteTableType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[RouteTable]{
		p:          *(p.(*Provider)),
		createFunc: createRouteTable,
		updateFunc: updateRouteTable,
		readFunc:   readRouteTable,
		deleteFunc: deleteRouteTable,
		name:       "multy_route_table",
		schema:     routeTableSchema,
	}
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
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromRouteTable(plan),
	})
	if err != nil {
		return RouteTable{}, err
	}
	return convertToRouteTable(vn), nil
}

func readRouteTable(ctx context.Context, p Provider, state RouteTable) (RouteTable, error) {
	vn, err := p.Client.Client.ReadRouteTable(ctx, &resourcespb.ReadRouteTableRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return RouteTable{}, err
	}
	return convertToRouteTable(vn), nil
}

func deleteRouteTable(ctx context.Context, p Provider, state RouteTable) error {
	_, err := p.Client.Client.DeleteRouteTable(ctx, &resourcespb.DeleteRouteTableRequest{
		ResourceId: state.Id.ValueString(),
	})
	return err
}

type RouteTable struct {
	Id               types.String      `tfsdk:"id"`
	Name             types.String      `tfsdk:"name"`
	VirtualNetworkId types.String      `tfsdk:"virtual_network_id"`
	Routes           []RouteTableRoute `tfsdk:"route"`
	AwsOutputs       types.Object      `tfsdk:"aws"`
	AzureOutputs     types.Object      `tfsdk:"azure"`
	GcpOutputs       types.Object      `tfsdk:"gcp"`
	ResourceStatus   types.Map         `tfsdk:"resource_status"`
}

type RouteTableRoute struct {
	CidrBlock   types.String                                   `tfsdk:"cidr_block"`
	Destination mtypes.EnumValue[resourcespb.RouteDestination] `tfsdk:"destination"`
}

func convertToRouteTable(res *resourcespb.RouteTableResource) RouteTable {
	var routes []RouteTableRoute
	for _, i := range res.Routes {
		routes = append(routes, RouteTableRoute{
			CidrBlock:   types.StringValue(i.CidrBlock),
			Destination: mtypes.RouteDestinationType.NewVal(i.Destination),
		})
	}

	result := RouteTable{
		Id:               types.StringValue(res.CommonParameters.ResourceId),
		Name:             types.StringValue(res.Name),
		Routes:           routes,
		VirtualNetworkId: types.StringValue(res.VirtualNetworkId),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, routeTableAwsOutputs, map[string]attr.Value{
			"route_table_id": common.DefaultToNull[types.String](res.GetAwsOutputs().GetRouteTableId()),
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, routeTableAzureOutputs, map[string]attr.Value{
			"route_table_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetRouteTableId()),
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, routeTableGcpOutputs, map[string]attr.Value{
			"compute_route_ids": common.TypesStringListToListType(res.GetGcpOutputs().GetComputeRouteId()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}

	return result
}

func convertFromRouteTable(plan RouteTable) *resourcespb.RouteTableArgs {
	var routes []*resourcespb.Route
	for _, i := range plan.Routes {
		routes = append(routes, &resourcespb.Route{
			CidrBlock:   i.CidrBlock.ValueString(),
			Destination: i.Destination.Value,
		})
	}

	return &resourcespb.RouteTableArgs{
		Name:             plan.Name.ValueString(),
		Routes:           routes,
		VirtualNetworkId: plan.VirtualNetworkId.ValueString(),
	}
}
