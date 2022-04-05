package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"strings"
	"terraform-provider-multy/multy/common"
)

type ResourceNetworkInterfaceType struct{}

func (r ResourceNetworkInterfaceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of Network Interface",
				Required:    true,
			},
			"subnet_id": {
				Type:        types.StringType,
				Description: "ID of `subnet` resource",
				Required:    true,
			},
			"cloud": common.CloudsSchema,
		},
	}, nil
}

func (r ResourceNetworkInterfaceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceNetworkInterface{
		p: *(p.(*Provider)),
	}, nil
}

type resourceNetworkInterface struct {
	p Provider
}

func (r resourceNetworkInterface) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan NetworkInterface
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
	vn, err := c.Client.CreateNetworkInterface(ctx, &resourcespb.CreateNetworkInterfaceRequest{
		Resource: r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating network_interface", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "created network_interface", map[string]interface{}{"network_interface_id": vn.CommonParameters.ResourceId})

	// Map response body to resource schema attribute
	state := r.convertResponseToResource(vn)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceNetworkInterface) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state NetworkInterface
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

	// Get network_interface from API and then update what is in state from what the API returns
	vn, err := r.p.Client.Client.ReadNetworkInterface(ctx, &resourcespb.ReadNetworkInterfaceRequest{ResourceId: state.Id.Value})
	if err != nil {
		resp.Diagnostics.AddError("Error getting network_interface", common.ParseGrpcErrors(err))
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

func (r resourceNetworkInterface) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state NetworkInterface
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

	// Update network_interface
	vn, err := c.Client.UpdateNetworkInterface(ctx, &resourcespb.UpdateNetworkInterfaceRequest{
		// fixme state vs plan
		ResourceId: state.Id.Value,
		Resource:   r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating network_interface", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "updated network_interface", map[string]interface{}{"network_interface_id": state.Id.Value})

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(vn)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceNetworkInterface) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state NetworkInterface
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

	// Delete network_interface
	_, err = c.Client.DeleteNetworkInterface(ctx, &resourcespb.DeleteNetworkInterfaceRequest{ResourceId: state.Id.Value})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting network_interface",
			common.ParseGrpcErrors(err),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceNetworkInterface) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

type NetworkInterface struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	SubnetId types.String `tfsdk:"subnet_id"`
	Cloud    types.String `tfsdk:"cloud"`
}

func (r resourceNetworkInterface) convertResponseToResource(res *resourcespb.NetworkInterfaceResource) NetworkInterface {
	return NetworkInterface{
		Id:       types.String{Value: res.CommonParameters.ResourceId},
		Name:     types.String{Value: res.Name},
		SubnetId: types.String{Value: res.SubnetId},
		Cloud:    types.String{Value: strings.ToLower(res.CommonParameters.CloudProvider.String())},
	}
}

func (r resourceNetworkInterface) convertResourcePlanToArgs(plan NetworkInterface) *resourcespb.NetworkInterfaceArgs {
	return &resourcespb.NetworkInterfaceArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			CloudProvider: common.StringToCloud(plan.Cloud.Value),
		},
		Name:     plan.Name.Value,
		SubnetId: plan.SubnetId.Value,
	}
}
