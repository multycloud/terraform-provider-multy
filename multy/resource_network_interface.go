package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

var networkInterfaceSchema = tfsdk.Schema{
	MarkdownDescription: "Provides Multy Network Interface resource",
	Attributes: map[string]tfsdk.Attribute{
		"id": {
			Type:          types.StringType,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		},
		"resource_group_id": {
			Type:          types.StringType,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
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
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"public_ip_id": {
			Type:          types.StringType,
			Description:   "ID of `public_ip` resource",
			Optional:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		},
		"availability_zone": {
			Type:          types.Int64Type,
			Description:   "Availability zone where this machine should be placed",
			Optional:      true,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
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
		"resource_status": common.ResourceStatusSchema,
	},
}

func (r ResourceNetworkInterfaceType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[NetworkInterface]{
		p:          *(p.(*Provider)),
		createFunc: createNetworkInterface,
		updateFunc: updateNetworkInterface,
		readFunc:   readNetworkInterface,
		deleteFunc: deleteNetworkInterface,
		name:       "multy_network_interface",
		schema:     networkInterfaceSchema,
	}
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
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromNetworkInterface(plan),
	})
	if err != nil {
		return NetworkInterface{}, err
	}
	return convertToNetworkInterface(vn), nil
}

func readNetworkInterface(ctx context.Context, p Provider, state NetworkInterface) (NetworkInterface, error) {
	vn, err := p.Client.Client.ReadNetworkInterface(ctx, &resourcespb.ReadNetworkInterfaceRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return NetworkInterface{}, err
	}
	return convertToNetworkInterface(vn), nil
}

func deleteNetworkInterface(ctx context.Context, p Provider, state NetworkInterface) error {
	_, err := p.Client.Client.DeleteNetworkInterface(ctx, &resourcespb.DeleteNetworkInterfaceRequest{
		ResourceId: state.Id.ValueString(),
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
	ResourceStatus   types.Map                                `tfsdk:"resource_status"`
}

func convertToNetworkInterface(res *resourcespb.NetworkInterfaceResource) NetworkInterface {
	return NetworkInterface{
		Id:               types.StringValue(res.CommonParameters.ResourceId),
		ResourceGroupId:  types.StringValue(res.CommonParameters.ResourceGroupId),
		Name:             types.StringValue(res.Name),
		SubnetId:         types.StringValue(res.SubnetId),
		PublicIpId:       common.DefaultToNull[types.String](res.PublicIpId),
		AvailabilityZone: types.Int64Value(int64(res.AvailabilityZone)),
		Cloud:            mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:         mtypes.LocationType.NewVal(res.CommonParameters.Location),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, networkInterfaceAwsOutputs, map[string]attr.Value{
			"network_interface_id": common.DefaultToNull[types.String](res.GetAwsOutputs().GetNetworkInterfaceId()),
			"eip_association_id":   common.DefaultToNull[types.String](res.GetAwsOutputs().GetEipAssociationId()),
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, networkInterfaceAzureOutputs, map[string]attr.Value{
			"network_interface_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetNetworkInterfaceId()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromNetworkInterface(plan NetworkInterface) *resourcespb.NetworkInterfaceArgs {
	return &resourcespb.NetworkInterfaceArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			ResourceGroupId: plan.ResourceGroupId.ValueString(),
			Location:        plan.Location.Value,
			CloudProvider:   plan.Cloud.Value,
		},
		Name:             plan.Name.ValueString(),
		SubnetId:         plan.SubnetId.ValueString(),
		PublicIpId:       plan.PublicIpId.ValueString(),
		AvailabilityZone: int32(plan.AvailabilityZone.ValueInt64()),
	}
}
