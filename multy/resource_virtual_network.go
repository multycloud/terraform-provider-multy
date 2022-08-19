package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/commonpb"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceVirtualNetworkType struct{}

var virtualNetworkAwsOutputs = map[string]attr.Type{
	"vpc_id":                    types.StringType,
	"internet_gateway_id":       types.StringType,
	"default_security_group_id": types.StringType,
}

var virtualNetworkAzureOutputs = map[string]attr.Type{
	"virtual_network_id":   types.StringType,
	"local_route_table_id": types.StringType,
}

var virtualNetworkGcpOutputs = map[string]attr.Type{
	"compute_network_id":          types.StringType,
	"default_compute_firewall_id": types.StringType,
}

func (r ResourceVirtualNetworkType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {

	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Virtual Network resource",
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
				Description:   "Name of Virtual Network",
				Required:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("azure")},
			},
			"cidr_block": {
				Type:          types.StringType,
				Description:   "CIDR Block of Virtual Network",
				Required:      true,
				Validators:    []tfsdk.AttributeValidator{validators.IsCidrValidator{}},
				PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("aws")},
			},
			"gcp_overrides": {
				Description: "GCP-specific attributes that will be set if this resource is deployed in GCP",
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"project": {
						Type:          types.StringType,
						Description:   fmt.Sprintf("The project to use for this resource."),
						Optional:      true,
						Computed:      true,
						PlanModifiers: []tfsdk.AttributePlanModifier{common.RequiresReplaceIfCloudEq("gcp"), resource.UseStateForUnknown()},
						Validators:    []tfsdk.AttributeValidator{mtypes.NonEmptyStringValidator},
					},
				}),
				Optional: true,
				Computed: true,
			},
			"aws": {
				Description: "AWS-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: virtualNetworkAwsOutputs},
				Computed:    true,
			},
			"azure": {
				Description: "Azure-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: virtualNetworkAzureOutputs},
				Computed:    true,
			},
			"gcp": {
				Description: "GCP-specific ids of the underlying generated resources",
				Type:        types.ObjectType{AttrTypes: virtualNetworkGcpOutputs},
				Computed:    true,
			},
			"resource_status": common.ResourceStatusSchema,
			"cloud":           common.CloudsSchema,
			"location":        common.LocationSchema,
		},
	}, nil
}

func (r ResourceVirtualNetworkType) NewResource(_ context.Context, p provider.Provider) (resource.Resource, diag.Diagnostics) {
	return MultyResource[VirtualNetwork]{
		p:          *(p.(*Provider)),
		createFunc: createVirtualNetwork,
		updateFunc: updateVirtualNetwork,
		readFunc:   readVirtualNetwork,
		deleteFunc: deleteVirtualNetwork,
	}, nil
}

func createVirtualNetwork(ctx context.Context, p Provider, plan VirtualNetwork) (VirtualNetwork, error) {
	vn, err := p.Client.Client.CreateVirtualNetwork(ctx, &resourcespb.CreateVirtualNetworkRequest{
		Resource: convertFromVirtualNetwork(plan),
	})
	if err != nil {
		return VirtualNetwork{}, err
	}
	return convertToVirtualNetwork(vn), nil
}

func updateVirtualNetwork(ctx context.Context, p Provider, plan VirtualNetwork) (VirtualNetwork, error) {
	vn, err := p.Client.Client.UpdateVirtualNetwork(ctx, &resourcespb.UpdateVirtualNetworkRequest{
		ResourceId: plan.Id.Value,
		Resource:   convertFromVirtualNetwork(plan),
	})
	if err != nil {
		return VirtualNetwork{}, err
	}
	return convertToVirtualNetwork(vn), nil
}

func readVirtualNetwork(ctx context.Context, p Provider, state VirtualNetwork) (VirtualNetwork, error) {
	vn, err := p.Client.Client.ReadVirtualNetwork(ctx, &resourcespb.ReadVirtualNetworkRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return VirtualNetwork{}, err
	}
	return convertToVirtualNetwork(vn), nil
}

func deleteVirtualNetwork(ctx context.Context, p Provider, state VirtualNetwork) error {
	_, err := p.Client.Client.DeleteVirtualNetwork(ctx, &resourcespb.DeleteVirtualNetworkRequest{
		ResourceId: state.Id.Value,
	})
	return err
}

type VirtualNetwork struct {
	Id                 types.String `tfsdk:"id"`
	ResourceGroupId    types.String `tfsdk:"resource_group_id"`
	Name               types.String `tfsdk:"name"`
	CidrBlock          types.String `tfsdk:"cidr_block"`
	GcpOverridesObject types.Object `tfsdk:"gcp_overrides"`

	Cloud        mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location     mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
	AwsOutputs   types.Object                             `tfsdk:"aws"`
	AzureOutputs types.Object                             `tfsdk:"azure"`
	GcpOutputs   types.Object                             `tfsdk:"gcp"`

	ResourceStatus types.Map `tfsdk:"resource_status"`
}

func (v VirtualNetwork) UpdatePlan(_ context.Context, config VirtualNetwork, p Provider) (VirtualNetwork, []path.Path) {
	if config.Cloud.Value != commonpb.CloudProvider_GCP || p.Client.Gcp == nil {
		return v, nil
	}
	var requiresReplace []path.Path
	gcpOverrides := v.GetGcpOverrides()
	if o := config.GetGcpOverrides(); o == nil || o.Project.Unknown {
		if gcpOverrides == nil {
			gcpOverrides = &VirtualNetworkGcpOverrides{}
		}

		gcpOverrides.Project = types.String{
			Unknown: false,
			Null:    false,
			Value:   p.Client.Gcp.Project,
		}

		v.GcpOverridesObject = gcpOverrides.GcpOverridesToObj()
		requiresReplace = append(requiresReplace, path.Root("gcp_overrides").AtName("project"))
	}
	return v, requiresReplace
}

func convertToVirtualNetwork(res *resourcespb.VirtualNetworkResource) VirtualNetwork {
	return VirtualNetwork{
		Id:                 types.String{Value: res.CommonParameters.ResourceId},
		ResourceGroupId:    types.String{Value: res.CommonParameters.ResourceGroupId},
		Name:               types.String{Value: res.Name},
		CidrBlock:          types.String{Value: res.CidrBlock},
		GcpOverridesObject: convertToVirtualNetworkGcpOverrides(res.GcpOverride).GcpOverridesToObj(),
		Cloud:              mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:           mtypes.LocationType.NewVal(res.CommonParameters.Location),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"vpc_id":                    common.DefaultToNull[types.String](res.GetAwsOutputs().GetVpcId()),
				"internet_gateway_id":       common.DefaultToNull[types.String](res.GetAwsOutputs().GetInternetGatewayId()),
				"default_security_group_id": common.DefaultToNull[types.String](res.GetAwsOutputs().GetDefaultSecurityGroupId()),
			},
			AttrTypes: virtualNetworkAwsOutputs,
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"virtual_network_id":   common.DefaultToNull[types.String](res.GetAzureOutputs().GetVirtualNetworkId()),
				"local_route_table_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetLocalRouteTableId()),
			},
			AttrTypes: virtualNetworkAzureOutputs,
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, types.Object{
			Attrs: map[string]attr.Value{
				"compute_network_id":          common.DefaultToNull[types.String](res.GetGcpOutputs().GetComputeNetworkId()),
				"default_compute_firewall_id": common.DefaultToNull[types.String](res.GetGcpOutputs().GetDefaultComputeFirewallId()),
			},
			AttrTypes: virtualNetworkGcpOutputs,
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromVirtualNetwork(plan VirtualNetwork) *resourcespb.VirtualNetworkArgs {
	return &resourcespb.VirtualNetworkArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			ResourceGroupId: plan.ResourceGroupId.Value,
			Location:        plan.Location.Value,
			CloudProvider:   plan.Cloud.Value,
		},
		Name:        plan.Name.Value,
		CidrBlock:   plan.CidrBlock.Value,
		GcpOverride: convertFromVirtualNetworkGcpOverrides(plan.GetGcpOverrides()),
	}
}

func convertFromVirtualNetworkGcpOverrides(ref *VirtualNetworkGcpOverrides) *resourcespb.VirtualNetworkGcpOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.VirtualNetworkGcpOverride{Project: ref.Project.Value}
}

func convertToVirtualNetworkGcpOverrides(ref *resourcespb.VirtualNetworkGcpOverride) *VirtualNetworkGcpOverrides {
	if ref == nil {
		return nil
	}

	return &VirtualNetworkGcpOverrides{Project: common.DefaultToNull[types.String](ref.Project)}
}

func (v VirtualNetwork) GetGcpOverrides() (o *VirtualNetworkGcpOverrides) {
	if v.GcpOverridesObject.Null || v.GcpOverridesObject.Unknown {
		return
	}
	o = &VirtualNetworkGcpOverrides{
		Project: v.GcpOverridesObject.Attrs["project"].(types.String),
	}
	return
}

func (o *VirtualNetworkGcpOverrides) GcpOverridesToObj() types.Object {
	result := types.Object{
		Unknown: false,
		Null:    false,
		AttrTypes: map[string]attr.Type{
			"project": types.StringType,
		},
		Attrs: map[string]attr.Value{
			"project": types.String{Null: true},
		},
	}
	if o != nil {
		result.Attrs = map[string]attr.Value{
			"project": o.Project,
		}
	}

	return result
}

type VirtualNetworkGcpOverrides struct {
	Project types.String
}
