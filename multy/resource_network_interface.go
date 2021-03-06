package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
)

type ResourceNetworkInterfaceType struct{}

var networkInterfaceAwsOutputs = map[string]attr.Type{
	"network_interface_id": types.StringType,
	"eip_association_id":   types.StringType,
}

var networkInterfaceAzureOutputs = map[string]attr.Type{
	"network_interface_id": types.StringType,
}

func (r ResourceNetworkInterfaceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Network Interface resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"resource_group_id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"name": {
				Type:          types.StringType,
				Description:   "Name of Network Interface",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
			},
			"subnet_id": {
				Type:          types.StringType,
				Description:   "ID of `subnet` resource",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"public_ip_id": {
				Type:          types.StringType,
				Description:   "ID of `public_ip` resource",
				Optional:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
			},
			"availability_zone": {
				Type:          types.Int64Type,
				Description:   "Availability zone where this machine should be placed",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
				Validators:    []tfsdk.AttributeValidator{mtypes.NonEmptyIntValidator},
			},
			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,
			"aws": {
				Description: "AWS-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: networkInterfaceAwsOutputs},
				Computed:    true,
			},
			"azure": {
				Description: "Azure-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: networkInterfaceAzureOutputs},
				Computed:    true,
			},
		},
	}, nil
}

func (r ResourceNetworkInterfaceType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[NetworkInterface]{
		p:          *(p.(*Provider)),
		createFunc: createNetworkInterface,
		updateFunc: updateNetworkInterface,
		readFunc:   readNetworkInterface,
		deleteFunc: deleteNetworkInterface,
	}, nil
}

func createNetworkInterface(ctx context.Context, p Provider, plan NetworkInterface) (NetworkInterface, error) {
	vn, err := p.Client.Client.CreateNetworkInterface(ctx, &resourcespb.CreateNetworkInterfaceRequest{
		Resource: convertFromNetworkInterface(plan),
	})
	if err != nil {
		return NetworkInterface{}, err
	}
	return convertToNetworkInterface(vn), nil
}

func updateNetworkInterface(ctx context.Context, p Provider, plan NetworkInterface) (NetworkInterface, error) {
	vn, err := p.Client.Client.UpdateNetworkInterface(ctx, &resourcespb.UpdateNetworkInterfaceRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromNetworkInterface(plan),
	})
	if err != nil {
		return NetworkInterface{}, err
	}
	return convertToNetworkInterface(vn), nil
}

func readNetworkInterface(ctx context.Context, p Provider, state NetworkInterface) (NetworkInterface, error) {
	vn, err := p.Client.Client.ReadNetworkInterface(ctx, &resourcespb.ReadNetworkInterfaceRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return NetworkInterface{}, err
	}
	return convertToNetworkInterface(vn), nil
}

func deleteNetworkInterface(ctx context.Context, p Provider, state NetworkInterface) error {
	_, err := p.Client.Client.DeleteNetworkInterface(ctx, &resourcespb.DeleteNetworkInterfaceRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type NetworkInterface struct {
	Id               types.String                             `tfsdk:"id"`
	Name             types.String                             `tfsdk:"name"`
	SubnetId         types.String                             `tfsdk:"subnet_id"`
	PublicIpId       types.String                             `tfsdk:"public_ip_id"`
	Cloud            mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location         mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
	ResourceGroupId  types.String                             `tfsdk:"resource_group_id"`
	AvailabilityZone types.Int64                              `tfsdk:"availability_zone"`
	AwsOutputs       types.Object                             `tfsdk:"aws"`
	AzureOutputs     types.Object                             `tfsdk:"azure"`
}

func convertToNetworkInterface(res *resourcespb.NetworkInterfaceResource) NetworkInterface {
	return NetworkInterface{
		Id:               types.String{Value: res.CommonParameters.ResourceId},
		ResourceGroupId:  types.String{Value: res.CommonParameters.ResourceGroupId},
		Name:             types.String{Value: res.Name},
		SubnetId:         types.String{Value: res.SubnetId},
		PublicIpId:       common.DefaultToNull[types.String](res.PublicIpId),
		AvailabilityZone: types.Int64{Value: int64(res.AvailabilityZone)},
		Cloud:            mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:         mtypes.LocationType.NewVal(res.CommonParameters.Location),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"network_interface_id": common.DefaultToNull[types.String](res.GetAwsOutputs().GetNetworkInterfaceId()),
				"eip_association_id":   common.DefaultToNull[types.String](res.GetAwsOutputs().GetEipAssociationId()),
			},
			AttrTypes: networkInterfaceAwsOutputs,
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"network_interface_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetNetworkInterfaceId()),
			},
			AttrTypes: networkInterfaceAzureOutputs,
		}),
	}
}

func convertFromNetworkInterface(plan NetworkInterface) *resourcespb.NetworkInterfaceArgs {
	return &resourcespb.NetworkInterfaceArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			ResourceGroupId: plan.ResourceGroupId.Value,
			Location:        plan.Location.Value,
			CloudProvider:   plan.Cloud.Value,
		},
		Name:             plan.Name.Value,
		SubnetId:         plan.SubnetId.Value,
		PublicIpId:       plan.PublicIpId.Value,
		AvailabilityZone: int32(plan.AvailabilityZone.Value),
	}
}
