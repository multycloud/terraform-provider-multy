package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"strings"
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
				Description: "Name of RouteTable",
				Required:    true,
			},
			"virtual_network_id": {
				Type:        types.StringType,
				Description: "ID of `virtual_network` resource",
				// fixme is it optional or required?
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"cloud": common.CloudsSchema,
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
						Type:        types.StringType,
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
	return resourceRouteTable{
		p: *(p.(*Provider)),
	}, nil
}

type resourceRouteTable struct {
	p Provider
}

func (r resourceRouteTable) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan RouteTable
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error communicating with server", err.Error())
		return
	}

	// Create new order from plan values
	route_table, err := c.Client.CreateRouteTable(ctx, &resourcespb.CreateRouteTableRequest{
		Resource: r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating route_table", err.Error())
		return
	}

	tflog.Trace(ctx, "created route_table", map[string]interface{}{"route_table_id": route_table.CommonParameters.ResourceId})

	// Map response body to resource schema attribute
	state := r.convertResponseToResource(route_table)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceRouteTable) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state RouteTable
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error communicating with server", err.Error())
		return
	}

	// Get route_table from API and then update what is in state from what the API returns
	rt, err := r.p.Client.Client.ReadRouteTable(ctx, &resourcespb.ReadRouteTableRequest{ResourceId: state.Id.Value})
	if err != nil {
		resp.Diagnostics.AddError("Error getting route_table", err.Error())
		return
	}

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(rt)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceRouteTable) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state RouteTable
	// Get plan values
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Get current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error communicating with server", err.Error())
		return
	}

	// Update route_table
	vn, err := c.Client.UpdateRouteTable(ctx, &resourcespb.UpdateRouteTableRequest{
		// fixme state vs plan
		ResourceId: state.Id.Value,
		Resource:   r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating route_table", err.Error())
		return
	}

	tflog.Trace(ctx, "updated route_table", map[string]interface{}{"route_table_id": state.Id.Value})

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(vn)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceRouteTable) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state RouteTable
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx, err := c.AddHeaders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error communicating with server", err.Error())
		return
	}

	// Delete route_table
	_, err = c.Client.DeleteRouteTable(ctx, &resourcespb.DeleteRouteTableRequest{ResourceId: state.Id.Value})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting route_table",
			err.Error(),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceRouteTable) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

type RouteTable struct {
	Id               types.String      `tfsdk:"id"`
	Name             types.String      `tfsdk:"name"`
	VirtualNetworkId types.String      `tfsdk:"virtual_network_id"`
	Routes           []RouteTableRoute `tfsdk:"routes"`
	Cloud            types.String      `tfsdk:"cloud"`
}

type RouteTableRoute struct {
	CidrBlock   types.String `tfsdk:"cidr_block"`
	Destination types.String `tfsdk:"destination"`
}

func (r resourceRouteTable) convertResponseToResource(res *resourcespb.RouteTableResource) RouteTable {
	var routes []RouteTableRoute
	for _, i := range res.Routes {
		routes = append(routes, RouteTableRoute{
			CidrBlock:   types.String{Value: i.CidrBlock},
			Destination: types.String{Value: strings.ToLower(i.Destination.String())},
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

func (r resourceRouteTable) convertResourcePlanToArgs(plan RouteTable) *resourcespb.RouteTableArgs {
	var routes []*resourcespb.Route
	for _, i := range plan.Routes {
		routes = append(routes, &resourcespb.Route{
			CidrBlock:   i.CidrBlock.Value,
			Destination: common.StringToRouteDestination(i.Destination.Value),
		})
	}

	return &resourcespb.RouteTableArgs{
		Name:             plan.Name.Value,
		Routes:           routes,
		VirtualNetworkId: plan.VirtualNetworkId.Value,
	}
}
