package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

var (
	ruleDirections = []string{"ingress", "egress", "both"}
	ruleProtocols  = []string{"tcp", "udp", "icmp"}
)

type ResourceNetworkSecurityGroupType struct{}

var networkSecurityGroupAwsOutputs = map[string]attr.Type{
	"security_group_id": types.StringType,
}

var networkSecurityGroupAzureOutputs = map[string]attr.Type{
	"network_security_group_id": types.StringType,
}

var networkSecurityGroupGcpOutputs = map[string]attr.Type{
	"compute_firewall_ids": types.ListType{ElemType: types.StringType},
}

var nsgSchema = tfsdk.Schema{
	MarkdownDescription: "Provides Multy Network Security Group resource",
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
			Type:        types.StringType,
			Description: "Name of Network Security Group",
			Required:    true,
		},
		"virtual_network_id": {
			Type:          types.StringType,
			Description:   "ID of `virtual_network` resource",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"cloud":    common.CloudsSchema,
		"location": common.LocationSchema,
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
			Type:        types.ObjectType{AttrTypes: networkSecurityGroupAwsOutputs},
			Computed:    true,
		},
		"azure": {
			Description: "Azure-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: networkSecurityGroupAzureOutputs},
			Computed:    true,
		},
		"gcp": {
			Description: "GCP-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: networkSecurityGroupGcpOutputs},
			Computed:    true,
		},
		"resource_status": common.ResourceStatusSchema,
	},
	Blocks: map[string]tfsdk.Block{
		"rule": {
			//Optional: true,
			Description: "Network rule block definition",
			Attributes: map[string]tfsdk.Attribute{
				"protocol": {
					Type:        types.StringType,
					Description: fmt.Sprintf("Protocol of network rule. Accepted values are %s", common.StringSliceToDocsMarkdown(ruleProtocols)),
					Required:    true,
					Validators:  []tfsdk.AttributeValidator{validators.StringInSliceValidator{Values: ruleProtocols}},
				},
				"priority": {
					Type:        types.Int64Type,
					Description: fmt.Sprintf("Priority of network rule. Value must be in between %d and %d", 0, 0),
					Required:    true,
				},
				"from_port": {
					Type:        types.Int64Type,
					Description: fmt.Sprintf("From port of network rule port range. Value must be in between %d and %d", 0, 65535),
					Required:    true,
					//Validators: validateRulePort,
				},
				"to_port": {
					Type:        types.Int64Type,
					Description: fmt.Sprintf("To port of network rule port range. Value must be in between %d and %d", 0, 65535),
					Required:    true,
					//Validators: validateRulePort,
				},
				"cidr_block": {
					Type:        types.StringType,
					Description: "CIDR block of network rule",
					Required:    true,
					Validators:  []tfsdk.AttributeValidator{validators.IsCidrValidator{}},
				},
				"direction": {
					Type:        types.StringType,
					Description: fmt.Sprintf("Direction of network rule. Accepted values are %s", common.StringSliceToDocsMarkdown(ruleDirections)),
					Required:    true,
					Validators:  []tfsdk.AttributeValidator{validators.StringInSliceValidator{Values: ruleDirections}},
				},
			},
			NestingMode: tfsdk.BlockNestingModeList,
		},
	},
}

func (r ResourceNetworkSecurityGroupType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[NetworkSecurityGroup]{
		p:          *(p.(*Provider)),
		createFunc: createNetworkSecurityGroup,
		updateFunc: updateNetworkSecurityGroup,
		readFunc:   readNetworkSecurityGroup,
		deleteFunc: deleteNetworkSecurityGroup,
		name:       "multy_network_security_group",
		schema:     nsgSchema,
	}
}

func createNetworkSecurityGroup(ctx context.Context, p Provider, plan NetworkSecurityGroup) (NetworkSecurityGroup, error) {
	vn, err := p.Client.Client.CreateNetworkSecurityGroup(ctx, &resourcespb.CreateNetworkSecurityGroupRequest{
		Resource: convertFromNetworkSecurityGroup(plan),
	})
	if err != nil {
		return NetworkSecurityGroup{}, err
	}
	return convertToNetworkSecurityGroup(vn), nil
}

func updateNetworkSecurityGroup(ctx context.Context, p Provider, plan NetworkSecurityGroup) (NetworkSecurityGroup, error) {
	vn, err := p.Client.Client.UpdateNetworkSecurityGroup(ctx, &resourcespb.UpdateNetworkSecurityGroupRequest{
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromNetworkSecurityGroup(plan),
	})
	if err != nil {
		return NetworkSecurityGroup{}, err
	}
	return convertToNetworkSecurityGroup(vn), nil
}

func readNetworkSecurityGroup(ctx context.Context, p Provider, state NetworkSecurityGroup) (NetworkSecurityGroup, error) {
	vn, err := p.Client.Client.ReadNetworkSecurityGroup(ctx, &resourcespb.ReadNetworkSecurityGroupRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return NetworkSecurityGroup{}, err
	}
	return convertToNetworkSecurityGroup(vn), nil
}

func deleteNetworkSecurityGroup(ctx context.Context, p Provider, state NetworkSecurityGroup) error {
	_, err := p.Client.Client.DeleteNetworkSecurityGroup(ctx, &resourcespb.DeleteNetworkSecurityGroupRequest{
		ResourceId: state.Id.ValueString(),
	})
	return err
}

type NetworkSecurityGroup struct {
	Id               types.String                             `tfsdk:"id"`
	Name             types.String                             `tfsdk:"name"`
	VirtualNetworkId types.String                             `tfsdk:"virtual_network_id"`
	Rules            []Rule                                   `tfsdk:"rule"`
	Cloud            mtypes.EnumValue[commonpb.CloudProvider] `tfsdk:"cloud"`
	Location         mtypes.EnumValue[commonpb.Location]      `tfsdk:"location"`
	ResourceGroupId  types.String                             `tfsdk:"resource_group_id"`

	GcpOverridesObject types.Object `tfsdk:"gcp_overrides"`
	AwsOutputs         types.Object `tfsdk:"aws"`
	AzureOutputs       types.Object `tfsdk:"azure"`
	GcpOutputs         types.Object `tfsdk:"gcp"`
	ResourceStatus     types.Map    `tfsdk:"resource_status"`
}

type Rule struct {
	Protocol  types.String `tfsdk:"protocol"`
	Priority  types.Int64  `tfsdk:"priority"`
	FromPort  types.Int64  `tfsdk:"from_port"`
	ToPort    types.Int64  `tfsdk:"to_port"`
	CidrBlock types.String `tfsdk:"cidr_block"`
	Direction types.String `tfsdk:"direction"`
}

func convertToNetworkSecurityGroup(res *resourcespb.NetworkSecurityGroupResource) NetworkSecurityGroup {
	var rules []Rule
	for _, rule := range res.Rules {
		rules = append(rules, Rule{
			Protocol:  types.StringValue(rule.Protocol),
			Priority:  types.Int64Value(rule.Priority),
			FromPort:  types.Int64Value(int64(rule.PortRange.From)),
			ToPort:    types.Int64Value(int64(rule.PortRange.To)),
			CidrBlock: types.StringValue(rule.CidrBlock),
			Direction: types.StringValue(common.RuleDirectionToString(rule.Direction)),
		})
	}
	return NetworkSecurityGroup{
		Id:                 types.StringValue(res.CommonParameters.ResourceId),
		Name:               types.StringValue(res.Name),
		VirtualNetworkId:   types.StringValue(res.VirtualNetworkId),
		Rules:              rules,
		Cloud:              mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:           mtypes.LocationType.NewVal(res.CommonParameters.Location),
		ResourceGroupId:    types.StringValue(res.CommonParameters.ResourceGroupId),
		GcpOverridesObject: convertToNetworkSecurityGroupGcpOverrides(res.GcpOverride).GcpOverridesToObj(),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, networkSecurityGroupAwsOutputs, map[string]attr.Value{
			"security_group_id": common.DefaultToNull[types.String](res.GetAwsOutputs().GetSecurityGroupId()),
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, networkSecurityGroupAzureOutputs, map[string]attr.Value{
			"network_security_group_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetNetworkSecurityGroupId()),
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, networkSecurityGroupGcpOutputs, map[string]attr.Value{
			"compute_firewall_ids": common.TypesStringListToListType(res.GetGcpOutputs().GetComputeFirewallId()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromNetworkSecurityGroup(plan NetworkSecurityGroup) *resourcespb.NetworkSecurityGroupArgs {
	var rules []*resourcespb.NetworkSecurityRule
	for _, item := range plan.Rules {
		ruleDirection := common.StringToRuleDirection(item.Direction.ValueString())
		rules = append(rules, &resourcespb.NetworkSecurityRule{
			Protocol: item.Protocol.ValueString(),
			Priority: item.Priority.ValueInt64(),
			PortRange: &resourcespb.PortRange{
				From: int32(item.FromPort.ValueInt64()),
				To:   int32(item.ToPort.ValueInt64()),
			},
			CidrBlock: item.CidrBlock.ValueString(),
			Direction: ruleDirection,
		})
	}
	return &resourcespb.NetworkSecurityGroupArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			Location:        plan.Location.Value,
			CloudProvider:   plan.Cloud.Value,
			ResourceGroupId: plan.ResourceGroupId.ValueString(),
		},
		Name:             plan.Name.ValueString(),
		VirtualNetworkId: plan.VirtualNetworkId.ValueString(),
		Rules:            rules,
		GcpOverride:      convertFromNetworkSecurityGroupGcpOverrides(plan.GetGcpOverrides()),
	}
}

func convertFromNetworkSecurityGroupGcpOverrides(ref *NetworkSecurityGroupGcpOverrides) *resourcespb.NetworkSecurityGroupGcpOverride {
	if ref == nil {
		return nil
	}

	return &resourcespb.NetworkSecurityGroupGcpOverride{Project: ref.Project.ValueString()}
}

func convertToNetworkSecurityGroupGcpOverrides(ref *resourcespb.NetworkSecurityGroupGcpOverride) *NetworkSecurityGroupGcpOverrides {
	if ref == nil {
		return nil
	}

	return &NetworkSecurityGroupGcpOverrides{Project: common.DefaultToNull[types.String](ref.Project)}
}

func (v NetworkSecurityGroup) GetGcpOverrides() (o *NetworkSecurityGroupGcpOverrides) {
	if v.GcpOverridesObject.IsNull() || v.GcpOverridesObject.IsUnknown() {
		return
	}
	o = &NetworkSecurityGroupGcpOverrides{
		Project: v.GcpOverridesObject.Attributes()["project"].(types.String),
	}
	return
}

func (o *NetworkSecurityGroupGcpOverrides) GcpOverridesToObj() types.Object {
	attrTypes := map[string]attr.Type{
		"project": types.StringType,
	}
	if o == nil {
		return types.ObjectNull(attrTypes)
	}
	result, _ := types.ObjectValue(attrTypes, map[string]attr.Value{"project": o.Project})
	return result
}

type NetworkSecurityGroupGcpOverrides struct {
	Project types.String
}

func (v NetworkSecurityGroup) UpdatePlan(_ context.Context, config NetworkSecurityGroup, p Provider) (NetworkSecurityGroup, []path.Path) {
	if config.Cloud.Value != commonpb.CloudProvider_GCP || p.Client.Gcp == nil {
		return v, nil
	}
	var requiresReplace []path.Path
	gcpOverrides := v.GetGcpOverrides()
	if o := config.GetGcpOverrides(); o == nil || o.Project.IsUnknown() {
		if gcpOverrides == nil {
			gcpOverrides = &NetworkSecurityGroupGcpOverrides{}
		}

		gcpOverrides.Project = types.StringValue(p.Client.Gcp.Project)

		v.GcpOverridesObject = gcpOverrides.GcpOverridesToObj()
		requiresReplace = append(requiresReplace, path.Root("gcp_overrides").AtName("project"))
	}
	return v, requiresReplace
}
