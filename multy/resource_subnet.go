package multy

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/validators"
)

type ResourceSubnetType struct{}

var subnetAwsOutputs = map[string]attr.Type{
	"subnet_id_by_availability_zone": types.MapType{ElemType: types.StringType},
}

var subnetAzureOutputs = map[string]attr.Type{
	"subnet_id": types.StringType,
}

var subnetGcpOutputs = map[string]attr.Type{
	"compute_subnetwork_id": types.StringType,
}

var subnetSchema = tfsdk.Schema{
	MarkdownDescription: "Provides Multy Subnet resource",
	Attributes: map[string]tfsdk.Attribute{
		"id": {
			Type:          types.StringType,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		},
		"name": {
			Type:        types.StringType,
			Description: "Name of Subnet",
			Required:    true,
			//PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"cidr_block": {
			Type:        types.StringType,
			Description: "CIDR block of Subnet",
			Required:    true,
			Validators:  []tfsdk.AttributeValidator{validators.IsCidrValidator{}},
			//PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("aws")},
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"virtual_network_id": {
			Type:          types.StringType,
			Description:   "ID of `virtual_network` resource",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"aws": {
			Description: "AWS-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: subnetAwsOutputs},
			Computed:    true,
		},
		"azure": {
			Description: "Azure-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: subnetAzureOutputs},
			Computed:    true,
		},
		"gcp": {
			Description: "GCP-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: subnetGcpOutputs},
			Computed:    true,
		},
		"resource_status": common.ResourceStatusSchema,
	},
}

func (r ResourceSubnetType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[Subnet]{
		p:          *(p.(*Provider)),
		createFunc: createSubnet,
		updateFunc: updateSubnet,
		readFunc:   readSubnet,
		deleteFunc: deleteSubnet,
		name:       "multy_subnet",
		schema:     subnetSchema,
	}
}

func createSubnet(ctx context.Context, p Provider, plan Subnet) (Subnet, error) {
	vn, err := p.Client.Client.CreateSubnet(ctx, &resourcespb.CreateSubnetRequest{
		Resource: convertFromSubnet(plan),
	})
	if err != nil {
		return Subnet{}, err
	}
	return convertToSubnet(vn), nil
}

func updateSubnet(ctx context.Context, p Provider, plan Subnet) (Subnet, error) {
	vn, err := p.Client.Client.UpdateSubnet(ctx, &resourcespb.UpdateSubnetRequest{
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromSubnet(plan),
	})
	if err != nil {
		return Subnet{}, err
	}
	return convertToSubnet(vn), nil
}

func readSubnet(ctx context.Context, p Provider, state Subnet) (Subnet, error) {
	vn, err := p.Client.Client.ReadSubnet(ctx, &resourcespb.ReadSubnetRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return Subnet{}, err
	}
	return convertToSubnet(vn), nil
}

func deleteSubnet(ctx context.Context, p Provider, state Subnet) error {
	_, err := p.Client.Client.DeleteSubnet(ctx, &resourcespb.DeleteSubnetRequest{
		ResourceId: state.Id.ValueString(),
	})
	return err
}

type Subnet struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	CidrBlock        types.String `tfsdk:"cidr_block"`
	VirtualNetworkId types.String `tfsdk:"virtual_network_id"`
	AwsOutputs       types.Object `tfsdk:"aws"`
	AzureOutputs     types.Object `tfsdk:"azure"`
	GcpOutputs       types.Object `tfsdk:"gcp"`
	ResourceStatus   types.Map    `tfsdk:"resource_status"`
}

func convertToSubnet(res *resourcespb.SubnetResource) Subnet {
	result := Subnet{
		Id:               types.StringValue(res.CommonParameters.ResourceId),
		Name:             types.StringValue(res.Name),
		CidrBlock:        types.StringValue(res.CidrBlock),
		VirtualNetworkId: types.StringValue(res.VirtualNetworkId),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, subnetAwsOutputs, map[string]attr.Value{
			"subnet_id_by_availability_zone": common.GoMapToMapType(res.GetAwsOutputs().GetSubnetIdByAvailabilityZone()),
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, subnetAzureOutputs, map[string]attr.Value{
			"subnet_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetSubnetId()),
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, subnetGcpOutputs, map[string]attr.Value{
			"compute_subnetwork_id": common.DefaultToNull[types.String](res.GetGcpOutputs().GetComputeSubnetworkId()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}

	return result
}

func convertFromSubnet(plan Subnet) *resourcespb.SubnetArgs {
	return &resourcespb.SubnetArgs{
		Name:             plan.Name.ValueString(),
		CidrBlock:        plan.CidrBlock.ValueString(),
		VirtualNetworkId: plan.VirtualNetworkId.ValueString(),
	}
}
