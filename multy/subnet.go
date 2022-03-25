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

type ResourceSubnetType struct{}

func (r ResourceSubnetType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"cidr_block": {
				Type:     types.StringType,
				Required: true,
				//ValidateFunc: validation.IsCIDR,
			},
			"virtual_network_id": {
				Type:     types.StringType,
				Required: true,
			},
			"availability_zone": {
				Type:     types.Int64Type,
				Optional: true,
				Computed: true,
			},
			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,
		},
	}, nil
}

func (r ResourceSubnetType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceSubnet{
		p: *(p.(*Provider)),
	}, nil
}

type resourceSubnet struct {
	p Provider
}

func (r resourceSubnet) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan Subnet
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Create new order from plan values
	subnet, err := c.Client.CreateSubnet(ctx, &resources.CreateSubnetRequest{
		Resources: convertSubnetPlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating subnet", err.Error())
		return
	}

	tflog.Trace(ctx, "created subnet", map[string]interface{}{"subnet_id": subnet.CommonParameters.ResourceId})

	// Map response body to resource schema attribute
	state := r.convertResponseToSubnet(subnet)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceSubnet) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state Subnet
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Get subnet from API and then update what is in state from what the API returns
	subnet, err := r.p.Client.Client.ReadSubnet(ctx, &resources.ReadSubnetRequest{ResourceId: state.Id.Value})
	if err != nil {
		resp.Diagnostics.AddError("Error getting subnet", err.Error())
		return
	}

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToSubnet(subnet)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceSubnet) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state Subnet
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

	// Update subnet
	vn, err := c.Client.UpdateSubnet(ctx, &resources.UpdateSubnetRequest{
		// fixme state vs plan
		ResourceId: state.Id.Value,
		Resources:  convertSubnetPlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating subnet", err.Error())
		return
	}

	tflog.Trace(ctx, "updated subnet", map[string]interface{}{"subnet_id": state.Id.Value})

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToSubnet(vn)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceSubnet) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state Subnet
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Delete subnet
	_, err := c.Client.DeleteSubnet(ctx, &resources.DeleteSubnetRequest{ResourceId: state.Id.Value})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting subnet",
			err.Error(),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceSubnet) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

type Subnet struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	CidrBlock        types.String `tfsdk:"cidr_block"`
	VirtualNetworkId types.String `tfsdk:"virtual_network_id"`
	AvailabilityZone types.Int64  `tfsdk:"availability_zone"`
	Cloud            types.String `tfsdk:"cloud"`
	Location         types.String `tfsdk:"location"`
}

//type CommonResourceParams struct {
//	Cloud    types.String `tfsdk:"cloud"`
//	Location types.String `tfsdk:"location"`
//}

func (r resourceSubnet) convertResponseToSubnet(res *resources.SubnetResource) Subnet {
	result := Subnet{
		Id:               types.String{Value: res.CommonParameters.ResourceId},
		Name:             types.String{Value: res.Resources[0].Name},
		CidrBlock:        types.String{Value: res.Resources[0].CidrBlock},
		AvailabilityZone: types.Int64{Value: int64(res.Resources[0].AvailabilityZone)},
		VirtualNetworkId: types.String{Value: res.Resources[0].VirtualNetworkId},
		Cloud:            types.String{Value: strings.ToLower(res.Resources[0].CommonParameters.CloudProvider.String())},
		Location:         types.String{Value: strings.ToLower(res.Resources[0].CommonParameters.Location.String())},
	}

	return result
}

func convertSubnetPlanToArgs(plan Subnet) []*resources.CloudSpecificSubnetArgs {
	return []*resources.CloudSpecificSubnetArgs{{
		CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
			Location:      common.StringToLocation(plan.Location.Value),
			CloudProvider: common.StringToCloud(plan.Cloud.Value),
		},
		Name:             plan.Name.Value,
		CidrBlock:        plan.CidrBlock.Value,
		VirtualNetworkId: plan.VirtualNetworkId.Value,
		AvailabilityZone: int32(plan.AvailabilityZone.Value),
	}}
}
