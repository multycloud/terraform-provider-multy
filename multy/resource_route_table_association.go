package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	common_proto "github.com/multycloud/multy/api/proto/common"
	"github.com/multycloud/multy/api/proto/resources"
	"strings"
	"terraform-provider-multy/multy/common"
)

type ResourceRouteTableAssociationType struct{}

func (r ResourceRouteTableAssociationType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
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
	return resourceRouteTableAssociation{
		p: *(p.(*Provider)),
	}, nil
}

type resourceRouteTableAssociation struct {
	p Provider
}

func (r resourceRouteTableAssociation) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan RouteTableAssociation
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Create new order from plan values
	vn, err := c.Client.CreateRouteTableAssociation(ctx, &resources.CreateRouteTableAssociationRequest{
		Resources: r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating route_table_association", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "created route_table_association", map[string]interface{}{"route_table_association_id": vn.CommonParameters.ResourceId})

	// Map response body to resource schema attribute
	state := r.convertResponseToResource(vn)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceRouteTableAssociation) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state RouteTableAssociation
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Get route_table_association from API and then update what is in state from what the API returns
	vn, err := r.p.Client.Client.ReadRouteTableAssociation(ctx, &resources.ReadRouteTableAssociationRequest{ResourceId: state.Id.Value})
	if err != nil {
		resp.Diagnostics.AddError("Error getting route_table_association", common.ParseGrpcErrors(err))
		return
	}

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(vn)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceRouteTableAssociation) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state RouteTableAssociation
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
	ctx = c.AddHeaders(ctx)

	// Update route_table_association
	vn, err := c.Client.UpdateRouteTableAssociation(ctx, &resources.UpdateRouteTableAssociationRequest{
		// fixme state vs plan
		ResourceId: state.Id.Value,
		Resources:  r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating route_table_association", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "updated route_table_association", map[string]interface{}{"route_table_association_id": state.Id.Value})

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(vn)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceRouteTableAssociation) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state RouteTableAssociation
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Delete route_table_association
	_, err := c.Client.DeleteRouteTableAssociation(ctx, &resources.DeleteRouteTableAssociationRequest{ResourceId: state.Id.Value})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting route_table_association",
			common.ParseGrpcErrors(err),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceRouteTableAssociation) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

type RouteTableAssociation struct {
	Id           types.String `tfsdk:"id"`
	SubnetId     types.String `tfsdk:"subnet_id"`
	RouteTableId types.String `tfsdk:"route_table_id"`
	Cloud        types.String `tfsdk:"cloud"`
}

func (r resourceRouteTableAssociation) convertResponseToResource(res *resources.RouteTableAssociationResource) RouteTableAssociation {
	return RouteTableAssociation{
		Id:           types.String{Value: res.CommonParameters.ResourceId},
		SubnetId:     types.String{Value: res.Resources[0].SubnetId},
		RouteTableId: types.String{Value: res.Resources[0].RouteTableId},
		Cloud:        types.String{Value: strings.ToLower(res.Resources[0].CommonParameters.CloudProvider.String())},
	}
}

func (r resourceRouteTableAssociation) convertResourcePlanToArgs(plan RouteTableAssociation) []*resources.CloudSpecificRouteTableAssociationArgs {
	return []*resources.CloudSpecificRouteTableAssociationArgs{{
		CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
			CloudProvider: common.StringToCloud(plan.Cloud.Value),
		},
		SubnetId:     plan.SubnetId.Value,
		RouteTableId: plan.RouteTableId.Value,
	}}
}
