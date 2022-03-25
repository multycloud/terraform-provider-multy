package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	common_proto "github.com/multycloud/multy/api/proto/common"
	"github.com/multycloud/multy/api/proto/resources"
	"strings"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/validators"
)

type ResourceVirtualMachineType struct{}

func (r ResourceVirtualMachineType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of Virtual Machine",
				Required:    true,
			},
			"operating_system": {
				Type:        types.StringType,
				Description: fmt.Sprintf("Operating System of Virtual Machine. Accepted values are %s", common.GetVmOperatingSystem()),
				Required:    true,
				Validators:  []tfsdk.AttributeValidator{validators.StringInSliceValidator{Enum: common.GetVmOperatingSystem()}},
			},
			"size": {
				Type:        types.StringType,
				Description: fmt.Sprintf("Size of Virtual Machine. Accepted values are %s", common.GetVmSize()),
				Required:    true,
				Validators:  []tfsdk.AttributeValidator{validators.StringInSliceValidator{Enum: common.GetVmSize()}},
			},
			"subnet_id": {
				Type:        types.StringType,
				Description: "ID of `subnet` resource",
				Required:    true,
			},
			"network_interface_ids": {
				Type:        types.ListType{ElemType: types.StringType},
				Description: "IDs of `network_interface` resource",
				Optional:    true,
			},
			"network_security_group_ids": {
				Type:        types.ListType{ElemType: types.StringType},
				Description: "IDs of `network_security_group` resource",
				Optional:    true,
			},
			"user_data": {
				Type: types.StringType,
				// fixme check instance launch or boot
				Description: "User Data script of Virtual Machine that will run on instance launch",
				Optional:    true,
			},
			"public_ssh_key": {
				Type:        types.StringType,
				Description: "Public SSH Key of Virtual Machine",
				Optional:    true,
			},
			"public_ip_id": {
				Type:        types.StringType,
				Description: "ID of `public_ip` resource. Cannot be used with `public_ip`",
				Optional:    true,
				// TODO: validate if not empty string
			},
			"public_ip": {
				Type:        types.BoolType,
				Description: "If true, a public IP will be automatically generated. Cannot be used with `public_ip_id`",
				Optional:    true,
				// defaults to false
				Computed: true,
			},
			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,
		},
	}, nil
}

func (r ResourceVirtualMachineType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceVirtualMachine{
		p: *(p.(*Provider)),
	}, nil
}

type resourceVirtualMachine struct {
	p Provider
}

func (r resourceVirtualMachine) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan VirtualMachine
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	pIpId := !plan.PublicIpId.Null && !plan.PublicIpId.Unknown && plan.PublicIpId.Value != ""
	pIp := !plan.PublicIp.Null && !plan.PublicIp.Unknown && plan.PublicIp.Value
	// fixme check isnt working
	if pIpId && pIp {
		resp.Diagnostics.AddError("Error creating virtual_machine", "cannot set both public_ip and public_ip_id")
		return
	}

	// Create new order from plan values
	vm, err := c.Client.CreateVirtualMachine(ctx, &resources.CreateVirtualMachineRequest{
		Resources: r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating virtual_machine", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "created virtual network", map[string]interface{}{"virtual_machine_id": vm.CommonParameters.ResourceId})

	// Map response body to resource schema attribute
	state := r.convertResponseToResource(vm)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceVirtualMachine) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state VirtualMachine
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Get virtual_machine from API and then update what is in state from what the API returns
	vm, err := r.p.Client.Client.ReadVirtualMachine(ctx, &resources.ReadVirtualMachineRequest{ResourceId: state.Id.Value})
	if err != nil {
		resp.Diagnostics.AddError("Error getting virtual_machine", common.ParseGrpcErrors(err))
		return
	}

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(vm)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceVirtualMachine) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state VirtualMachine
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

	// Update virtual_machine
	vm, err := c.Client.UpdateVirtualMachine(ctx, &resources.UpdateVirtualMachineRequest{
		// fixme state vs plan
		ResourceId: state.Id.Value,
		Resources:  r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating virtual_machine", common.ParseGrpcErrors(err))
		return
	}

	tflog.Trace(ctx, "updated virtual_machine", map[string]interface{}{"virtual_machine_id": state.Id.Value})

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(vm)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceVirtualMachine) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state VirtualMachine
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Delete virtual_machine
	_, err := c.Client.DeleteVirtualMachine(ctx, &resources.DeleteVirtualMachineRequest{ResourceId: state.Id.Value})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting virtual_machine",
			common.ParseGrpcErrors(err),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceVirtualMachine) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func (r resourceVirtualMachine) convertResponseToResource(res *resources.VirtualMachineResource) VirtualMachine {
	return VirtualMachine{
		Id:                      types.String{Value: res.CommonParameters.ResourceId},
		Name:                    types.String{Value: res.Resources[0].Name},
		OperatingSystem:         types.String{Value: strings.ToLower(res.Resources[0].OperatingSystem.String())},
		Size:                    common.DefaultEnumToNull(res.Resources[0].VmSize),
		SubnetId:                types.String{Value: res.Resources[0].SubnetId},
		NetworkInterfaceIds:     common.DefaultSliceToNull(common.TypesStringToStringSlice(res.Resources[0].NetworkInterfaceIds)),
		NetworkSecurityGroupIds: common.DefaultSliceToNull(common.TypesStringToStringSlice(res.Resources[0].NetworkSecurityGroupIds)),
		UserData:                types.String{Value: res.Resources[0].UserData},
		PublicSshKey:            types.String{Value: res.Resources[0].PublicSshKey},
		PublicIpId:              common.DefaultToNull[types.String](res.Resources[0].PublicIpId),
		PublicIp:                types.Bool{Value: res.Resources[0].GeneratePublicIp},
		Cloud:                   types.String{Value: strings.ToLower(res.Resources[0].CommonParameters.CloudProvider.String())},
		Location:                types.String{Value: strings.ToLower(res.Resources[0].CommonParameters.Location.String())},
	}
}

func (r resourceVirtualMachine) convertResourcePlanToArgs(plan VirtualMachine) []*resources.CloudSpecificVirtualMachineArgs {
	return []*resources.CloudSpecificVirtualMachineArgs{{
		CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
			Location:      common.StringToLocation(plan.Location.Value),
			CloudProvider: common.StringToCloud(plan.Cloud.Value),
		},
		Name:                    plan.Name.Value,
		OperatingSystem:         common.StringToVmOperatingSystem(plan.OperatingSystem.Value),
		NetworkInterfaceIds:     common.StringSliceToTypesString(plan.NetworkInterfaceIds),
		NetworkSecurityGroupIds: common.StringSliceToTypesString(plan.NetworkSecurityGroupIds),
		VmSize:                  common.StringToVmSize(plan.Size.Value),
		UserData:                plan.UserData.Value,
		SubnetId:                plan.SubnetId.Value,
		PublicSshKey:            plan.PublicSshKey.Value,
		PublicIpId:              plan.PublicIpId.Value,
		GeneratePublicIp:        plan.PublicIp.Value,
	}}
}

type VirtualMachine struct {
	Id                      types.String   `tfsdk:"id"`
	Name                    types.String   `tfsdk:"name"`
	OperatingSystem         types.String   `tfsdk:"operating_system"`
	Size                    types.String   `tfsdk:"size"`
	SubnetId                types.String   `tfsdk:"subnet_id"`
	NetworkInterfaceIds     []types.String `tfsdk:"network_interface_ids"`
	NetworkSecurityGroupIds []types.String `tfsdk:"network_security_group_ids"`
	UserData                types.String   `tfsdk:"user_data"`
	PublicSshKey            types.String   `tfsdk:"ssh_key"`
	PublicIpId              types.String   `tfsdk:"public_ip_id"`
	PublicIp                types.Bool     `tfsdk:"public_ip"`
	Cloud                   types.String   `tfsdk:"cloud"`
	Location                types.String   `tfsdk:"location"`
}
