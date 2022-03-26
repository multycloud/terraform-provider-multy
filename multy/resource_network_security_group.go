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
	"strconv"
	"strings"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/validators"
)

var (
	ruleDirections = []string{"ingress", "egress", "both"}
	ruleProtocols  = []string{"tcp"}
)

type ResourceNetworkSecurityGroupType struct{}

func (r ResourceNetworkSecurityGroupType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:     types.StringType,
				Computed: true,
			},
			"name": {
				Type:        types.StringType,
				Description: "Name of Network Security Group",
				Required:    true,
			},
			"virtual_network_id": {
				Type:        types.StringType,
				Description: "ID of `virtual_network` resource",
				Required:    true,
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
						Validators:  []tfsdk.AttributeValidator{validators.StringInSliceValidator{Enum: ruleProtocols}},
					},
					"priority": {
						Type:        types.Int64Type,
						Description: fmt.Sprintf("Priority of network rule. Value must be in between %d and %d", 0, 0),
						Required:    true,
					},
					"from_port": {
						Type:        types.Int64Type,
						Description: fmt.Sprintf("From port of network rule port range. Value must be in between %d and %d", 0, 0),
						Required:    true,
						//Validators: validateRulePort,
					},
					"to_port": {
						Type:        types.Int64Type,
						Description: fmt.Sprintf("To port of network rule port range. Value must be in between %d and %d", 0, 0),
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
						Validators:  []tfsdk.AttributeValidator{validators.StringInSliceValidator{Enum: ruleDirections}},
					},
				},
				NestingMode: tfsdk.BlockNestingModeList,
			},
		},
	}, nil
}

func (r ResourceNetworkSecurityGroupType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceNetworkSecurityGroup{
		p: *(p.(*Provider)),
	}, nil
}

type resourceNetworkSecurityGroup struct {
	p Provider
}

func (r resourceNetworkSecurityGroup) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.Configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	// Retrieve values from plan
	var plan NetworkSecurityGroup
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	nsg, err := c.Client.CreateNetworkSecurityGroup(ctx, &resources.CreateNetworkSecurityGroupRequest{
		Resources: r.convertResourcePlanToArgs(plan),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating network_security_group", err.Error())
		return
	}

	tflog.Trace(ctx, "created nsg", map[string]interface{}{"network_security_group_id": nsg.CommonParameters.ResourceId})

	// Map response body to resource schema attribute
	state := r.convertResponseToResource(nsg)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceNetworkSecurityGroup) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state NetworkSecurityGroup
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Get network_security_group from API and then update what is in state from what the API returns
	nsg, err := r.p.Client.Client.ReadNetworkSecurityGroup(ctx, &resources.ReadNetworkSecurityGroupRequest{ResourceId: state.Id.Value})
	if err != nil {
		resp.Diagnostics.AddError("Error getting network_security_group", err.Error())
		return
	}

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(nsg)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceNetworkSecurityGroup) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan, state NetworkSecurityGroup
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

	request := &resources.UpdateNetworkSecurityGroupRequest{
		ResourceId: state.Id.Value,
		Resources:  r.convertResourcePlanToArgs(plan),
	}

	// Update network_security_group
	nsg, err := c.Client.UpdateNetworkSecurityGroup(ctx, request)
	if err != nil {
		resp.Diagnostics.AddError("Error creating network_security_group", err.Error())
		return
	}

	tflog.Trace(ctx, "updated network_security_group", map[string]interface{}{"network_security_group_id": state.Id.Value})

	// Map response body to resource schema attribute & Set state
	state = r.convertResponseToResource(nsg)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceNetworkSecurityGroup) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state NetworkSecurityGroup
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	c := r.p.Client
	ctx = c.AddHeaders(ctx)

	// Delete network_security_group
	_, err := c.Client.DeleteNetworkSecurityGroup(ctx, &resources.DeleteNetworkSecurityGroupRequest{ResourceId: state.Id.Value})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting network_security_group",
			err.Error(),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

func (r resourceNetworkSecurityGroup) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	// Save the import identifier in the id attribute
	tfsdk.ResourceImportStatePassthroughID(ctx, tftypes.NewAttributePath().WithAttributeName("id"), req, resp)
}

func validateRulePort(i interface{}, k string) (warnings []string, errors []error) {
	v, ok := i.(string)
	if !ok {
		errors = append(errors, fmt.Errorf("expected type of %s to be string", k))
		return warnings, errors
	}

	if i, err := strconv.Atoi(v); err != nil || i < -1 {
		errors = append(errors, fmt.Errorf("expected %s to be between greater than -1, got %s", k, v))
	}
	return warnings, errors
}

type NetworkSecurityGroup struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	VirtualNetworkId types.String `tfsdk:"virtual_network_id"`
	Rules            []Rule       `tfsdk:"rule"`
	Cloud            types.String `tfsdk:"cloud"`
	Location         types.String `tfsdk:"location"`
}

type Rule struct {
	Protocol  types.String `tfsdk:"protocol"`
	Priority  types.Int64  `tfsdk:"priority"`
	FromPort  types.Int64  `tfsdk:"from_port"`
	ToPort    types.Int64  `tfsdk:"to_port"`
	CidrBlock types.String `tfsdk:"cidr_block"`
	Direction types.String `tfsdk:"direction"`
}

func (r resourceNetworkSecurityGroup) convertResponseToResource(res *resources.NetworkSecurityGroupResource) NetworkSecurityGroup {
	var rules []Rule
	for _, rule := range res.Resources[0].Rules {
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
		Name:             types.String{Value: res.Resources[0].Name},
		VirtualNetworkId: types.String{Value: res.Resources[0].VirtualNetworkId},
		Rules:            rules,
		Cloud:            types.String{Value: strings.ToLower(res.Resources[0].CommonParameters.CloudProvider.String())},
		Location:         types.String{Value: strings.ToLower(res.Resources[0].CommonParameters.Location.String())},
	}
}

func (r resourceNetworkSecurityGroup) convertResourcePlanToArgs(plan NetworkSecurityGroup) []*resources.CloudSpecificNetworkSecurityGroupArgs {
	var rules []*resources.NetworkSecurityRule
	for _, item := range plan.Rules {
		ruleDirection := common.StringToRuleDirection(item.Direction.Value)
		rules = append(rules, &resources.NetworkSecurityRule{
			Protocol: item.Protocol.Value,
			Priority: item.Priority.Value,
			PortRange: &resources.PortRange{
				From: int32(item.FromPort.Value),
				To:   int32(item.FromPort.Value),
			},
			CidrBlock: item.CidrBlock.Value,
			Direction: ruleDirection,
		})
	}
	return []*resources.CloudSpecificNetworkSecurityGroupArgs{{
		CommonParameters: &common_proto.CloudSpecificResourceCommonArgs{
			Location:      common.StringToLocation(plan.Location.Value),
			CloudProvider: common.StringToCloud(plan.Cloud.Value),
		},
		Name:             plan.Name.Value,
		VirtualNetworkId: plan.VirtualNetworkId.Value,
		Rules:            rules,
	}}
}
