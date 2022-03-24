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

type ResourceVirtualNetworkType struct{}

func (r ResourceVirtualNetworkType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"cloud":              common.CloudsSchema,
			"location":           common.LocationSchema,
			"effective_location": common.EffectiveLocationSchema,
		},
	}, nil
}

func (r ResourceVirtualNetworkType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceVirtualNetwork{
		p: *(p.(*Provider)),
	}, nil
}

type resourceVirtualNetwork struct {
	p Provider
}

func (r resourceVirtualNetwork) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan VirtualNetwork
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Create new order from plan values
	vn, err := c.Client.CreateVirtualNetwork(ctx, &resources.CreateVirtualNetworkRequest{
		Resources: []*resources.CloudSpecificVirtualNetworkArgs{{
			CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
				Location:      c.GetLocation(plan.Location),
				CloudProvider: common.StringToCloud(plan.Cloud.Value),
			},
			Name:      plan.Name.Value,
			CidrBlock: plan.CidrBlock.Value,
		}},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating virtual_network", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "created virtual network", map[string]interface{}{"virtual_network_id": vn.CommonParameters.ResourceId})

	// Map response body to resource schema attribute
	state := r.convertResponseToVn(vn)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceVirtualNetwork) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state VirtualNetwork
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Get virtual_network from API and then update what is in state from what the API returns
	vn, err := r.p.Client.Client.ReadVirtualNetwork(ctx, &resources.ReadVirtualNetworkRequest{ResourceId: state.Id.Value})
	if err != nil {
		resp.Diagnostics.AddError("Error getting virtual_network", common.ParseGrpcErrors(err))
		return
	}

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToVn(vn)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceVirtualNetwork) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state VirtualNetwork
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

	// Update virtual_network
	vn, err := c.Client.UpdateVirtualNetwork(ctx, &resources.UpdateVirtualNetworkRequest{
		// fixme state vs plan
		ResourceId: state.Id.Value,
		Resources: []*resources.CloudSpecificVirtualNetworkArgs{{
			CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
				Location:      c.GetLocation(plan.Location),
				CloudProvider: common.StringToCloud(plan.Cloud.Value),
			},
			Name:      plan.Name.Value,
			CidrBlock: plan.CidrBlock.Value,
		}},
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating virtual_network", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "updated virtual_network", map[string]interface{}{"virtual_network_id": state.Id.Value})

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToVn(vn)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceVirtualNetwork) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state VirtualNetwork
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Delete virtual_network
	_, err := c.Client.DeleteVirtualNetwork(ctx, &resources.DeleteVirtualNetworkRequest{ResourceId: state.Id.Value})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting virtual_network",
			common.ParseGrpcErrors(err),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceVirtualNetwork) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

type VirtualNetwork struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	CidrBlock         types.String `tfsdk:"cidr_block"`
	Cloud             types.String `tfsdk:"cloud"`
	Location          types.String `tfsdk:"location"`
	EffectiveLocation types.String `tfsdk:"effective_location"`
}

func (r resourceVirtualNetwork) convertResponseToVn(res *resources.VirtualNetworkResource) VirtualNetwork {
	result := VirtualNetwork{
		Id:        types.String{Value: res.CommonParameters.ResourceId},
		Name:      types.String{Value: res.Resources[0].Name},
		CidrBlock: types.String{Value: res.Resources[0].CidrBlock},
		Cloud:     types.String{Value: strings.ToLower(res.Resources[0].CommonParameters.CloudProvider.String())},
	}
	common.SetLocation(r.p.Client.Location, res.Resources[0].CommonParameters.Location, &result.EffectiveLocation, &result.Location)
	return result
}
