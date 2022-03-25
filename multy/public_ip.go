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

type ResourcePublicIpType struct{}

func (r ResourcePublicIpType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of Public IP",
				Required:    true,
			},
			"network_interface_id": {
				Type:        types.StringType,
				Description: "Associate Public IP to `network_interface` resource",
				Optional:    true,
			},
			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,
		},
	}, nil
}

func (r ResourcePublicIpType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourcePublicIp{
		p: *(p.(*Provider)),
	}, nil
}

type resourcePublicIp struct {
	p Provider
}

func (r resourcePublicIp) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan PublicIp
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Create new order from plan values
	vn, err := c.Client.CreatePublicIp(ctx, &resources.CreatePublicIpRequest{
		Resources: r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating public_ip", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "created public_ip", map[string]interface{}{"public_ip_id": vn.CommonParameters.ResourceId})

	// Map response body to resource schema attribute
	state := r.convertResponseToResource(vn)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourcePublicIp) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state PublicIp
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Get public_ip from API and then update what is in state from what the API returns
	vn, err := r.p.Client.Client.ReadPublicIp(ctx, &resources.ReadPublicIpRequest{ResourceId: state.Id.Value})
	if err != nil {
		resp.Diagnostics.AddError("Error getting public_ip", common.ParseGrpcErrors(err))
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

func (r resourcePublicIp) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state PublicIp
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

	// Update public_ip
	vn, err := c.Client.UpdatePublicIp(ctx, &resources.UpdatePublicIpRequest{
		// fixme state vs plan
		ResourceId: state.Id.Value,
		Resources:  r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating public_ip", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "updated public_ip", map[string]interface{}{"public_ip_id": state.Id.Value})

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(vn)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourcePublicIp) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state PublicIp
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Delete public_ip
	_, err := c.Client.DeletePublicIp(ctx, &resources.DeletePublicIpRequest{ResourceId: state.Id.Value})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting public_ip",
			common.ParseGrpcErrors(err),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourcePublicIp) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

type PublicIp struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	NetworkInterfaceId types.String `tfsdk:"network_interface_id"`
	Cloud              types.String `tfsdk:"cloud"`
	Location           types.String `tfsdk:"location"`
}

func (r resourcePublicIp) convertResponseToResource(res *resources.PublicIpResource) PublicIp {
	return PublicIp{
		Id:                 types.String{Value: res.CommonParameters.ResourceId},
		Name:               types.String{Value: res.Resources[0].Name},
		NetworkInterfaceId: types.String{Value: res.Resources[0].NetworkInterfaceId},
		Cloud:              types.String{Value: strings.ToLower(res.Resources[0].CommonParameters.CloudProvider.String())},
		Location:           types.String{Value: strings.ToLower(res.Resources[0].CommonParameters.Location.String())},
	}
}

func (r resourcePublicIp) convertResourcePlanToArgs(plan PublicIp) []*resources.CloudSpecificPublicIpArgs {
	return []*resources.CloudSpecificPublicIpArgs{{
		CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
			Location:      common.StringToLocation(plan.Location.Value),
			CloudProvider: common.StringToCloud(plan.Cloud.Value),
		},
		Name:               plan.Name.Value,
		NetworkInterfaceId: plan.NetworkInterfaceId.Value,
	}}
}
