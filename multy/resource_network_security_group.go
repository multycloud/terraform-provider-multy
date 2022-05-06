package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	ruleProtocols  = []string{"tcp"}
)

type ResourceNetworkSecurityGroupType struct{}

func (r ResourceNetworkSecurityGroupType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		MarkdownDescription: "Provides Multy Network Security Group resource",
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:          types.StringType,
				Computed:      true,
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.UseStateForUnknown()},
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
				PlanModifiers: []tfsdk.AttributePlanModifier{tfsdk.RequiresReplace()},
			},
			"cloud":    common.CloudsSchema,
			"location": common.LocationSchema,
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
	}, nil
}

func (r ResourceNetworkSecurityGroupType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return MultyResource[NetworkSecurityGroup]{
		p:          *(p.(*Provider)),
		createFunc: createNetworkSecurityGroup,
		updateFunc: updateNetworkSecurityGroup,
		readFunc:   readNetworkSecurityGroup,
		deleteFunc: deleteNetworkSecurityGroup,
	}, nil
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
		ResourceId: plan.Id.Value,
		Resource:   convertFromNetworkSecurityGroup(plan),
	})
	if err != nil {
		return NetworkSecurityGroup{}, err
	}
	return convertToNetworkSecurityGroup(vn), nil
}

func readNetworkSecurityGroup(ctx context.Context, p Provider, state NetworkSecurityGroup) (NetworkSecurityGroup, error) {
	vn, err := p.Client.Client.ReadNetworkSecurityGroup(ctx, &resourcespb.ReadNetworkSecurityGroupRequest{
		ResourceId: state.Id.Value,
	})
	if err != nil {
		return NetworkSecurityGroup{}, err
	}
	return convertToNetworkSecurityGroup(vn), nil
}

func deleteNetworkSecurityGroup(ctx context.Context, p Provider, state NetworkSecurityGroup) error {
	_, err := p.Client.Client.DeleteNetworkSecurityGroup(ctx, &resourcespb.DeleteNetworkSecurityGroupRequest{
		ResourceId: state.Id.Value,
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
			Protocol:  types.String{Value: rule.Protocol},
			Priority:  types.Int64{Value: rule.Priority},
			FromPort:  types.Int64{Value: int64(rule.PortRange.From)},
			ToPort:    types.Int64{Value: int64(rule.PortRange.To)},
			CidrBlock: types.String{Value: rule.CidrBlock},
			Direction: types.String{Value: common.RuleDirectionToString(rule.Direction)},
		})
	}
	return NetworkSecurityGroup{
		Id:               types.String{Value: res.CommonParameters.ResourceId},
		Name:             types.String{Value: res.Name},
		VirtualNetworkId: types.String{Value: res.VirtualNetworkId},
		Rules:            rules,
		Cloud:            mtypes.CloudType.NewVal(res.CommonParameters.CloudProvider),
		Location:         mtypes.LocationType.NewVal(res.CommonParameters.Location),
	}
}

func convertFromNetworkSecurityGroup(plan NetworkSecurityGroup) *resourcespb.NetworkSecurityGroupArgs {
	var rules []*resourcespb.NetworkSecurityRule
	for _, item := range plan.Rules {
		ruleDirection := common.StringToRuleDirection(item.Direction.Value)
		rules = append(rules, &resourcespb.NetworkSecurityRule{
			Protocol: item.Protocol.Value,
			Priority: item.Priority.Value,
			PortRange: &resourcespb.PortRange{
				From: int32(item.FromPort.Value),
				To:   int32(item.ToPort.Value),
			},
			CidrBlock: item.CidrBlock.Value,
			Direction: ruleDirection,
		})
	}
	return &resourcespb.NetworkSecurityGroupArgs{
		CommonParameters: &commonpb.ResourceCommonArgs{
			Location:      plan.Location.Value,
			CloudProvider: plan.Cloud.Value,
		},
		Name:             plan.Name.Value,
		VirtualNetworkId: plan.VirtualNetworkId.Value,
		Rules:            rules,
	}
}
