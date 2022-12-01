package multy

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/multycloud/multy/api/proto/resourcespb"
	"terraform-provider-multy/multy/common"
	"terraform-provider-multy/multy/mtypes"
	"terraform-provider-multy/multy/validators"
)

type ResourceVaultAccessPolicyType struct{}

var vaultAccessPolicyAwsOutputs = map[string]attr.Type{
	"iam_policy_arn": types.StringType,
}

var vaultAccessPolicyAzureOutputs = map[string]attr.Type{
	"key_vault_access_policy_id": types.StringType,
}

var vaultAccessPolicyGcpOutputs = map[string]attr.Type{
	"secret_manager_secret_iam_membership_ids": types.ListType{ElemType: types.StringType},
}

var vapSchema = tfsdk.Schema{
	MarkdownDescription: "Provides Multy Object Storage resource",
	Attributes: map[string]tfsdk.Attribute{
		"id": {
			Type:          types.StringType,
			Computed:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.UseStateForUnknown()},
		},
		"vault_id": {
			Type:          types.StringType,
			Description:   "Id of the associated vault",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"identity": {
			Type:          types.StringType,
			Description:   "Identity of the resource that is being granted access to the `vault`",
			Required:      true,
			PlanModifiers: []tfsdk.AttributePlanModifier{resource.RequiresReplace()},
		},
		"access": {
			Type:        mtypes.VaultAclType,
			Description: fmt.Sprintf("Access control, available values are %v", mtypes.VaultAclType.GetAllValues()),
			Required:    true,
			Validators:  []tfsdk.AttributeValidator{validators.NewValidator(mtypes.VaultAclType)},
		},
		"aws": {
			Description: "AWS-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: vaultAccessPolicyAwsOutputs},
			Computed:    true,
		},
		"azure": {
			Description: "Azure-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: vaultAccessPolicyAzureOutputs},
			Computed:    true,
		},
		"gcp": {
			Description: "GCP-specific ids of the underlying generated resources",
			Type:        types.ObjectType{AttrTypes: vaultAccessPolicyGcpOutputs},
			Computed:    true,
		},
		"resource_status": common.ResourceStatusSchema,
	},
}

func (r ResourceVaultAccessPolicyType) NewResource(_ context.Context, p provider.Provider) resource.Resource {
	return MultyResource[VaultAccessPolicy]{
		p:          *(p.(*Provider)),
		createFunc: createVaultAccessPolicy,
		updateFunc: updateVaultAccessPolicy,
		readFunc:   readVaultAccessPolicy,
		deleteFunc: deleteVaultAccessPolicy,
		name:       "multy_vault_access_policy",
		schema:     vapSchema,
	}
}

func createVaultAccessPolicy(ctx context.Context, p Provider, plan VaultAccessPolicy) (VaultAccessPolicy, error) {
	vn, err := p.Client.Client.CreateVaultAccessPolicy(ctx, &resourcespb.CreateVaultAccessPolicyRequest{
		Resource: convertFromVaultAccessPolicy(plan),
	})
	if err != nil {
		return VaultAccessPolicy{}, err
	}
	return convertToVaultAccessPolicy(vn), nil
}

func updateVaultAccessPolicy(ctx context.Context, p Provider, plan VaultAccessPolicy) (VaultAccessPolicy, error) {
	vn, err := p.Client.Client.UpdateVaultAccessPolicy(ctx, &resourcespb.UpdateVaultAccessPolicyRequest{
		ResourceId: plan.Id.ValueString(),
		Resource:   convertFromVaultAccessPolicy(plan),
	})
	if err != nil {
		return VaultAccessPolicy{}, err
	}
	return convertToVaultAccessPolicy(vn), nil
}

func readVaultAccessPolicy(ctx context.Context, p Provider, state VaultAccessPolicy) (VaultAccessPolicy, error) {
	vn, err := p.Client.Client.ReadVaultAccessPolicy(ctx, &resourcespb.ReadVaultAccessPolicyRequest{
		ResourceId: state.Id.ValueString(),
	})
	if err != nil {
		return VaultAccessPolicy{}, err
	}
	return convertToVaultAccessPolicy(vn), nil
}

func deleteVaultAccessPolicy(ctx context.Context, p Provider, state VaultAccessPolicy) error {
	_, err := p.Client.Client.DeleteVaultAccessPolicy(ctx, &resourcespb.DeleteVaultAccessPolicyRequest{
		ResourceId: state.Id.ValueString(),
	})
	return err
}

type VaultAccessPolicy struct {
	Id             types.String                                   `tfsdk:"id"`
	VaultId        types.String                                   `tfsdk:"vault_id"`
	Identity       types.String                                   `tfsdk:"identity"`
	Access         mtypes.EnumValue[resourcespb.VaultAccess_Enum] `tfsdk:"access"`
	AwsOutputs     types.Object                                   `tfsdk:"aws"`
	AzureOutputs   types.Object                                   `tfsdk:"azure"`
	GcpOutputs     types.Object                                   `tfsdk:"gcp"`
	ResourceStatus types.Map                                      `tfsdk:"resource_status"`
}

func convertToVaultAccessPolicy(res *resourcespb.VaultAccessPolicyResource) VaultAccessPolicy {
	return VaultAccessPolicy{
		Id:       types.StringValue(res.CommonParameters.ResourceId),
		VaultId:  types.StringValue(res.VaultId),
		Identity: types.StringValue(res.Identity),
		Access:   mtypes.VaultAclType.NewVal(res.Access),
		AwsOutputs: common.OptionallyObj(res.AwsOutputs, vaultAccessPolicyAwsOutputs, map[string]attr.Value{
			"iam_policy_arn": common.DefaultToNull[types.String](res.GetAwsOutputs().GetIamPolicyArn()),
		}),
		AzureOutputs: common.OptionallyObj(res.AzureOutputs, vaultAccessPolicyAzureOutputs, map[string]attr.Value{
			"key_vault_access_policy_id": common.DefaultToNull[types.String](res.GetAzureOutputs().GetKeyVaultAccessPolicyId()),
		}),
		GcpOutputs: common.OptionallyObj(res.GcpOutputs, vaultAccessPolicyGcpOutputs, map[string]attr.Value{
			"secret_manager_secret_iam_membership_ids": common.TypesStringListToListType(res.GetGcpOutputs().GetSecretManagerSecretIamMembershipId()),
		}),
		ResourceStatus: common.GetResourceStatus(res.CommonParameters.GetResourceStatus()),
	}
}

func convertFromVaultAccessPolicy(plan VaultAccessPolicy) *resourcespb.VaultAccessPolicyArgs {
	return &resourcespb.VaultAccessPolicyArgs{
		VaultId:  plan.VaultId.ValueString(),
		Identity: plan.Identity.ValueString(),
		Access:   plan.Access.Value,
	}
}
