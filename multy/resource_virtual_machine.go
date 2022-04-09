package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceVirtualMachineType struct{}

func (r ResourceVirtualMachineType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Virtual Machine resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of Virtual Machine",
				Required:    true,
			},
			"operating_system": {
				Type:        mtypes.OperatingSystemType,
				Description: fmt.Sprintf("Operating System of Virtual Machine. Accepted values are %s", common.StringSliceToDocsMarkdown(mtypes.OperatingSystemType.GetAllValues())),
				Required:    true,
				Validators:  []tfsdk.AttributeValidator{validators.NewValidator(mtypes.OperatingSystemType)},
			},
			"size": {
				Type:        mtypes.VmSizeType,
				Description: fmt.Sprintf("Size of Virtual Machine. Accepted values are %s", common.StringSliceToDocsMarkdown(mtypes.VmSizeType.GetAllValues())),
				Required:    true,
				Validators:  []tfsdk.AttributeValidator{validators.NewValidator(mtypes.VmSizeType)},
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
				Description: "Base64 User Data script of Virtual Machine that will run on instance launch",
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
	return MultyResource[VirtualMachine]{
		p:          *(p.(*Provider)),
		createFunc: createVirtualMachine,
		updateFunc: updateVirtualMachine,
		readFunc:   readVirtualMachine,
		deleteFunc: deleteVirtualMachine,
	}, nil
}

func createVirtualMachine(ctx context.Context, p Provider, plan VirtualMachine) (VirtualMachine, error) {
	pIpId := !plan.PublicIpId.Null && !plan.PublicIpId.Unknown && plan.PublicIpId.Value != ""
	pIp := !plan.PublicIp.Null && !plan.PublicIp.Unknown && plan.PublicIp.Value
	// fixme check isnt working
	if pIpId && pIp {
		return VirtualMachine{}, fmt.Errorf("cannot set both public_ip and public_ip_id")
	}

	// Create new order from plan values
	vm, err := p.Client.Client.CreateVirtualMachine(ctx, &resourcespb.CreateVirtualMachineRequest{
		Resource: convertFromVirtualMachine(plan),
	})
	if err != nil {
		return VirtualMachine{}, err
	}

	tflog.Trace(ctx, "created virtual network", map[string]interface{}{"virtual_machine_id": vm.CommonParameters.ResourceId})

	return convertToVirtualMachine(vm), nil
}

func updateVirtualMachine(ctx context.Context, p Provider, plan VirtualMachine) (VirtualMachine, error) {
	pIpId := !plan.PublicIpId.Null && !plan.PublicIpId.Unknown && plan.PublicIpId.Value != ""
	pIp := !plan.PublicIp.Null && !plan.PublicIp.Unknown && plan.PublicIp.Value
	// fixme check isnt working
	if pIpId && pIp {
		return VirtualMachine{}, fmt.Errorf("cannot set both public_ip and public_ip_id")
	}

	// Create new order from plan values
	vm, err := p.Client.Client.CreateVirtualMachine(ctx, &resourcespb.CreateVirtualMachineRequest{
		Resource: convertFromVirtualMachine(plan),
	})
	if err != nil {
		return VirtualMachine{}, err
	}

	tflog.Trace(ctx, "updated virtual network", map[string]interface{}{"virtual_machine_id": vm.CommonParameters.ResourceId})

	return convertToVirtualMachine(vm), nil
}

func readVirtualMachine(ctx context.Context, p Provider, state VirtualMachine) (VirtualMachine, error) {
	vm, err := p.Client.Client.ReadVirtualMachine(ctx, &resourcespb.ReadVirtualMachineRequest{ResourceId: state.Id.Value})
	if err != nil {
		return VirtualMachine{}, err
	}

	state = convertToVirtualMachine(vm)
	return convertToVirtualMachine(vm), nil
}

func deleteVirtualMachine(ctx context.Context, p Provider, state VirtualMachine) error {
	_, err := p.Client.Client.DeleteVirtualMachine(ctx, &resourcespb.DeleteVirtualMachineRequest{ResourceId: state.Id.Value})
	return err
}

func convertToVirtualMachine(res *resourcespb.VirtualMachineResource) VirtualMachine {
	return VirtualMachine{
		Id:                      types.String{Value: res.CommonParameters.ResourceId},
		Name:                    types.String{Value: res.Name},
		OperatingSystem:         mtypes.OperatingSystemType.NewVal(res.OperatingSystem),
		Size:                    mtypes.VmSizeType.NewVal(res.VmSize),
		SubnetId:                types.String{Value: res.SubnetId},
		NetworkInterfaceIds:     common.DefaultSliceToNull(common.TypesStringToStringSlice(res.NetworkInterfaceIds)),
		NetworkSecurityGroupIds: common.DefaultSliceToNull(common.TypesStringToStringSlice(res.NetworkSecurityGroupIds)),
		UserData:                types.String{Value: res.UserData},
		PublicSshKey:            types.String{Value: res.PublicSshKey},
		PublicIpId:              common.DefaultToNull[types.String](res.PublicIpId),
		PublicIp:                types.Bool{Value: res.GeneratePublicIp},
		Cloud:                   mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:                mtypes.LocationType.NewVal(res.CommonParameters.Location),
	}
}

func convertFromVirtualMachine(plan VirtualMachine) *resourcespb.VirtualMachineArgs {
	return &resourcespb.VirtualMachineArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			Location:      plan.Location.Value,
			CloudProvider: plan.Cloud.Value,
		},
		Name:                    plan.Name.Value,
		OperatingSystem:         plan.OperatingSystem.Value,
		NetworkInterfaceIds:     common.StringSliceToTypesString(plan.NetworkInterfaceIds),
		NetworkSecurityGroupIds: common.StringSliceToTypesString(plan.NetworkSecurityGroupIds),
		VmSize:                  plan.Size.Value,
		UserData:                plan.UserData.Value,
		SubnetId:                plan.SubnetId.Value,
		PublicSshKey:            plan.PublicSshKey.Value,
		PublicIpId:              plan.PublicIpId.Value,
		GeneratePublicIp:        plan.PublicIp.Value,
	}
}

type VirtualMachine struct {
	Id                      types.String                                    `tfsdk:"id"`
	Name                    types.String                                    `tfsdk:"name"`
	OperatingSystem         mtypes.EnumValue[commonpb.OperatingSystem_Enum] `tfsdk:"operating_system"`
	Size                    mtypes.EnumValue[commonpb.VmSize_Enum]          `tfsdk:"size"`
	SubnetId                types.String                                    `tfsdk:"subnet_id"`
	NetworkInterfaceIds     []types.String                                  `tfsdk:"network_interface_ids"`
	NetworkSecurityGroupIds []types.String                                  `tfsdk:"network_security_group_ids"`
	UserData                types.String                                    `tfsdk:"user_data"`
	PublicSshKey            types.String                                    `tfsdk:"public_ssh_key"`
	PublicIpId              types.String                                    `tfsdk:"public_ip_id"`
	PublicIp                types.Bool                                      `tfsdk:"public_ip"`
	Cloud                   mtypes.EnumValue[commonpb.CloudProvider]        `tfsdk:"cloud"`
	Location                mtypes.EnumValue[commonpb.Location]             `tfsdk:"location"`
}
